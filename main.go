package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"axle/utils"

	"github.com/go-redis/redis/v8"
)

// ConfigFilePath defines the standard location for the local Axle configuration file.
const ConfigFileName = "axle_config.json"

// LocalAppConfig represents the configuration stored in a local JSON file.
type LocalAppConfig struct {
	TeamID         string   `json:"teamID"`
	Username       string   `json:"username"`
	RootDir        string   `json:"rootDir"`
	RedisHost      string   `json:"redisHost"`
	RedisPort      int      `json:"redisPort"`
	IgnorePatterns []string `json:"ignorePatterns"`
}

var config utils.AppConfig // Global runtime config

func main() {
	// --- CLI FlagSet Definitions ---
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initTeamIDFlag := initCmd.String("team", "", "Team ID (required)")
	initUsernameFlag := initCmd.String("username", "", "Username for this Axle instance (required)")
	initRedisHostFlag := initCmd.String("host", "localhost", "Redis server host (default: localhost)")
	initRedisPortFlag := initCmd.Int("port", 6379, "Redis server port (default: 6379)")

	startCmd := flag.NewFlagSet("start", flag.ExitOnError)

	chatCmd := flag.NewFlagSet("chat", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Context for operations
	ctx := context.Background()

	switch os.Args[1] {
	case "init":
		err := initCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		if *initTeamIDFlag == "" || *initUsernameFlag == "" {
			fmt.Println("Error: --team and --username are required for init command")
			initCmd.PrintDefaults()
			os.Exit(1)
		}

		rootDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}

		localCfg := LocalAppConfig{
			TeamID:         *initTeamIDFlag,
			Username:       *initUsernameFlag,
			RootDir:        rootDir,
			RedisHost:      *initRedisHostFlag,
			RedisPort:      *initRedisPortFlag,
			IgnorePatterns: []string{".git", ConfigFileName}, // Watcher ignore
		}

		if err := initAxle(localCfg); err != nil {
			log.Fatalf("Failed to initialize Axle: %v", err)
		}
		fmt.Println("Axle initialized successfully. Configuration saved to " + ConfigFileName)

	case "start":
		err := startCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}

		// Load configuration from local JSON file
		localCfg, err := loadConfigFromFile()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v. Please run 'axle init' first.", err)
		}

		// Populate global runtime config from loaded local config
		config.TeamID = localCfg.TeamID
		config.Username = localCfg.Username
		config.RootDir = localCfg.RootDir
		config.RedisAddr = fmt.Sprintf("%s:%d", localCfg.RedisHost, localCfg.RedisPort)
		config.IgnorePatterns = localCfg.IgnorePatterns

		// Initialize Redis client using the address from local config
		rdb, err := utils.NewRedisClient(config.RedisAddr)
		if err != nil {
			log.Fatalf("Failed to connect to Redis at %s: %v", config.RedisAddr, err)
		}
		config.RedisClient = rdb // Assign to global config

		fmt.Printf("Starting Axle for team '%s' as user '%s' in '%s'.\n", config.TeamID, config.Username, config.RootDir)
		startAxle(ctx, config) // Pass context and config

	case "chat":
		err := chatCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		if len(os.Args) < 3 { // Check if message content is provided
			fmt.Println("Usage: axle chat \"<message>\"")
			os.Exit(1)
		}
		messageContent := strings.Join(os.Args[2:], " ") // Handle multi-word messages

		// Load configuration from local JSON file
		localCfg, err := loadConfigFromFile()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v. Please run 'axle init' first.", err)
		}

		// Populate global config for chat command's needs
		config.TeamID = localCfg.TeamID
		config.Username = localCfg.Username
		config.RedisAddr = fmt.Sprintf("%s:%d", localCfg.RedisHost, localCfg.RedisPort)

		// Initialize Redis client for chat operation
		rdb, err := utils.NewRedisClient(config.RedisAddr)
		if err != nil {
			log.Fatalf("Failed to connect to Redis at %s for chat: %v", config.RedisAddr, err)
		}
		defer rdb.Close() // Close connection after chat operation

		publishChatMessage(ctx, config.TeamID, config.Username, messageContent, rdb)
		return // Exit after publishing chat message

	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays the command-line usage.
func printUsage() {
	fmt.Println("Usage: axle <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init   - Initializes a new Axle repository for a team and creates local config.")
	fmt.Println("  start  - Starts the Axle synchronization and chat listener (reads from local config).")
	fmt.Println("  chat   - Sends a chat message to the team (reads from local config).")
}

// initAxle initializes the Axle environment.
func initAxle(localCfg LocalAppConfig) error {
	// --- Initialize Git Repository ---
	if err := utils.InitGitRepo(localCfg.RootDir); err != nil {
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}
	fmt.Println("Git repository initialized successfully.")

	// --- Add config to local git exclude file ---
	excludePath := filepath.Join(localCfg.RootDir, ".git", "info", "exclude")
	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .git/info/exclude: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString("\n" + ConfigFileName + "\n"); err != nil {
		return fmt.Errorf("failed to write to .git/info/exclude: %w", err)
	}
	fmt.Println("Added config file to .git/info/exclude.")

	// --- Store configuration in a local JSON file ---
	filePath := filepath.Join(localCfg.RootDir, ConfigFileName)
	jsonData, err := json.MarshalIndent(localCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal local config to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write local config to file %s: %w", filePath, err)
	}
	return nil
}

