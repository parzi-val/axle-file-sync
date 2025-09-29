package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Global variables for watcher state
var (
	changes         []FileChange
	mu              sync.Mutex
	lastEventTime   = make(map[string]time.Time) // Generic debounce map
	eventTimeMutex  sync.RWMutex                  // Mutex for lastEventTime map
	muApplyingPatch sync.Mutex
	isApplyingPatch bool
	// Batching variables
	pendingFiles  = make(map[string]string) // file path -> event type
	batchTimer    *time.Timer
	batchMutex    sync.Mutex
	batchDuration = 5 * time.Second // Wait 5 seconds to accumulate changes (increased for testing)
	// File size limits
	maxFileSize int64 = 10 * 1024 * 1024 // 10MB default, can be configured
	// Dynamic batching
	recentEventCount int                    // Track recent events for dynamic batching
	lastEventReset   = time.Now()           // When we last reset the event counter
	dynamicBatchMux  sync.RWMutex           // Mutex for dynamic batch variables
)

// SetIsApplyingPatch sets the state of the patch application flag.
// This is used to temporarily mute the watcher.
func SetIsApplyingPatch(state bool) {
	muApplyingPatch.Lock()
	defer muApplyingPatch.Unlock()
	isApplyingPatch = state
}

func getIsApplyingPatch() bool {
	muApplyingPatch.Lock()
	defer muApplyingPatch.Unlock()
	return isApplyingPatch
}

// debounceEvent prevents duplicate events within the debounce window
func debounceEvent(eventMap map[string]time.Time, key string, debounceTime time.Duration) bool {
	eventTimeMutex.Lock()
	defer eventTimeMutex.Unlock()

	now := time.Now()
	if lastTime, exists := eventMap[key]; exists {
		if now.Sub(lastTime) < debounceTime {
			return false // Skip duplicate event
		}
	}
	eventMap[key] = now
	return true // Accept event
}

// cleanupOldEventTimes removes entries older than 5 minutes from lastEventTime map
func cleanupOldEventTimes() {
	eventTimeMutex.Lock()
	defer eventTimeMutex.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for key, timestamp := range lastEventTime {
		if timestamp.Before(cutoff) {
			delete(lastEventTime, key)
		}
	}
}

// isIgnored checks if a path should be ignored.
func isIgnored(path string, ignorePatterns []string) bool {
	fileName := filepath.Base(path)

	// Always ignore .git folder and its contents
	if fileName == ".git" || strings.Contains(path, ".git"+string(filepath.Separator)) {
		return true
	}

	// Ignore temporary/swap files
	if strings.HasSuffix(fileName, ".tmp") || strings.HasSuffix(fileName, ".swp") || strings.HasSuffix(fileName, "~") {
		return true
	}

	for _, pattern := range ignorePatterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	return false
}

// shouldSkipFile checks if a file should be skipped based on size or other criteria
func shouldSkipFile(path string) (bool, string) {
	// Check if path exists and get file info
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "" // File doesn't exist (might be deleted), don't skip
		}
		return true, fmt.Sprintf("cannot stat file: %v", err)
	}

	// Skip directories (they're handled separately)
	if fileInfo.IsDir() {
		return false, ""
	}

	// Check file size
	if fileInfo.Size() > maxFileSize {
		return true, fmt.Sprintf("file size %d bytes exceeds limit of %d bytes", fileInfo.Size(), maxFileSize)
	}

	// Check for binary files (optional, basic heuristic)
	if isBinaryFile(path) {
		return true, "binary file detected"
	}

	return false, ""
}

// isBinaryFile performs a basic check to see if a file appears to be binary
func isBinaryFile(path string) bool {
	// Common binary extensions to skip
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".a", ".o",
		".zip", ".tar", ".gz", ".7z", ".rar",
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".ico", ".svg",
		".mp3", ".mp4", ".avi", ".mov", ".wmv",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".db", ".sqlite", ".mdb",
		".bin", ".dat", ".iso",
		".pyc", ".pyo", ".class",
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, binExt := range binaryExts {
		if ext == binExt {
			return true
		}
	}

	return false
}

