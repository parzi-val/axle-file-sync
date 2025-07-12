package utils

import (
	"context"
	"encoding/json"
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
)

// SetIsApplyingPatch sets the state of the patch application flag.
// This is used to temporarily mute the watcher.
func SetIsApplyingPatch(state bool) {
	muApplyingPatch.Lock()
	defer muApplyingPatch.Unlock()
	isApplyingPatch = state
	if state {
		log.Println("Watcher is now muted for patch application.")
	} else {
		log.Println("Watcher is now active.")
	}
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
	for _, pattern := range ignorePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// WatchDirectory watches a directory and all subdirectories
func WatchDirectory(cfg AppConfig) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Watching directory:", cfg.RootDir)

	// Recursively add existing directories
	err = filepath.Walk(cfg.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if isIgnored(path, cfg.IgnorePatterns) {
				return filepath.SkipDir
			}
			fmt.Println("Adding watcher to:", path)
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
					log.Printf("Skipping event for %s due to ongoing patch application\n", event.Name)
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

				mu.Lock()

				if event.Op&fsnotify.Create == fsnotify.Create {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						fmt.Println(relPath, "created")
						commitHash, err := CommitChanges(cfg.RootDir, fmt.Sprintf("Created %s", relPath))
						if err != nil {
							log.Println("Error committing changes:", err)
						} else if commitHash != "" {
							patch, err := GetPatch(cfg.RootDir, commitHash)
							if err != nil {
								log.Println("Error getting patch:", err)
							} else {
								changes = append(changes, FileChange{File: relPath, Event: "created", CommitHash: commitHash, Patch: patch})
							}
						}
					}
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
								log.Printf("Adding watcher to new subdirectory: %s", path)
								err = watcher.Add(path)
								if err != nil {
									log.Printf("Failed to add watcher to %s: %v", path, err)
								}
							}
							return nil
						})
						if err != nil {
							log.Printf("Error walking new directory %s: %v", event.Name, err)
						}
					}
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						fmt.Println(relPath, "modified")
						commitHash, err := CommitChanges(cfg.RootDir, fmt.Sprintf("Modified %s", relPath))
						if err != nil {
							log.Println("Error committing changes:", err)
						} else if commitHash != "" {
							patch, err := GetPatch(cfg.RootDir, commitHash)
							if err != nil {
								log.Println("Error getting patch:", err)
							} else {
								changes = append(changes, FileChange{File: relPath, Event: "modified", CommitHash: commitHash, Patch: patch})
							}
						}
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						fmt.Println(relPath, "deleted")
						changes = append(changes, FileChange{File: relPath, Event: "deleted"})
					}
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					// On rename, fsnotify might remove the old path from the watcher.
					// We might need to re-add the new path if it's a directory.
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						log.Printf("Re-adding watcher to renamed directory: %s", event.Name)
						watcher.Add(event.Name)
					}
					fmt.Println(relPath, "renamed")
					changes = append(changes, FileChange{File: relPath, Event: "renamed"})
				}

				mu.Unlock()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	// Start polling changes
	go pollChanges(cfg)

	select {} // Block forever
}

// pollChanges writes changes to a JSON file and publishes to Redis every 5 seconds
func pollChanges(cfg AppConfig) {
	for {
		time.Sleep(5 * time.Second)

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

		// Log the metadata before publishing
		metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			log.Println("Error marshaling metadata for logging:", err)
		} else {
			log.Printf("Publishing SyncMetadata:\n%s\n", string(metadataJSON))
		}

		// Publish metadata to Redis
		ctx := context.Background()
		channel := fmt.Sprintf("axle:team:%s", cfg.TeamID)
		if err := PublishMessage(ctx, cfg.RedisClient, channel, metadata); err != nil {
			log.Println("Error publishing metadata to Redis:", err)
		} else {
			fmt.Println("Sync metadata published to Redis.")
		}

		// Clear changes after publishing
		changes = nil
		mu.Unlock()
	}
}
