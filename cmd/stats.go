package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/parzi-val/axle-file-sync/utils"
)

// SyncStats holds statistics about the sync session
type SyncStats struct {
	// Git stats
	TotalCommits   int
	LastCommitTime time.Time
	LastCommitMsg  string

	// File stats
	TotalFiles      int
	TrackedFiles    int
	IgnoredFiles    int
	LargestFile     string
	LargestFileSize int64

	// Sync stats
	TeamMembers     int
	OnlineMembers   int
	LastSyncTime    time.Time
	PendingChanges  int

	// Activity stats
	ChangesInLastHour int
	MostActiveFile    string
	MostActiveCount   int
}

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display comprehensive synchronization statistics",
	Long: utils.RenderTitle("📊 Sync Statistics") + `

Shows detailed statistics about your Axle synchronization session including:
• Git repository status and recent commits
• File statistics and ignored patterns
• Team member activity and presence
• Recent synchronization activity
• Current pending changes

Use this command to get a quick overview of your team's sync status
and identify any issues or bottlenecks.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		if err := loadConfig(); err != nil {
			return fmt.Errorf("configuration error: %w. Please run 'axle init' first", err)
		}
		defer config.RedisClient.Close()

		ctx := context.Background()

		// Gather statistics
		stats, err := gatherStats(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to gather statistics: %w", err)
		}

		// Display statistics
		displayStats(stats, config)

		return nil
	},
}

func gatherStats(ctx context.Context, cfg utils.AppConfig) (*SyncStats, error) {
	stats := &SyncStats{}

	// Get Git statistics
	if err := getGitStats(cfg.RootDir, stats); err != nil {
		return nil, fmt.Errorf("failed to get git stats: %w", err)
	}

	// Get file statistics
	if err := getFileStats(cfg.RootDir, cfg.IgnorePatterns, stats); err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Get team presence
	presenceList, err := utils.GetTeamPresence(ctx, cfg)
	if err == nil {
		stats.TeamMembers = len(presenceList)
		for _, presence := range presenceList {
			if presence.Status == "online" {
				stats.OnlineMembers++
			}
		}
	}

	// Get pending changes (git status)
	pendingCmd := exec.Command("git", "-C", cfg.RootDir, "status", "--porcelain")
	if output, err := pendingCmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				stats.PendingChanges++
			}
		}
	}

	return stats, nil
}

func getGitStats(rootDir string, stats *SyncStats) error {
	// Get total commit count
	countCmd := exec.Command("git", "-C", rootDir, "rev-list", "--count", "HEAD")
	if output, err := countCmd.Output(); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &stats.TotalCommits)
	}

	// Get last commit info
	lastCommitCmd := exec.Command("git", "-C", rootDir, "log", "-1", "--format=%ct|%s")
	if output, err := lastCommitCmd.Output(); err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(output)), "|", 2)
		if len(parts) == 2 {
			var timestamp int64
			fmt.Sscanf(parts[0], "%d", &timestamp)
			stats.LastCommitTime = time.Unix(timestamp, 0)
			stats.LastCommitMsg = parts[1]
		}
	}

	// Get commits in last hour
	hourAgo := time.Now().Add(-time.Hour).Unix()
	recentCmd := exec.Command("git", "-C", rootDir, "rev-list", "--count", "--since", fmt.Sprintf("%d", hourAgo), "HEAD")
	if output, err := recentCmd.Output(); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &stats.ChangesInLastHour)
	}

	// Get most frequently changed file
	freqCmd := exec.Command("git", "-C", rootDir, "log", "--pretty=format:", "--name-only")
	if output, err := freqCmd.Output(); err == nil {
		fileCount := make(map[string]int)
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if file := strings.TrimSpace(line); file != "" {
				fileCount[file]++
			}
		}

		for file, count := range fileCount {
			if count > stats.MostActiveCount {
				stats.MostActiveFile = file
				stats.MostActiveCount = count
			}
		}
	}

	return nil
}

func getFileStats(rootDir string, ignorePatterns []string, stats *SyncStats) error {
	var largestSize int64
	var largestFile string
	var totalFiles, trackedFiles, ignoredFiles int

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		totalFiles++
		relPath, _ := filepath.Rel(rootDir, path)

		// Check if ignored
		if isFileIgnored(relPath, ignorePatterns) {
			ignoredFiles++
		} else {
			trackedFiles++

			// Track largest file
			if info.Size() > largestSize {
				largestSize = info.Size()
				largestFile = relPath
			}
		}

		return nil
	})

	stats.TotalFiles = totalFiles
	stats.TrackedFiles = trackedFiles
	stats.IgnoredFiles = ignoredFiles
	stats.LargestFile = largestFile
	stats.LargestFileSize = largestSize

	return err
}

func isFileIgnored(path string, patterns []string) bool {
	// Check .git folder
	if strings.Contains(path, ".git"+string(filepath.Separator)) || strings.HasPrefix(path, ".git") {
		return true
	}

	fileName := filepath.Base(path)
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}
	return false
}

func displayStats(stats *SyncStats, cfg utils.AppConfig) {
	fmt.Println(utils.RenderTitle("📊 Axle Sync Statistics"))
	fmt.Printf("Team: %s | User: %s\n\n", cfg.TeamID, cfg.Username)

	// Repository Stats
	fmt.Println(utils.RenderInfo("📁 Repository"))
	fmt.Printf("  Total Commits:      %d\n", stats.TotalCommits)
	if stats.TotalCommits > 0 {
		fmt.Printf("  Last Commit:        %s\n", formatTime(stats.LastCommitTime))
		fmt.Printf("  Last Message:       %s\n", truncateString(stats.LastCommitMsg, 50))
	}
	fmt.Printf("  Pending Changes:    %d\n", stats.PendingChanges)
	fmt.Println()

	// File Stats
	fmt.Println(utils.RenderInfo("📄 Files"))
	fmt.Printf("  Total Files:        %d\n", stats.TotalFiles)
	fmt.Printf("  Tracked:            %d\n", stats.TrackedFiles)
	fmt.Printf("  Ignored:            %d\n", stats.IgnoredFiles)
	if stats.LargestFile != "" {
		fmt.Printf("  Largest File:       %s (%s)\n", stats.LargestFile, formatFileSize(stats.LargestFileSize))
	}
	fmt.Println()

	// Team Stats
	fmt.Println(utils.RenderInfo("👥 Team"))
	fmt.Printf("  Total Members:      %d\n", stats.TeamMembers)
	fmt.Printf("  Currently Online:   %d\n", stats.OnlineMembers)
	fmt.Printf("  Currently Offline:  %d\n", stats.TeamMembers-stats.OnlineMembers)
	fmt.Println()

	// Activity Stats
	fmt.Println(utils.RenderInfo("📈 Activity"))
	fmt.Printf("  Changes (Last Hour): %d\n", stats.ChangesInLastHour)
	if stats.MostActiveFile != "" {
		fmt.Printf("  Most Active File:    %s (%d changes)\n", stats.MostActiveFile, stats.MostActiveCount)
	}
	fmt.Println()

	// Status Summary
	if stats.PendingChanges > 0 {
		fmt.Println(utils.RenderWarning(fmt.Sprintf("You have %d uncommitted changes", stats.PendingChanges)))
	} else {
		fmt.Println(utils.RenderSuccess("Working directory is clean"))
	}

	if stats.OnlineMembers == 0 {
		fmt.Println(utils.RenderWarning("No team members are currently online"))
	} else if stats.OnlineMembers == 1 {
		fmt.Println(utils.RenderInfo("You are the only one online"))
	} else {
		fmt.Println(utils.RenderSuccess(fmt.Sprintf("%d team members are online", stats.OnlineMembers)))
	}
}

func formatTime(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	}
	return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(statsCmd)
}