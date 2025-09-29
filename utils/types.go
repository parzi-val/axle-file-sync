// utils/types.go
package utils

import "github.com/go-redis/redis/v8"

// ChatMessage represents a single chat message sent between Axle users.
type ChatMessage struct {
	Sender    string `json:"sender"`    // Username of the sender
	Message   string `json:"message"`   // The chat message content
	Timestamp int64  `json:"timestamp"` // Unix timestamp of when the message was sent
}

// AxleConfig defines the structure for configuration stored in Redis.
// Note: RedisClient is NOT part of this struct as it's a runtime connection.
type AxleConfig struct {
	TeamID       string `json:"teamID"`
	PasswordHash string `json:"passwordHash"`
}

// PresenceInfo represents information about a team member's presence
type PresenceInfo struct {
	Username  string `json:"username"`
	Status    string `json:"status"`    // "online" or "offline"
	LastSeen  int64  `json:"lastSeen"`  // Unix timestamp
	IPAddress string `json:"ipAddress"` // IP address of the node
	NodeID    string `json:"nodeID"`    // Unique identifier for this node instance
}

// PresenceMessage represents presence-related messages
type PresenceMessage struct {
	Type      string `json:"type"`      // "heartbeat", "announce", "goodbye"
	NodeID    string `json:"nodeID"`    // Unique identifier for this node
	Username  string `json:"username"`  // Username of the sender
	IPAddress string `json:"ipAddress"` // IP address
	Timestamp int64  `json:"timestamp"` // Unix timestamp
}

// AppConfig holds the application's runtime configuration.
type AppConfig struct {
	TeamID           string
	Username         string
	RootDir          string
	RedisAddr        string
	RedisClient      *redis.Client
	IgnorePatterns   []string
	NodeID           string           // Unique identifier for this node instance
	ConflictStrategy ConflictStrategy // Strategy for handling merge conflicts
}
