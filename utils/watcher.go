package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Global variables
var changes []FileChange
var mu sync.Mutex
var lastEventTime = make(map[string]time.Time) // Generic debounce map

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

// WatchDirectory watches a directory and all subdirectories
func WatchDirectory(rootPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Watching directory:", rootPath)

	// Recursively add existing directories
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
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

				mu.Lock()

				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println(event.Name, "created")
					changes = append(changes, FileChange{File: event.Name, Event: "created"})

					// If a new folder is created, add a watcher to it
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						fmt.Println("New folder detected, adding watcher:", event.Name)
						watcher.Add(event.Name)
					}
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						fmt.Println(event.Name, "modified")
						changes = append(changes, FileChange{File: event.Name, Event: "modified"})
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					if debounceEvent(lastEventTime, event.Name, 500*time.Millisecond) {
						fmt.Println(event.Name, "deleted")
						changes = append(changes, FileChange{File: event.Name, Event: "deleted"})
					}
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					fmt.Println(event.Name, "renamed")
					changes = append(changes, FileChange{File: event.Name, Event: "renamed"})
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
	go pollChanges()

	select {} // Block forever
}

// pollChanges writes changes to a JSON file every 5 seconds
func pollChanges() {
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
			PeerID:    "peer_abc123", // Placeholder, should be dynamically set
			Changes:   changes,
		}

		// Save metadata to file
		err := SaveMetadata(metadata, "sync_metadata.json")
		if err != nil {
			log.Println("Error saving metadata:", err)
		} else {
			fmt.Println("Sync metadata saved.")
		}

		// Clear changes after writing
		changes = nil
		mu.Unlock()
	}
}
