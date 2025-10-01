package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/parzi-val/axle-file-sync/utils"
)

// teamCmd represents the team command
var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Display team member status and presence information",
	Long: utils.RenderTitle("👥 Team Status") + `

Shows a table with all team members including their online/offline status,
last seen time, IP addresses, and node IDs. This helps you see who's
currently active on your team and available for collaboration.

The status is updated in real-time based on heartbeat messages sent
every 30 seconds by each team member's Axle instance.`,
	
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := loadConfig(); err != nil {
			return fmt.Errorf("configuration error: %w. Please run 'axle init' first", err)
		}
		
		defer config.RedisClient.Close()

		ctx := context.Background()
		
		// Get team presence information
		presenceList, err := utils.GetTeamPresence(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to retrieve team presence: %w", err)
		}

		// Display the team status table
		fmt.Println(utils.RenderTitle("👥 Team: " + config.TeamID))
		fmt.Println(utils.RenderPresenceTable(presenceList))
		
		// Show summary
		onlineCount := 0
		for _, presence := range presenceList {
			if presence.Status == "online" {
				onlineCount++
			}
		}
		
		totalCount := len(presenceList)
		summaryMsg := fmt.Sprintf("Total: %d members, %d online, %d offline", 
			totalCount, onlineCount, totalCount-onlineCount)
		fmt.Println(utils.RenderInfo(summaryMsg))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
}
