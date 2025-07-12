package utils

import (
	"encoding/json"
	"os"
)

// Struct for individual file changes
type FileChange struct {
	File        string `json:"file"`
	Event       string `json:"event"`
	CommitHash  string `json:"commit_hash,omitempty"`
	Patch       string `json:"patch,omitempty"`
	NewBlobID   string `json:"new_blob_id,omitempty"`
	PrevBlobID  string `json:"prev_blob_id,omitempty"`
}

// Struct for batch sync metadata
type SyncMetadata struct {
	Version   int          `json:"version"`
	Timestamp int64        `json:"timestamp"`
	PeerID    string       `json:"peer_id"`
	Changes   []FileChange `json:"changes"`
}

// Save metadata to JSON file
func SaveMetadata(metadata SyncMetadata, filePath string) error {
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, 0644)
}
