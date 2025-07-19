// utils/redis_utils.go
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time" // Added for context timeouts or other time-based operations

	"github.com/go-redis/redis/v8"
)

// NewRedisClient creates and returns a new Redis client, performing a ping to ensure connectivity.
func NewRedisClient(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // No password in our current Docker setup
		DB:       0,  // Default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		rdb.Close() // Close the client if ping fails
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}
	log.Printf("[REDIS] Connected to %s", addr)
	return rdb, nil
}

// PublishMessage marshals the given message (interface{}) to JSON and publishes it to the specified channel.
func PublishMessage(ctx context.Context, rdb *redis.Client, channel string, message interface{}) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message for channel %s: %w", channel, err)
	}

	err = rdb.Publish(ctx, channel, jsonMessage).Err()
	if err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}
	return nil
}

// SubscribeToChannels subscribes to the given Redis channels and returns the PubSub instance.
// The caller is responsible for receiving messages from the PubSub.Channel() and closing the PubSub.
func SubscribeToChannels(ctx context.Context, rdb *redis.Client, channels ...string) (*redis.PubSub, error) {
	pubsub := rdb.Subscribe(ctx, channels...)
	_, err := pubsub.Receive(ctx) // Blocks until subscription is confirmed or error occurs
	if err != nil {
		pubsub.Close() // Ensure pubsub is closed on error
		return nil, fmt.Errorf("failed to subscribe to channels %v: %w", channels, err)
	}
	log.Printf("[REDIS] Subscribed to channels: %v", channels)
	return pubsub, nil
}
