package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"axle/utils"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start file synchronization and team collaboration",
	Long: utils.RenderTitle("🔄 Start Axle Synchronization") + `

Starts the main Axle synchronization process including:
• Real-time file watching and synchronization
• Team presence tracking and heartbeat system
• Chat message monitoring
• Automatic conflict resolution

The process runs continuously until you press Ctrl+C for graceful shutdown.
All team members running 'axle start' will be synchronized in real-time.`,
	
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := loadConfig(); err != nil {
			return fmt.Errorf("configuration error: %w. Please run 'axle init' first", err)
		}

		defer config.RedisClient.Close()

		ctx := context.Background()
		
		fmt.Println(utils.RenderTitle("🔄 Starting Axle"))
		fmt.Printf("Team: %s | User: %s | Directory: %s\n", 
			config.TeamID, config.Username, config.RootDir)
		fmt.Println(utils.RenderInfo("Press Ctrl+C to stop"))
		fmt.Println("")

		// Start Axle with presence tracking
		startAxleWithPresence(ctx, config)
		
		return nil
	},
}

// startAxleWithPresence starts Axle with integrated presence tracking
func startAxleWithPresence(ctx context.Context, cfg utils.AppConfig) {
	// Create a cancellable context for coordinated shutdown
	appCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 1. Start presence heartbeat system
	go utils.StartPresenceHeartbeat(appCtx, cfg)
	log.Printf("[PRESENCE] Started heartbeat system (Node ID: %s)", cfg.NodeID)

	// 2. Start the file system watcher
	go utils.WatchDirectory(appCtx, cfg)
	log.Println("[WATCHER] Started file system watcher")

	// 3. Start the Redis subscriber (with presence handling)
	go startRedisSubscriberWithPresence(appCtx, cfg)
	log.Println("[SUBSCRIBER] Started Redis subscriber")

	// 4. Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[AXLE] All systems started. Watching for changes and team activity...")

	// 5. Main event loop: wait for a shutdown signal
	<-sigCh
	log.Println("[AXLE] Shutdown signal received. Gracefully shutting down...")

	// Cancel all contexts to signal goroutines to stop
	cancel()

	// Give goroutines time to clean up
	log.Println("[AXLE] Waiting for goroutines to finish...")
	time.Sleep(2 * time.Second)

	// Clean up any remaining batch processing
	utils.ForceProcessPendingBatch(cfg)

	// Clean up presence information
	utils.CleanupPresence(ctx, cfg)

	// Close Redis connection
	if cfg.RedisClient != nil {
		cfg.RedisClient.Close()
	}

	// Clear any remaining mutex state
	utils.CleanupWatcherState()

	log.Println("[AXLE] Shutdown complete")
}

// startRedisSubscriberWithPresence subscribes to Redis channels including presence
func startRedisSubscriberWithPresence(ctx context.Context, cfg utils.AppConfig) {
	defer log.Println("[SUBSCRIBER] Redis subscriber stopped")

	channels := []string{
		fmt.Sprintf("axle:team:%s", cfg.TeamID),      // Sync messages
		fmt.Sprintf("axle:chat:%s", cfg.TeamID),      // Chat messages  
		fmt.Sprintf("axle:presence:%s", cfg.TeamID),  // Presence messages
	}

	pubsub, err := utils.SubscribeToChannels(ctx, cfg.RedisClient, channels...)
	if err != nil {
		log.Printf("[SUBSCRIBER] Failed to subscribe to Redis channels: %v", err)
		return
	}
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			switch msg.Channel {
			case fmt.Sprintf("axle:team:%s", cfg.TeamID):
				handleSyncMessage(ctx, cfg, msg.Payload)
			case fmt.Sprintf("axle:chat:%s", cfg.TeamID):
				handleChatMessage(msg.Payload)
			case fmt.Sprintf("axle:presence:%s", cfg.TeamID):
				utils.ProcessPresenceMessage(ctx, cfg, msg.Payload)
			}
		case <-ctx.Done():
			return
		}
	}
}

// handleSyncMessage processes file synchronization messages
func handleSyncMessage(ctx context.Context, cfg utils.AppConfig, payload string) {
	var syncMeta utils.SyncMetadata
	if err := json.Unmarshal([]byte(payload), &syncMeta); err != nil {
		log.Printf("[SYNC] Error unmarshaling sync metadata: %v", err)
		return
	}

	// Skip our own messages
	if syncMeta.PeerID == cfg.Username {
		return
	}

	// Track changed files for committing
	var changedFiles []string
	utils.SetIsApplyingPatch(true)
	
	var autoCommittedAny bool
	for _, change := range syncMeta.Changes {
		// Handle Patches (Create/Modify)
		if change.Patch != "" {
			autoCommitted, err := utils.ApplyPatch(cfg.RootDir, change.Patch)
			if err != nil {
				log.Printf("[SYNC] Error applying patch: %v", err)
			} else {
				changedFiles = append(changedFiles, change.File)
				if autoCommitted {
					autoCommittedAny = true
				}
			}
			continue
		}

		// Handle Deletion
		if change.Event == "deleted" {
			localPathToDelete := filepath.Join(cfg.RootDir, change.File)
			err := os.RemoveAll(localPathToDelete)
			if err != nil && !os.IsNotExist(err) {
				log.Printf("[SYNC] Error deleting file/directory %s: %v", localPathToDelete, err)
			} else {
				changedFiles = append(changedFiles, change.File)
			}
		}
	}
	
	// Auto-stage and commit synced changes (only if not auto-committed by git am)
	if len(changedFiles) > 0 && !autoCommittedAny {
		commitMessage := fmt.Sprintf("[SYNC] Received %d changes from %s", len(changedFiles), syncMeta.PeerID)
		log.Printf("[SYNC] Attempting to commit %d changed files: %v", len(changedFiles), changedFiles)
		
		if _, err := utils.CommitChanges(cfg.RootDir, commitMessage); err != nil {
			log.Printf("[SYNC] Error committing synced changes in directory '%s': %v", cfg.RootDir, err)
			log.Printf("[SYNC] Failed files were: %v", changedFiles)
		} else {
			log.Printf("[SYNC] Applied and committed %d changes from %s", len(changedFiles), syncMeta.PeerID)
		}
	} else if autoCommittedAny {
		log.Printf("[SYNC] Applied and committed %d changes from %s (auto-committed by git am)", len(changedFiles), syncMeta.PeerID)
	}
	
	time.Sleep(100 * time.Millisecond) // Brief pause for FS events
	utils.SetIsApplyingPatch(false)
}

// handleChatMessage processes chat messages
func handleChatMessage(payload string) {
	var chatMsg utils.ChatMessage
	if err := json.Unmarshal([]byte(payload), &chatMsg); err != nil {
		log.Printf("[CHAT] Error unmarshaling chat message: %v", err)
		return
	}
	
	timestamp := time.Unix(chatMsg.Timestamp, 0).Format("15:04:05")
	fmt.Printf("[CHAT %s] <%s> %s\n", timestamp, chatMsg.Sender, chatMsg.Message)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
