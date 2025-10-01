package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/parzi-val/axle-file-sync/utils"
)

var priorityFlag bool

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat [message]",
	Short: "Send a message to your team members",
	Long: utils.RenderTitle("💬 Team Chat") + `

Send a message to all team members who are currently running 'axle start'.
Messages are delivered in real-time through Redis pub/sub.

Use -p flag to send priority messages that trigger desktop notifications.

Examples:
  axle chat "Hello team!"
  axle chat "Ready to review the PR"
  axle chat -p "URGENT: Production issue needs immediate attention!"
  axle chat --priority "Please review this ASAP"`,
	
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
		if priorityFlag {
			fmt.Printf("Priority: 🔔 HIGH (will trigger notifications)\n")
		}

		// Send the chat message
		ctx := context.Background()
		if err := publishChatMessage(ctx, config, messageContent); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		if priorityFlag {
			fmt.Println(utils.RenderSuccess("Priority message sent with notifications!"))
		} else {
			fmt.Println(utils.RenderSuccess("Message sent successfully!"))
		}
		
		return nil
	},
}

// publishChatMessage publishes a chat message to the Redis channel
func publishChatMessage(ctx context.Context, cfg utils.AppConfig, messageContent string) error {
	msg := utils.ChatMessage{
		Sender:    cfg.Username,
		Message:   messageContent,
		Timestamp: time.Now().Unix(),
		Priority:  priorityFlag,
	}

	chatChannel := fmt.Sprintf("axle:chat:%s", cfg.TeamID)
	return utils.PublishMessage(ctx, cfg.RedisClient, chatChannel, msg)
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().BoolVarP(&priorityFlag, "priority", "p", false, "Send as priority message (triggers desktop notifications)")
}
