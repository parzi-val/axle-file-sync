package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// SendNotification sends a desktop notification
func SendNotification(title, message string) error {
	switch runtime.GOOS {
	case "windows":
		return sendWindowsNotification(title, message)
	case "darwin":
		return sendMacNotification(title, message)
	case "linux":
		return sendLinuxNotification(title, message)
	default:
		// Silently fail on unsupported platforms
		return nil
	}
}

func sendWindowsNotification(title, message string) error {
	// Method 1: PowerShell balloon notification (works on all Windows)
	script := fmt.Sprintf(`
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$notification = New-Object System.Windows.Forms.NotifyIcon
$notification.Icon = [System.Drawing.SystemIcons]::Information
$notification.BalloonTipIcon = 'Info'
$notification.BalloonTipTitle = '%s'
$notification.BalloonTipText = '%s'
$notification.Visible = $true
$notification.ShowBalloonTip(5000)
Start-Sleep -Milliseconds 5000
$notification.Dispose()
`, title, escapeForPowerShell(message))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	err := cmd.Run()

	if err != nil {
		// Fallback: Try msg command (simpler but less pretty)
		msgCmd := exec.Command("msg", "*", fmt.Sprintf("%s: %s", title, message))
		return msgCmd.Run()
	}

	return nil
}

func sendMacNotification(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "default"`,
		escapeForAppleScript(message), escapeForAppleScript(title))

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func sendLinuxNotification(title, message string) error {
	// First try notify-send (most common)
	cmd := exec.Command("notify-send", title, message, "-u", "normal", "-t", "5000")
	err := cmd.Run()

	if err != nil {
		// Fallback: Try zenity
		zenityCmd := exec.Command("zenity", "--notification", "--text", fmt.Sprintf("%s: %s", title, message))
		return zenityCmd.Run()
	}

	return nil
}

// Helper functions to escape special characters
func escapeForPowerShell(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, "$", "`$")
	return s
}

func escapeForAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// SendChatNotification sends a notification for incoming chat messages
func SendChatNotification(sender, message string) {
	title := fmt.Sprintf("Axle Chat - %s", sender)

	// Truncate message if too long
	if len(message) > 100 {
		message = message[:97] + "..."
	}

	// Send notification (ignore errors silently)
	_ = SendNotification(title, message)
}