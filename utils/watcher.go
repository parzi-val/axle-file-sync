package utils

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchDirectory(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Println("Event:", event)

				// Handle different file events
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("File Created:", event.Name)
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("File Modified:", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Println("File Deleted:", event.Name)
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					fmt.Println("File Renamed:", event.Name)
				}
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

	// Keep the process running
	fmt.Println("Watching directory:", path)
	<-make(chan struct{}) // Block forever
}