// processBatch processes accumulated file changes and commits them as a batch
func processBatch(cfg AppConfig) {
	batchMutex.Lock()
	defer batchMutex.Unlock()
	processBatchInternal(cfg)
}

// processBatchInternal does the actual batch processing (assumes lock is held)
func processBatchInternal(cfg AppConfig) {
	if len(pendingFiles) == 0 {
		return
	}

	// Create commit message based on changes
	var commitMessage string
	if len(pendingFiles) == 1 {
		for path, event := range pendingFiles {
			commitMessage = fmt.Sprintf("%s %s", strings.Title(event), path)
			break
		}
	} else {
		commitMessage = fmt.Sprintf("Batch update: %d files changed", len(pendingFiles))
	}

	// Commit all changes at once
	commitHash, err := CommitChanges(cfg.RootDir, commitMessage)
	if err != nil {
		log.Printf("Error committing batched changes: %v", err)
		pendingFiles = make(map[string]string) // Clear pending files even on error
		return
	}
	
	// If no commit hash, it means there was nothing to commit
	if commitHash == "" {
		log.Printf("[BATCH] No changes to commit for batch (working tree was already clean)")
		pendingFiles = make(map[string]string)
		return
	}

	if commitHash != "" {
		// Generate patch for the commit
		patch, err := GetPatch(cfg.RootDir, commitHash)
		if err != nil {
			log.Printf("Error getting patch for batched commit: %v", err)
		} else {
			// Create file changes for all files in the batch
			mu.Lock()
			for path, event := range pendingFiles {
				changes = append(changes, FileChange{
					File:       path,
					Event:      event,
					CommitHash: commitHash,
					Patch:      patch,
				})
			}
			mu.Unlock()
		}
	}

	// Clear pending files
	pendingFiles = make(map[string]string)

	// Reset timer
	batchTimer = nil
}

// getDynamicBatchDuration calculates batch duration based on recent activity
func getDynamicBatchDuration() time.Duration {
	dynamicBatchMux.Lock()
	defer dynamicBatchMux.Unlock()

	// Reset counter every minute
	if time.Since(lastEventReset) > time.Minute {
		recentEventCount = 0
		lastEventReset = time.Now()
	}

	recentEventCount++

	// Calculate events per second in the last minute
	eventsPerSecond := float64(recentEventCount) / time.Since(lastEventReset).Seconds()

	// Adjust batch duration based on activity
	// High activity (>5 events/sec): wait longer (5 seconds) to batch more
	// Medium activity (1-5 events/sec): normal wait (2 seconds)
	// Low activity (<1 event/sec): quick response (1 second)
	switch {
	case eventsPerSecond > 5:
		return 5 * time.Second
	case eventsPerSecond > 1:
		return 2 * time.Second
	default:
		return 1 * time.Second
	}
}

// addToBatch adds a file change to the pending batch and starts/resets the timer
func addToBatch(cfg AppConfig, filePath, eventType string) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	// Add to pending files (this will overwrite if the same file has multiple events)
	pendingFiles[filePath] = eventType

	// Calculate dynamic batch duration
	dynamicDuration := getDynamicBatchDuration()

	// Reset the timer with dynamic duration
	if batchTimer != nil {
		batchTimer.Stop()
	}

	batchTimer = time.AfterFunc(dynamicDuration, func() {
		processBatch(cfg)
	})

	// Log when duration changes significantly
	if dynamicDuration != batchDuration {
		log.Printf("[BATCH] Adjusted batch window to %v based on activity", dynamicDuration)
		batchDuration = dynamicDuration
	}
}

// ForceProcessPendingBatch processes any pending batch changes before shutdown
func ForceProcessPendingBatch(cfg AppConfig) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	// Stop any pending timer first
	if batchTimer != nil {
		batchTimer.Stop()
		batchTimer = nil
	}

	if len(pendingFiles) > 0 {
		log.Printf("[SHUTDOWN] Processing %d pending changes before exit", len(pendingFiles))
		processBatchInternal(cfg) // Call internal version since we already hold the lock
	}
}

