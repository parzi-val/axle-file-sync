package utils

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Global list to store changes within a polling window
var changes []FileChange
var mu sync.Mutex // To prevent race conditions
var lastModified = make(map[string]time.Time)

// WatchDirectory watches the specified path for file events
func WatchDirectory(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Watching directory:", path)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				mu.Lock()

				if event.Op&fsnotify.Create == fsnotify.Create {
					changes = append(changes, FileChange{File: event.Name, Event: "created"})
					fmt.Println(event.Name, "created")
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					// Debounce to avoid duplicate writes
					now := time.Now()
					if lastTime, exists := lastModified[event.Name]; exists {
						if now.Sub(lastTime) < 500*time.Millisecond {
							mu.Unlock()
							continue
						}
					}
					lastModified[event.Name] = now

					changes = append(changes, FileChange{File: event.Name, Event: "modified"})
					fmt.Println(event.Name, "modified")
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					changes = append(changes, FileChange{File: event.Name, Event: "deleted"})
					fmt.Println(event.Name, "deleted")
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					changes = append(changes, FileChange{File: event.Name, Event: "renamed"})
					fmt.Println(event.Name, "renamed")
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

	// Watch the specified directory
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

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
