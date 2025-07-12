package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// InitGitRepo initializes a Git repository in the specified directory if not already initialized.
func InitGitRepo(directory string) error {
	cmd := exec.Command("git", "-C", directory, "init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Git repo: %w\n%s", err, string(output))
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

// CommitChanges stages all changes and commits them.
// It returns the new commit hash. If there are no changes to commit, it returns an empty string.
func CommitChanges(directory, message string) (string, error) {
	// Stage all changes
	addCmd := exec.Command("git", "-C", directory, "add", ".")
	var addErr bytes.Buffer
	addCmd.Stderr = &addErr
	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to stage changes (git add): %s", addErr.String())
	}

	// Commit the staged changes
	commitCmd := exec.Command("git", "-C", directory, "commit", "-m", message)
	var out bytes.Buffer
	var stderr bytes.Buffer
	commitCmd.Stdout = &out
	commitCmd.Stderr = &stderr

	if err := commitCmd.Run(); err != nil {
		// If commit fails because there's nothing to commit, it's not a fatal error.
		// We return an empty hash to signify that no new patch should be generated.
		if strings.Contains(stderr.String(), "nothing to commit") || strings.Contains(stderr.String(), "no changes added to commit") {
			return "", nil
		}
		return "", fmt.Errorf("failed to commit changes: %s", stderr.String())
	}

	// Get the commit hash of the new commit
	hashCmd := exec.Command("git", "-C", directory, "rev-parse", "HEAD")
	output, err := hashCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get new commit hash: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetPatch generates a patch for a given commit.
func GetPatch(directory, commitHash string) (string, error) {
	// Check if the commit has a parent. If not, it's the initial commit.
	parentCheckCmd := exec.Command("git", "-C", directory, "rev-parse", "--verify", commitHash+"^")
	isInitial := parentCheckCmd.Run() != nil

	var cmd *exec.Cmd
	if isInitial {
		// For the initial commit, create a patch from the root.
		cmd = exec.Command("git", "-C", directory, "show", commitHash)
	} else {
		// For subsequent commits, create a patch from the previous commit.
		cmd = exec.Command("git", "-C", directory, "format-patch", "--stdout", commitHash+"^.."+commitHash)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// The error message from git is useful for debugging.
		return "", fmt.Errorf("failed to generate patch: %w\n%s", err, string(output))
	}
	return string(output), nil
}


// ApplyPatch applies a patch to the repository.
func ApplyPatch(directory, patch string) error {
	cmd := exec.Command("git", "-C", directory, "apply", "-")
	cmd.Stdin = strings.NewReader(patch)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply patch: %s", out.String())
	}
	return nil
}