// CleanupWatcherState clears all global watcher state
func CleanupWatcherState() {
	mu.Lock()
	defer mu.Unlock()
	
	batchMutex.Lock()
	defer batchMutex.Unlock()
	
	// Clear all state
	changes = nil
	lastEventTime = make(map[string]time.Time)
	pendingFiles = make(map[string]string)
	
	// Stop any running timer
	if batchTimer != nil {
		batchTimer.Stop()
		batchTimer = nil
	}
	
	log.Println("[SHUTDOWN] Cleared watcher state")
}

// WatchDirectory watches a directory and all subdirectories
func WatchDirectory(ctx context.Context, cfg AppConfig) {
	defer log.Println("[WATCHER] File watcher stopped")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	log.Printf("[WATCHER] Watching directory: %s", cfg.RootDir)

	// Start cleanup goroutine for lastEventTime map
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cleanupOldEventTimes()
				eventTimeMutex.RLock()
				mapSize := len(lastEventTime)
				eventTimeMutex.RUnlock()
				log.Printf("[WATCHER] Cleaned up old event times, map size: %d", mapSize)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Recursively add existing directories
	err = filepath.Walk(cfg.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if isIgnored(path, cfg.IgnorePatterns) {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if getIsApplyingPatch() {
					continue
				}

				if isIgnored(event.Name, cfg.IgnorePatterns) {
					continue
				}

				// Make path relative
				relPath, err := filepath.Rel(cfg.RootDir, event.Name)
				if err != nil {
					log.Printf("Could not find relative path for %s: %v", event.Name, err)
					continue
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						// Check file size and type before processing
						if skip, reason := shouldSkipFile(event.Name); skip {
							log.Printf("[WATCHER] Skipping %s: %s", relPath, reason)
							continue
						}
						addToBatch(cfg, relPath, "created")

						// Check if the created path is a directory. If so, walk it and add all subdirectories to the watcher.
						info, err := os.Stat(event.Name)
						if err == nil && info.IsDir() {
							err := filepath.Walk(event.Name, func(path string, fi os.FileInfo, err error) error {
								if err != nil {
									return err
								}
								if fi.IsDir() {
									if isIgnored(path, cfg.IgnorePatterns) {
										return filepath.SkipDir
									}
									err = watcher.Add(path)
									if err != nil {
									}
								}
								return nil
							})
							if err != nil {
							}
						}
					}
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						// Check file size and type before processing
						if skip, reason := shouldSkipFile(event.Name); skip {
							log.Printf("[WATCHER] Skipping %s: %s", relPath, reason)
							continue
						}
						addToBatch(cfg, relPath, "modified")
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						addToBatch(cfg, relPath, "deleted")
					}
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					// On rename, fsnotify might remove the old path from the watcher.
					// We might need to re-add the new path if it's a directory.
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						watcher.Add(event.Name)
					}
					addToBatch(cfg, relPath, "renamed")
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[WATCHER] Error: %v", err)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start polling changes
	go pollChanges(ctx, cfg)

	// Block until context is cancelled
	<-ctx.Done()
}

// pollChanges writes changes to a JSON file and publishes to Redis every 5 seconds
func pollChanges(ctx context.Context, cfg AppConfig) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			if len(changes) == 0 {
				mu.Unlock()
				continue
			}

			// Create metadata
			metadata := SyncMetadata{
				Version:   1,
				Timestamp: time.Now().Unix(),
				PeerID:    cfg.Username, // Use username from config
				Changes:   changes,
			}

			// Publish metadata to Redis
			channel := fmt.Sprintf("axle:team:%s", cfg.TeamID)
			if err := PublishMessage(ctx, cfg.RedisClient, channel, metadata); err != nil {
				log.Println("Error publishing metadata to Redis:", err)
			} else {
				log.Printf("[SYNC] Published batch with %d changes to team %s", len(metadata.Changes), cfg.TeamID)
			}

			// Clear changes after publishing
			changes = nil
			mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
