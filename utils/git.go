package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// InitGitRepo initializes a Git repository in the specified directory if not already initialized.
func InitGitRepo(directory string) error {
	cmd := exec.Command("git", "init", directory)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize Git repo: %w", err)
	}
	if err := exec.Command("git", "add", "-N", ".").Run(); err != nil {
		return fmt.Errorf("failed to add files to Git tracking: %w", err)
	}
	return nil
}

// GetGitDiff retrieves the Git diff for a given file.
func GetGitDiff(filePath string) (string, error) {
	cmd := exec.Command("git", "diff", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get Git diff: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
