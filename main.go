package main

import (
	"axle/utils" // Replace with actual module name
	"log"
	"os"
)

func main() {
	dir := "./test_folder"
	err := os.MkdirAll(dir, os.ModePerm) // Ensure directory exists
	if err != nil {
		log.Fatal(err)
	}

	// Start watching for file changes
	utils.WatchDirectory(dir)
}
