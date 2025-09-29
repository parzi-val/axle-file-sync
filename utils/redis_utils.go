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

// NewRedisClient creates and returns a new Redis client with retry logic and exponential backoff.
func NewRedisClient(addr string) (*redis.Client, error) {
	return NewRedisClientWithRetry(addr, 5, 1*time.Second)
}

// NewRedisClientWithRetry creates a Redis client with configurable retry attempts and backoff.
func NewRedisClientWithRetry(addr string, maxRetries int, initialBackoff time.Duration) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        "", // No password in our current Docker setup
		DB:              0,  // Default DB
		MaxRetries:      3,  // Internal retries per operation
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 3 * time.Second,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	})

	// Try to connect with exponential backoff
	backoff := initialBackoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := rdb.Ping(ctx).Result()
		cancel()

		if err == nil {
			log.Printf("[REDIS] Connected to %s on attempt %d", addr, attempt)
			return rdb, nil
		}

		if attempt == maxRetries {
			rdb.Close()
			return nil, fmt.Errorf("failed to connect to Redis at %s after %d attempts: %w", addr, maxRetries, err)
		}

		log.Printf("[REDIS] Connection attempt %d/%d failed, retrying in %v: %v", attempt, maxRetries, backoff, err)
		time.Sleep(backoff)
		backoff = backoff * 2 // Exponential backoff
		if backoff > 30*time.Second {
			backoff = 30 * time.Second // Cap at 30 seconds
		}
	}

	rdb.Close()
	return nil, fmt.Errorf("failed to connect to Redis at %s", addr)
}

// PublishMessage marshals the given message (interface{}) to JSON and publishes it to the specified channel.
// Includes automatic retry logic for transient failures.
func PublishMessage(ctx context.Context, rdb *redis.Client, channel string, message interface{}) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message for channel %s: %w", channel, err)
	}

	// Retry logic for publish operations
	maxRetries := 3
	retryDelay := 500 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = rdb.Publish(ctx, channel, jsonMessage).Err()
		if err == nil {
			return nil
		}

		// Check if it's a connection error
		if err == redis.Nil || err.Error() == "redis: client is closed" {
			if attempt < maxRetries {
				log.Printf("[REDIS] Publish attempt %d/%d failed, retrying in %v: %v", attempt, maxRetries, retryDelay, err)
				time.Sleep(retryDelay)
				retryDelay *= 2 // Exponential backoff
				continue
			}
		}

		return fmt.Errorf("failed to publish message to channel %s after %d attempts: %w", channel, attempt, err)
	}

	return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
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
