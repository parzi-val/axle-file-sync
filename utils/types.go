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
	TeamID   string `json:"teamID"`
	Username string `json:"username"`
	RootDir  string `json:"rootDir"`
}

// AppConfig holds the application's runtime configuration.
type AppConfig struct {
	TeamID         string
	Username       string
	RootDir        string
	RedisAddr      string
	RedisClient    *redis.Client
	IgnorePatterns []string
}
