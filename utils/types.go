// utils/types.go
package utils

// ChatMessage represents a single chat message sent between Axle users.
type ChatMessage struct {
	Sender    string `json:"sender"`    // Username of the sender
	Message   string `json:"message"`   // The chat message content
	Timestamp int64  `json:"timestamp"` // Unix timestamp of when the message was sent
}

// // SyncMetadata represents metadata about a file system change to be synchronized.
// type SyncMetadata struct {
// 	Type         string `json:"type"`          // "modified", "created", "deleted"
// 	RelativePath string `json:"relativePath"`  // Path relative to the root directory
// 	OldBlobID    string `json:"oldBlobID"`     // Previous Git blob ID (if applicable)
// 	NewBlobID    string `json:"newBlobID"`     // Current Git blob ID
// 	Patch        string `json:"patch,omitempty"` // Git diff patch (for "modified")
// }

// AxleConfig defines the structure for configuration stored in Redis.
// Note: RedisClient is NOT part of this struct as it's a runtime connection.
type AxleConfig struct {
	TeamID   string `json:"teamID"`
	Username string `json:"username"`
	RootDir  string `json:"rootDir"`
}
