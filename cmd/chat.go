package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"axle/utils"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat [message]",
	Short: "Send a message to your team members",
	Long: utils.RenderTitle("💬 Team Chat") + `

Send a message to all team members who are currently running 'axle start'.
Messages are delivered in real-time through Redis pub/sub.

Examples:
  axle chat "Hello team!"
  axle chat "Ready to review the PR"
  axle chat "Taking a break, will be back in 30 mins"`,
	
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := loadConfig(); err != nil {
			return fmt.Errorf("configuration error: %w. Please run 'axle init' first", err)
		}
		
		defer config.RedisClient.Close()

		// Join all arguments to form the message
		messageContent := strings.Join(args, " ")
		
		fmt.Println(utils.RenderTitle("💬 Sending Message"))
		fmt.Printf("To team: %s\n", config.TeamID)
		fmt.Printf("Message: \"%s\"\n", messageContent)

		// Send the chat message
		ctx := context.Background()
		if err := publishChatMessage(ctx, config, messageContent); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		fmt.Println(utils.RenderSuccess("Message sent successfully!"))
		
		return nil
	},
}

// publishChatMessage publishes a chat message to the Redis channel
func publishChatMessage(ctx context.Context, cfg utils.AppConfig, messageContent string) error {
	msg := utils.ChatMessage{
		Sender:    cfg.Username,
		Message:   messageContent,
		Timestamp: time.Now().Unix(),
	}

	chatChannel := fmt.Sprintf("axle:chat:%s", cfg.TeamID)
	return utils.PublishMessage(ctx, cfg.RedisClient, chatChannel, msg)
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
