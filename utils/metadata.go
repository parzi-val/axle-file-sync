package utils

import (
	"encoding/json"
	"os"
)

// Struct for individual file changes
type FileChange struct {
	File         string `json:"file"`
	Event        string `json:"event"`
	Hash         string `json:"hash,omitempty"`          // Omit if empty
	Delta        string `json:"delta,omitempty"`         // Base64-encoded XOR diff
	FileSize     int64  `json:"file_size,omitempty"`     // Omit if empty
	PreviousHash string `json:"previous_hash,omitempty"` // Omit if empty
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
