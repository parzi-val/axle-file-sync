package main

import (
	"axle/utils"
	"fmt"
	"log"
	"os"
)

func main() {
	dir := "./test_folder"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting file watcher and XOR sync...")
	go utils.WatchDirectory(dir) // Start directory monitoring

	// Example XOR usage
	oldFile := "example/old.txt"
	newFile := "example/new.txt"
	deltaFile := "outputs/delta.bin"
	reconstructedFile := "outputs/reconstructed.txt"

	delta, err := utils.ComputeXORDelta(oldFile, newFile)
	if err != nil {
		log.Fatal("Error computing XOR:", err)
	}
	err = os.WriteFile(deltaFile, delta, 0644)
	if err != nil {
		log.Fatal("Error saving delta file:", err)
	}
	fmt.Println("XOR delta saved.")

	// Test reconstruction
	err = utils.ApplyXORDelta(oldFile, delta, reconstructedFile)
	if err != nil {
		log.Fatal("Error reconstructing file:", err)
	}
	fmt.Println("Reconstructed file created.")

	// Keep main running
	select {}
}
