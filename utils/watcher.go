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
	muApplyingPatch sync.Mutex
	isApplyingPatch bool
	// Batching variables
	pendingFiles  = make(map[string]string) // file path -> event type
	batchTimer    *time.Timer
	batchMutex    sync.Mutex
	batchDuration = 2 * time.Second // Wait 2 seconds to accumulate changes
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
	now := time.Now()
	if lastTime, exists := eventMap[key]; exists {
		if now.Sub(lastTime) < debounceTime {
			return false // Skip duplicate event
		}
	}
	eventMap[key] = now
	return true // Accept event
}

// isIgnored checks if a path should be ignored.
func isIgnored(path string, ignorePatterns []string) bool {
	// Always ignore .git folder and its contents
	if strings.Contains(path, ".git") {
		return true
	}

	for _, pattern := range ignorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// processBatch processes accumulated file changes and commits them as a batch
func processBatch(cfg AppConfig) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

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

// addToBatch adds a file change to the pending batch and starts/resets the timer
func addToBatch(cfg AppConfig, filePath, eventType string) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	// Add to pending files (this will overwrite if the same file has multiple events)
	pendingFiles[filePath] = eventType

	// Reset the timer
	if batchTimer != nil {
		batchTimer.Stop()
	}

	batchTimer = time.AfterFunc(batchDuration, func() {
		processBatch(cfg)
	})
}

// ForceProcessPendingBatch processes any pending batch changes before shutdown
func ForceProcessPendingBatch(cfg AppConfig) {
	batchMutex.Lock()
	defer batchMutex.Unlock()
	
	if len(pendingFiles) > 0 {
		log.Printf("[SHUTDOWN] Processing %d pending changes before exit", len(pendingFiles))
		processBatch(cfg)
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