// loadConfigFromFile reads the LocalAppConfig from the local JSON file.
func loadConfigFromFile() (LocalAppConfig, error) {
	filePath := filepath.Join(".", ConfigFileName)
	jsonData, err := os.ReadFile(filePath)
		if err != nil {
		return LocalAppConfig{}, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var localCfg LocalAppConfig
	if err := json.Unmarshal(jsonData, &localCfg); err != nil {
		return LocalAppConfig{}, fmt.Errorf("failed to unmarshal config JSON from %s: %w", filePath, err)
	}
	return localCfg, nil
}

// startAxle starts the main Axle synchronization and chat processes.
func startAxle(ctx context.Context, cfg utils.AppConfig) {
	// 1. Start the file system watcher
	go utils.WatchDirectory(cfg)

	// 2. Start the Redis subscriber (in a goroutine)
	go startRedisSubscriber(ctx, cfg)

	// 3. Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Axle started. Watching for file changes and listening for chat messages. Press Ctrl+C to stop.")

	// 4. Main event loop: wait for a shutdown signal
	<-sigCh
	fmt.Println("\nShutting down Axle...")

	// Close Redis connection
	if cfg.RedisClient != nil {
		cfg.RedisClient.Close()
	}
	fmt.Println("Axle shutdown complete.")
}

// publishChatMessage publishes a chat message to the Redis channel.
func publishChatMessage(ctx context.Context, teamID, username, messageContent string, rdb *redis.Client) {
	msg := utils.ChatMessage{
		Sender:    username,
		Message:   messageContent,
		Timestamp: time.Now().Unix(),
	}

	chatChannel := fmt.Sprintf("axle:team:%s", teamID)
	err := utils.PublishMessage(ctx, rdb, chatChannel, msg)
	if err != nil {
		log.Printf("Error publishing chat message to Redis: %v", err)
		return
	}
	fmt.Printf("Chat message sent by %s to '%s': %s\n", username, chatChannel, messageContent)
}

// startRedisSubscriber subscribes to Redis channels and processes messages.
func startRedisSubscriber(ctx context.Context, cfg utils.AppConfig) {
	defer log.Println("Redis subscriber stopped.")

	channels := []string{
		fmt.Sprintf("axle:team:%s", cfg.TeamID),
		fmt.Sprintf("axle:chat:%s", cfg.TeamID),
	}

	pubsub, err := utils.SubscribeToChannels(ctx, cfg.RedisClient, channels...)
	if err != nil {
		log.Printf("Failed to subscribe to Redis channels: %v", err)
		return
	}
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			switch msg.Channel {
			case fmt.Sprintf("axle:team:%s", cfg.TeamID):
				var syncMeta utils.SyncMetadata
				if err := json.Unmarshal([]byte(msg.Payload), &syncMeta); err != nil {
					log.Printf("Error unmarshaling sync metadata: %v", err)
					continue
				}

				if syncMeta.PeerID == cfg.Username {
					continue
				}

				fmt.Printf("[SYNC] Received sync from %s\n", syncMeta.PeerID)

				for _, change := range syncMeta.Changes {
					// Handle Patches (Create/Modify)
					if change.Patch != "" {
						utils.SetIsApplyingPatch(true)
						if err := utils.ApplyPatch(cfg.RootDir, change.Patch); err != nil {
							log.Printf("Error applying patch: %v", err)
						}
						time.Sleep(100 * time.Millisecond) // Brief pause for FS events
						utils.SetIsApplyingPatch(false)
						continue // Continue to next change
					}

					// Handle Deletion
					if change.Event == "deleted" {
						// Construct the absolute path of the file to be deleted on the local system
						localPathToDelete := filepath.Join(cfg.RootDir, change.File)

						// Mute watcher during deletion
						utils.SetIsApplyingPatch(true)
						log.Printf("Attempting to delete: %s", localPathToDelete)
						err := os.RemoveAll(localPathToDelete) // Use RemoveAll to handle files and directories
						if err != nil {
							// Check if the error is because the file is already gone
							if !os.IsNotExist(err) {
								log.Printf("Error deleting file/directory %s: %v", localPathToDelete, err)
							}
						} else {
							log.Printf("Successfully deleted: %s", localPathToDelete)
						}
						// Un-mute watcher
						time.Sleep(100 * time.Millisecond) // Brief pause for FS events
						utils.SetIsApplyingPatch(false)
					}
				}

			case fmt.Sprintf("axle:chat:%s", cfg.TeamID):
				var chatMsg utils.ChatMessage
				if err := json.Unmarshal([]byte(msg.Payload), &chatMsg); err != nil {
					log.Printf("Error unmarshaling chat message: %v", err)
					continue
				}
				fmt.Printf("[CHAT %s] <%s> %s\n", time.Unix(chatMsg.Timestamp, 0).Format("15:04:05"), chatMsg.Sender, chatMsg.Message)
			}
		case <-ctx.Done():
			return
		}
	}
}