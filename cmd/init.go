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
	teamID    string
	username  string
	redisHost string
	redisPort int
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Axle repository for team collaboration",
	Long: utils.RenderTitle("🚀 Initialize Axle Repository") + `

Sets up a new Axle repository in the current directory with team configuration.
This creates the necessary Git repository, configuration files, and sets up
the environment for real-time file synchronization.

After initialization, you can use 'axle start' to begin synchronization
and 'axle team' to see who's online.`,
	
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if teamID == "" || username == "" {
			return fmt.Errorf("both --team and --username flags are required")
		}

		fmt.Println(utils.RenderTitle("🚀 Initializing Axle Repository"))
		
		// Get current working directory
		rootDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Create local config
		localCfg := LocalAppConfig{
			TeamID:         teamID,
			Username:       username,
			RootDir:        rootDir,
			RedisHost:      redisHost,
			RedisPort:      redisPort,
			IgnorePatterns: []string{".git", ConfigFileName},
		}

		// Initialize Axle environment
		if err := initAxleRepo(localCfg); err != nil {
			return fmt.Errorf("failed to initialize Axle: %w", err)
		}

		fmt.Println(utils.RenderSuccess("Axle repository initialized successfully!"))
		fmt.Println("")
		fmt.Println(utils.RenderInfo("Next steps:"))
		fmt.Println("  axle start    - Start file synchronization")
		fmt.Println("  axle team     - Check team member status")
		fmt.Println("  axle chat \"Hi team!\" - Send a message to your team")

		return nil
	},
}

// initAxleRepo initializes the Axle environment
func initAxleRepo(localCfg LocalAppConfig) error {
	// Initialize Git repository
	fmt.Print("Setting up Git repository... ")
	if err := utils.InitGitRepo(localCfg.RootDir); err != nil {
		fmt.Println(utils.RenderError("failed"))
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}
	fmt.Println(utils.RenderSuccess("done"))

	// Add config to local git exclude file
	fmt.Print("Configuring Git exclusions... ")
	excludePath := filepath.Join(localCfg.RootDir, ".git", "info", "exclude")
	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(utils.RenderError("failed"))
		return fmt.Errorf("failed to open .git/info/exclude: %w", err)
	}
	defer f.Close()
	
	if _, err := f.WriteString("\n" + ConfigFileName + "\n"); err != nil {
		fmt.Println(utils.RenderError("failed"))
		return fmt.Errorf("failed to write to .git/info/exclude: %w", err)
	}
	fmt.Println(utils.RenderSuccess("done"))

	// Store configuration in local JSON file
	fmt.Print("Creating configuration file... ")
	filePath := filepath.Join(localCfg.RootDir, ConfigFileName)
	jsonData, err := json.MarshalIndent(localCfg, "", "  ")
	if err != nil {
		fmt.Println(utils.RenderError("failed"))
		return fmt.Errorf("failed to marshal local config to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		fmt.Println(utils.RenderError("failed"))
		return fmt.Errorf("failed to write local config to file %s: %w", filePath, err)
	}
	fmt.Println(utils.RenderSuccess("done"))

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Add flags
	initCmd.Flags().StringVar(&teamID, "team", "", "Team ID (required)")
	initCmd.Flags().StringVar(&username, "username", "", "Username for this Axle instance (required)")
	initCmd.Flags().StringVar(&redisHost, "host", "localhost", "Redis server host")
	initCmd.Flags().IntVar(&redisPort, "port", 6379, "Redis server port")

	// Mark required flags
	initCmd.MarkFlagRequired("team")
	initCmd.MarkFlagRequired("username")
}
