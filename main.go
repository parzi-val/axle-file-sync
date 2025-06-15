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

	"github.com/go-redis/redis/v8"
	// "github.com/rjeczalik/notify"

	"axle/utils"
)

// ConfigFilePath defines the standard location for the local Axle configuration file.
const ConfigFileName = "axle_config.json"

// LocalAppConfig represents the configuration stored in a local JSON file.
type LocalAppConfig struct {
	TeamID    string `json:"teamID"`
	Username  string `json:"username"`
	RootDir   string `json:"rootDir"`
	RedisHost string `json:"redisHost"`
	RedisPort int    `json:"redisPort"`
}

// AppConfig holds the application's runtime configuration, derived from LocalAppConfig.
type AppConfig struct {
	TeamID      string
	Username    string
	RootDir     string
	RedisAddr   string        // "host:port" string
	RedisClient *redis.Client // Connected Redis client
}

var config AppConfig // Global runtime config

func main() {
	// --- CLI FlagSet Definitions ---
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initTeamIDFlag := initCmd.String("team", "", "Team ID (required)")
	initUsernameFlag := initCmd.String("username", "", "Username for this Axle instance (required)")
	initRootDirFlag := initCmd.String("root", ".", "Root directory to sync (default: current directory)")
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

		localCfg := LocalAppConfig{
			TeamID:    *initTeamIDFlag,
			Username:  *initUsernameFlag,
			RootDir:   *initRootDirFlag,
			RedisHost: *initRedisHostFlag,
			RedisPort: *initRedisPortFlag,
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
// It sets up Git (placeholder) and stores config in a local JSON file.
func initAxle(localCfg LocalAppConfig) error {
	// --- Placeholder for Git and local state initialization ---
	fmt.Printf("Initializing Git repository in %s... (placeholder)\n", localCfg.RootDir)
	fmt.Println("Creating axle_local_state.json... (placeholder)")
	// --- End of Placeholder ---

	// Store configuration in a local JSON file
	filePath := filepath.Join(localCfg.RootDir, ConfigFileName) // Store in the rootDir
	// Alternatively, store in a user's home directory like:
	// configDir := filepath.Join(os.Getenv("HOME"), ".axle")
	// if err := os.MkdirAll(configDir, 0755); err != nil {
	// 	return fmt.Errorf("failed to create config directory: %w", err)
	// }
	// filePath = filepath.Join(configDir, ConfigFileName)

	jsonData, err := json.MarshalIndent(localCfg, "", "  ") // Pretty print JSON
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
	// First, try to find the config file in the current working directory
	// (assuming Axle commands are run from the project root)
	filePath := filepath.Join(".", ConfigFileName) // Current directory

	// Alternatively, if you want a system-wide config:
	// configDir := filepath.Join(os.Getenv("HOME"), ".axle")
	// filePath = filepath.Join(configDir, ConfigFileName)

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
func startAxle(ctx context.Context, cfg AppConfig) {
	// 1. Start the file system watcher
	// watcherEvents, err := utils.StartFileWatcher(cfg.RootDir)
	// if err != nil {
	// 	log.Fatalf("Failed to start file watcher: %v", err)
	// }

	// var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for ev := range watcherEvents {
	// 		// Placeholder for actual file event handling
	// 		fmt.Printf("[FS_EVENT] %s: %s\n", ev.Event(), ev.Path())
	// 		// In a real scenario, you'd call utils.HandleFileEvent here,
	// 		// get SyncMetadata, and publish it to Redis.
	// 		// syncMeta, err := utils.HandleFileEvent(ev, cfg.RootDir)
	// 		// if err != nil {
	// 		//    log.Printf("Error processing file event: %v", err)
	// 		//    continue
	// 		// }
	// 		// // Marshal syncMeta and publish to Redis team channel
	// 		// err = utils.PublishMessage(ctx, cfg.RedisClient, fmt.Sprintf("axle:team:%s", cfg.TeamID), syncMeta)
	// 		// if err != nil {
	// 		//   log.Printf("Error publishing sync message: %v", err)
	// 		// }
	// 	}
	// }()

	// 2. Start the Redis subscriber (in a goroutine)
	go startRedisSubscriber(ctx, cfg.TeamID, cfg.RedisClient) // Removed username as it's in AppConfig.Username

	// 3. Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Axle started. Watching for file changes and listening for chat messages. Press Ctrl+C to stop.")

	// 4. Main event loop: wait for a shutdown signal
	<-sigCh
	fmt.Println("\nShutting down Axle...")

	// // Perform graceful shutdown:
	// notify.Stop(watcherEvents) // This stops the notify goroutine and closes the channel
	// wg.Wait()                  // Wait for the file watcher goroutine to finish processing remaining events

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

	chatChannel := fmt.Sprintf("axle:chat:%s", teamID)
	err := utils.PublishMessage(ctx, rdb, chatChannel, msg)
	if err != nil {
		log.Printf("Error publishing chat message to Redis: %v", err)
		return
	}
	fmt.Printf("Chat message sent by %s to '%s': %s\n", username, chatChannel, messageContent)
}

// startRedisSubscriber subscribes to Redis channels and processes messages.
func startRedisSubscriber(ctx context.Context, teamID string, rdb *redis.Client) {
	defer log.Println("Redis subscriber stopped.")

	// Subscribe to both chat and team synchronization channels
	channels := []string{
		fmt.Sprintf("axle:team:%s", teamID),
		fmt.Sprintf("axle:chat:%s", teamID),
	}

	pubsub, err := utils.SubscribeToChannels(ctx, rdb, channels...)
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
			case fmt.Sprintf("axle:team:%s", teamID):
				// This is a file synchronization message
				fmt.Printf("[SYNC] From channel '%s': %s\n", msg.Channel, msg.Payload)

			case fmt.Sprintf("axle:chat:%s", teamID):
				// This is a chat message
				var chatMsg utils.ChatMessage
				if err := json.Unmarshal([]byte(msg.Payload), &chatMsg); err != nil {
					log.Printf("Error unmarshaling chat message: %v", err)
					continue
				}
				fmt.Printf("[CHAT %s] <%s> %s\n", time.Unix(chatMsg.Timestamp, 0).Format("15:04:05"), chatMsg.Sender, chatMsg.Message)
			default:
				fmt.Printf("Received unexpected message from channel '%s': %s\n", msg.Channel, msg.Payload)
			}
		case <-ctx.Done():
			return
		}
	}
}
