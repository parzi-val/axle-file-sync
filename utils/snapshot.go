package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Snapshot structure to hold file data
var snapshots = make(map[string]string)

// Load ignore patterns from .axleignore
func loadIgnorePatterns(rootDir string) (map[string]bool, error) {
	ignorePatterns := make(map[string]bool)
	ignoreFile := filepath.Join(rootDir, ".axleignore")

	data, err := os.ReadFile(ignoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return ignorePatterns, nil // No ignore file, return empty map
		}
		return nil, err
	}

	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			ignorePatterns[trimmed] = true
		}
	}
	return ignorePatterns, nil
}

// isIgnored checks if a file or directory should be ignored
func isIgnored(path string, ignorePatterns map[string]bool) bool {
	// Normalize path
	path = strings.ReplaceAll(path, `\`, "/")

	if ignorePatterns[path] {
		return true
	}

	baseName := filepath.Base(path)
	if ignorePatterns[baseName] {
		return true
	}

	// Split path into directories and check each
	directories := filepath.Dir(path)
	parts := strings.Split(strings.ReplaceAll(directories, `\`, "/"), "/")[1:]
	for i := range parts {
		dirPath := parts[i] + "/"
		if ignorePatterns[dirPath] {
			return true
		}
	}

	return false
}

// CreateSnapshots scans the directory, reads files, and stores hex snapshots
func CreateSnapshots(rootDir, snapshotFile string) error {
	snapshots = make(map[string]string) // Reset snapshots

	ignoreFilePath := filepath.Join(rootDir, ".axleignore")
	ignoredEntries := []string{
		".axleignore",
		".git/",
		".env",
	}

	// Write default ignores to .axleignore
	err := os.WriteFile(ignoreFilePath, []byte(strings.Join(ignoredEntries, "\n")), 0644)
	if err != nil {
		return err
	}
	fmt.Println("Created .axleignore file...")

	// Load ignore patterns
	ignorePatterns, err := loadIgnorePatterns(rootDir)
	if err != nil {
		return err
	}

	// Walk through all files in the directory
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // Skip directories
		}
		if isIgnored(path, ignorePatterns) {
			return nil // Skip ignored files
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
