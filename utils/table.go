package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Table styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	onlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")). // Green
			Bold(true)

	offlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)

	tableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

// RenderPresenceTable creates a beautiful table showing team member presence
func RenderPresenceTable(presenceList []PresenceInfo) string {
	if len(presenceList) == 0 {
		noDataStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginTop(1).
			MarginBottom(1)

		return noDataStyle.Render("No team members found. Make sure other team members are online and connected to the same Redis instance.")
	}

	// Define headers
	headers := []string{"Username", "Status", "Last Seen", "IP Address", "Node ID"}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Check data for wider columns
	rows := make([][]string, 0, len(presenceList))
	for _, info := range presenceList {
		statusStr := formatStatus(info.Status)
		lastSeenStr := formatLastSeen(info.LastSeen)

		row := []string{
			info.Username,
			statusStr,
			lastSeenStr,
			info.IPAddress,
			truncateNodeID(info.NodeID),
		}
		rows = append(rows, row)

		// Update column widths
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Build table
	var table strings.Builder

	// Header row
	headerRow := make([]string, len(headers))
	for i, header := range headers {
		headerRow[i] = headerStyle.Width(colWidths[i]).Render(header)
	}
	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerRow...))
	table.WriteString("\n")

	// Data rows
	for _, row := range rows {
		formattedRow := make([]string, len(row))
		for i, cell := range row {
			style := cellStyle.Width(colWidths[i])

			// Apply special styling for status column
			if i == 1 { // Status column
				if strings.Contains(cell, "Online") {
					style = style.Foreground(lipgloss.Color("82")) // Green
				} else {
					style = style.Foreground(lipgloss.Color("196")) // Red
				}
			}

			formattedRow[i] = style.Render(cell)
		}
		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, formattedRow...))
		table.WriteString("\n")
	}

	// Wrap in border
	return tableStyle.Render(table.String())
}

// formatStatus formats the status with appropriate indicators
func formatStatus(status string) string {
	if status == "online" {
		return "🟢 Online"
	}
	return "🔴 Offline"
}

// formatLastSeen formats a Unix timestamp to a human-readable "time ago" format
func formatLastSeen(timestamp int64) string {
	now := time.Now().Unix()
	diff := now - timestamp

	if diff < 5 {
		return "just now"
	} else if diff < 60 {
		return fmt.Sprintf("%ds ago", diff)
	} else if diff < 3600 {
		return fmt.Sprintf("%dm ago", diff/60)
	} else if diff < 86400 {
		return fmt.Sprintf("%dh ago", diff/3600)
	} else {
		return fmt.Sprintf("%dd ago", diff/86400)
	}
}

// truncateNodeID shortens the node ID for display
func truncateNodeID(nodeID string) string {
	if len(nodeID) > 12 {
		return nodeID[:12] + "..."
	}
	return nodeID
}

// RenderTitle renders a styled title for the CLI
func RenderTitle(title string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")). // Cyan
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	return titleStyle.Render(title)
}

// RenderSuccess renders a success message
func RenderSuccess(message string) string {
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")). // Green
		Bold(true)

	return successStyle.Render("✅ " + message)
}

// RenderError renders an error message
func RenderError(message string) string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red
		Bold(true)

	return errorStyle.Render("❌ " + message)
}

// RenderWarning renders a warning message
func RenderWarning(message string) string {
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Orange
		Bold(true)

	return warningStyle.Render("⚠️  " + message)
}

// RenderInfo renders an info message
func RenderInfo(message string) string {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")). // Blue
		Bold(true)

	return infoStyle.Render("ℹ️  " + message)
}
