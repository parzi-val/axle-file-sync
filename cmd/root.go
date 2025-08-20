package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"axle/utils"
)

var (
	// Global config that will be populated from local config file
	config utils.AppConfig
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "axle",
	Short: "A fast, real-time file synchronization tool for development teams",
	Long: utils.RenderTitle("🔄 Axle - Team File Synchronization") + `

Axle is a powerful, real-time file synchronization tool designed for development teams.
It uses Redis for coordination and Git for version control to keep your team's
codebase perfectly synchronized across all members.

Features:
• Real-time file synchronization using Redis pub/sub
• Git-based version control with automatic commits
• Team presence tracking and status monitoring  
• Conflict resolution and patch management
• Cross-platform support (Windows, macOS, Linux)

Use "axle [command] --help" for more information about a command.`,
	
	// This runs when no subcommands are called
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.RenderTitle("🔄 Axle"))
		fmt.Println("Welcome to Axle! Use --help to see available commands.")
		fmt.Println("")
		fmt.Println(utils.RenderInfo("Get started:"))
		fmt.Println("  axle init     - Initialize a new team repository")
		fmt.Println("  axle start    - Start file synchronization")
		fmt.Println("  axle team     - View team member status")
		fmt.Println("  axle chat     - Send a message to your team")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(utils.RenderError(err.Error()))
		os.Exit(1)
	}
}

// loadConfig loads the configuration from the local JSON file and ensures a persistent NodeID.
func loadConfig() error {
	localCfg, err := loadConfigFromFile()
	if err != nil {
		return err
	}

	// Generate and save unique node ID if not present
	if localCfg.NodeID == "" {
		localCfg.NodeID = utils.GenerateNodeID()
		if err := saveConfigToFile(localCfg); err != nil {
			return fmt.Errorf("failed to save updated config with new NodeID: %w", err)
		}
	}

	// Populate global runtime config from loaded local config
	config.NodeID = localCfg.NodeID
	config.TeamID = localCfg.TeamID
	config.Username = localCfg.Username
	config.RootDir = localCfg.RootDir
	config.RedisAddr = fmt.Sprintf("%s:%d", localCfg.RedisHost, localCfg.RedisPort)
	config.IgnorePatterns = localCfg.IgnorePatterns

	// Initialize Redis client
	rdb, err := utils.NewRedisClient(config.RedisAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis at %s: %w", config.RedisAddr, err)
	}
	config.RedisClient = rdb

	return nil
}

// LocalAppConfig represents the configuration stored in a local JSON file.
type LocalAppConfig struct {
	TeamID         string   `json:"teamID"`
	Username       string   `json:"username"`
	NodeID         string   `json:"nodeID"`
	RootDir        string   `json:"rootDir"`
	RedisHost      string   `json:"redisHost"`
	RedisPort      int      `json:"redisPort"`
	IgnorePatterns []string `json:"ignorePatterns"`
}

// ConfigFilePath defines the standard location for the local Axle configuration file.
const ConfigFileName = "axle_config.json"

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

// saveConfigToFile saves the LocalAppConfig to the local JSON file.
func saveConfigToFile(localCfg LocalAppConfig) error {
	filePath := filepath.Join(".", ConfigFileName)
	jsonData, err := json.MarshalIndent(localCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal local config to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write local config to file %s: %w", filePath, err)
	}
	return nil
}
