package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"axle/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join an existing Axle team",
	Long: utils.RenderTitle("🤝 Join Axle Team") + `

Joins an existing Axle team by creating a local configuration file.
You will be prompted for the team password.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if teamID == "" || username == "" {
			return fmt.Errorf("both --team and --username flags are required")
		}

		// Prompt for password
		fmt.Print("Enter the team password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password := string(bytePassword)
		fmt.Println()

		fmt.Println(utils.RenderTitle("🤝 Joining Axle Team"))

		// Get current working directory
		rootDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Connect to Redis
		fmt.Print("Connecting to Redis... ")
		redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)
		redisClient, err := utils.NewRedisClient(redisAddr)
		if err != nil {
			fmt.Println(utils.RenderError("failed"))
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
		defer redisClient.Close()
		fmt.Println(utils.RenderSuccess("done"))

		// Fetch team config from Redis
		fmt.Print("Fetching team configuration... ")
		teamConfigKey := fmt.Sprintf("axle:config:%s", teamID)
		teamConfigData, err := redisClient.Get(context.Background(), teamConfigKey).Bytes()
		if err != nil {
			fmt.Println(utils.RenderError("failed"))
			return fmt.Errorf("failed to fetch team config from Redis: %w. Make sure the team exists and the team ID is correct.", err)
		}

		var teamConfig utils.AxleConfig
		if err := json.Unmarshal(teamConfigData, &teamConfig); err != nil {
			fmt.Println(utils.RenderError("failed"))
			return fmt.Errorf("failed to unmarshal team config: %w", err)
		}
		fmt.Println(utils.RenderSuccess("done"))

		// Verify password
		fmt.Print("Verifying password... ")
		if err := bcrypt.CompareHashAndPassword([]byte(teamConfig.PasswordHash), []byte(password)); err != nil {
			fmt.Println(utils.RenderError("failed"))
			return fmt.Errorf("invalid password")
		}
		fmt.Println(utils.RenderSuccess("done"))

		// Initialize Git repository
		fmt.Print("Setting up Git repository... ")
		if err := utils.InitGitRepo(rootDir); err != nil {
			fmt.Println(utils.RenderError("failed"))
			return fmt.Errorf("failed to initialize Git repository: %w", err)
		}
		fmt.Println(utils.RenderSuccess("done"))

		// Add config to local git exclude file
		fmt.Print("Configuring Git exclusions... ")
		excludePath := filepath.Join(rootDir, ".git", "info", "exclude")
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

		// Create local config
		localCfg := LocalAppConfig{
			TeamID:         teamID,
			Username:       username,
			RootDir:        rootDir,
			RedisHost:      redisHost,
			RedisPort:      redisPort,
			IgnorePatterns: []string{".git", ConfigFileName},
		}

		// Create local configuration file
		fmt.Print("Creating local configuration file... ")
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

		fmt.Println(utils.RenderSuccess("Successfully joined the team!"))
		fmt.Println("")
		fmt.Println(utils.RenderInfo("Next steps:"))
		fmt.Println("  axle start    - Start file synchronization")
		fmt.Println("  axle team     - Check team member status")
		fmt.Println("  axle chat \"Hi team!\" - Send a message to your team")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)

	// Add flags
	joinCmd.Flags().StringVar(&teamID, "team", "", "Team ID (required)")
	joinCmd.Flags().StringVar(&username, "username", "", "Username for this Axle instance (required)")
	joinCmd.Flags().StringVar(&redisHost, "host", "localhost", "Redis server host")
	joinCmd.Flags().IntVar(&redisPort, "port", 6379, "Redis server port")

	// Mark required flags
	joinCmd.MarkFlagRequired("team")
	joinCmd.MarkFlagRequired("username")
}
