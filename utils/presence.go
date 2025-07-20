package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	HeartbeatInterval = 30 * time.Second
	PresenceTimeout   = 60 * time.Second
)

// GenerateNodeID creates a unique identifier for this node instance
func GenerateNodeID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("node_%d", time.Now().UnixNano())
	}
	return "node_" + hex.EncodeToString(bytes)
}

// GetLocalIPAddress attempts to get the local IP address
func GetLocalIPAddress() string {
	// Try to get the local IP by connecting to a remote address
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Fallback to localhost if we can't determine IP
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// StartPresenceHeartbeat starts sending periodic heartbeat messages
func StartPresenceHeartbeat(ctx context.Context, cfg AppConfig) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	// Send initial announce message
	if err := sendPresenceMessage(ctx, cfg, "announce"); err != nil {
		log.Printf("[PRESENCE] Failed to send announce message: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := sendPresenceMessage(ctx, cfg, "heartbeat"); err != nil {
				log.Printf("[PRESENCE] Failed to send heartbeat: %v", err)
			}
		case <-ctx.Done():
			// Send goodbye message before exiting
			if err := sendPresenceMessage(ctx, cfg, "goodbye"); err != nil {
				log.Printf("[PRESENCE] Failed to send goodbye message: %v", err)
			}
			return
		}
	}
}

// sendPresenceMessage sends a presence message to Redis
func sendPresenceMessage(ctx context.Context, cfg AppConfig, msgType string) error {
	msg := PresenceMessage{
		Type:      msgType,
		NodeID:    cfg.NodeID,
		Username:  cfg.Username,
		IPAddress: GetLocalIPAddress(),
		Timestamp: time.Now().Unix(),
	}

	channel := fmt.Sprintf("axle:presence:%s", cfg.TeamID)
	return PublishMessage(ctx, cfg.RedisClient, channel, msg)
}

// ProcessPresenceMessage processes incoming presence messages
func ProcessPresenceMessage(ctx context.Context, cfg AppConfig, payload string) {
	var msg PresenceMessage
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		log.Printf("[PRESENCE] Error unmarshaling presence message: %v", err)
		return
	}

	// Don't process our own messages
	if msg.NodeID == cfg.NodeID {
		return
	}

	// Update presence information in Redis
	presenceKey := fmt.Sprintf("axle:team:%s:presence", cfg.TeamID)
	
	switch msg.Type {
	case "announce", "heartbeat":
		info := PresenceInfo{
			Username:  msg.Username,
			Status:    "online",
			LastSeen:  msg.Timestamp,
			IPAddress: msg.IPAddress,
			NodeID:    msg.NodeID,
		}
		
		infoJSON, err := json.Marshal(info)
		if err != nil {
			log.Printf("[PRESENCE] Error marshaling presence info: %v", err)
			return
		}
		
		if err := cfg.RedisClient.HSet(ctx, presenceKey, msg.NodeID, infoJSON).Err(); err != nil {
			log.Printf("[PRESENCE] Error updating presence in Redis: %v", err)
			return
		}
		
		if msg.Type == "announce" {
			log.Printf("[PRESENCE] %s (%s) joined the team", msg.Username, msg.IPAddress)
		}
		
	case "goodbye":
		if err := cfg.RedisClient.HDel(ctx, presenceKey, msg.NodeID).Err(); err != nil {
			log.Printf("[PRESENCE] Error removing presence from Redis: %v", err)
			return
		}
		log.Printf("[PRESENCE] %s (%s) left the team", msg.Username, msg.IPAddress)
	}
}

// GetTeamPresence retrieves all team member presence information
func GetTeamPresence(ctx context.Context, cfg AppConfig) ([]PresenceInfo, error) {
	presenceKey := fmt.Sprintf("axle:team:%s:presence", cfg.TeamID)
	
	// Get all presence information from Redis
	presenceData, err := cfg.RedisClient.HGetAll(ctx, presenceKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get team presence: %w", err)
	}

	var presenceList []PresenceInfo
	currentTime := time.Now().Unix()
	
	for nodeID, infoJSON := range presenceData {
		var info PresenceInfo
		if err := json.Unmarshal([]byte(infoJSON), &info); err != nil {
			log.Printf("[PRESENCE] Error unmarshaling presence info for node %s: %v", nodeID, err)
			continue
		}
		
		// Check if the node is considered offline (no heartbeat for more than PresenceTimeout)
		if currentTime-info.LastSeen > int64(PresenceTimeout.Seconds()) {
			info.Status = "offline"
			// Optionally remove stale entries
			go func(nodeID string) {
				cfg.RedisClient.HDel(context.Background(), presenceKey, nodeID)
			}(nodeID)
		}
		
		presenceList = append(presenceList, info)
	}
	
	return presenceList, nil
}

// CleanupPresence removes this node's presence information
func CleanupPresence(ctx context.Context, cfg AppConfig) {
	presenceKey := fmt.Sprintf("axle:team:%s:presence", cfg.TeamID)
	if err := cfg.RedisClient.HDel(ctx, presenceKey, cfg.NodeID).Err(); err != nil {
		log.Printf("[PRESENCE] Error cleaning up presence: %v", err)
	}
}
