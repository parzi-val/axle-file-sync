package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version   = "0.1.1"
	BuildDate = "2024-10-01"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display Axle version information",
	Long:  "Display the current version of Axle file sync tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Axle File Sync v%s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Real-time file synchronization for hackathon teams\n")
		fmt.Printf("https://github.com/parzi-val/axle-file-sync\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}