package utils

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Snapshot structure to hold file data
var snapshots = make(map[string]string)

// CreateSnapshots scans the directory, reads files, and stores hex snapshots
func CreateSnapshots(rootDir, snapshotFile string) error {
	snapshots = make(map[string]string) // Reset snapshots

	// Walk through all files in the directory
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // Skip directories
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Store file content as hex
		snapshots[strings.ReplaceAll(path, "\\", "/")] = hex.EncodeToString(data)
		return nil
	})

	if err != nil {
		return err
	}

	// Save snapshots to file
	return saveSnapshotToFile(snapshotFile)
}

// saveSnapshotToFile writes the snapshot map to a JSON file
func saveSnapshotToFile(snapshotFile string) error {
	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(snapshotFile, data, 0644)
}
