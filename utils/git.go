package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)


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
		stdErrStr := stderr.String()
		stdOutStr := out.String()
		
		// If commit fails because there's nothing to commit, it's not a fatal error.
		// We return an empty hash to signify that no new patch should be generated.
		// Check both stderr and stdout for the "nothing to commit" message
		if strings.Contains(stdErrStr, "nothing to commit") || 
		   strings.Contains(stdErrStr, "no changes added to commit") ||
		   strings.Contains(stdOutStr, "nothing to commit") ||
		   strings.Contains(stdOutStr, "working tree clean") {
			return "", nil // Not an error - just nothing to commit
		}
		
		// Provide detailed error information
		errorDetails := fmt.Sprintf("Git commit failed - Exit Code: %v", err)
		if stdErrStr != "" {
			errorDetails += fmt.Sprintf("\nStderr: %s", stdErrStr)
		}
		if stdOutStr != "" {
			errorDetails += fmt.Sprintf("\nStdout: %s", stdOutStr)
		}
		
		// Check git status to provide more context
		statusCmd := exec.Command("git", "-C", directory, "status", "--porcelain")
		if statusOutput, statusErr := statusCmd.CombinedOutput(); statusErr == nil {
			errorDetails += fmt.Sprintf("\nGit Status: %s", string(statusOutput))
		} else {
			errorDetails += fmt.Sprintf("\nFailed to get git status: %v", statusErr)
		}
		
		return "", fmt.Errorf("failed to commit changes: %s", errorDetails)
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


// InitGitRepo initializes a Git repository in the specified directory
func InitGitRepo(directory string) error {
	cmd := exec.Command("git", "-C", directory, "init")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %s", stderr.String())
	}

	// Check if this is a fresh repo and add an initial commit
	statusCmd := exec.Command("git", "-C", directory, "status", "--porcelain")
	_, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %v", err)
	}

	// If there are no commits yet, create an initial empty commit
	logCmd := exec.Command("git", "-C", directory, "log", "--oneline", "-n", "1")
	if err := logCmd.Run(); err != nil {
		// No commits exist, create initial commit
		initialCommitCmd := exec.Command("git", "-C", directory, "commit", "--allow-empty", "-m", "Initial commit")
		var commitStderr bytes.Buffer
		initialCommitCmd.Stderr = &commitStderr

		if err := initialCommitCmd.Run(); err != nil {
			return fmt.Errorf("failed to create initial commit: %s", commitStderr.String())
		}
	}

	return nil
}

// ApplyPatch applies a patch to the repository.
// Returns (autoCommitted bool, error) where autoCommitted indicates if the patch was auto-committed by git am.
func ApplyPatch(directory, patch string) (bool, error) {
	// First clean up any previous git am/rebase state
	abortCmd := exec.Command("git", "-C", directory, "am", "--abort")
	abortCmd.Run() // Ignore errors - this is cleanup
	
	rebaseAbortCmd := exec.Command("git", "-C", directory, "rebase", "--abort")
	rebaseAbortCmd.Run() // Ignore errors - this is cleanup
	
	// Check if this is a format-patch style patch (has "From" header)
	isFormatPatch := strings.Contains(patch, "From ") && strings.Contains(patch, "Subject:")
	
	if isFormatPatch {
		// For format-patch style patches, use git am with ignore flags
		cmd := exec.Command("git", "-C", directory, "am", "--whitespace=nowarn", "--ignore-whitespace", "--3way")
		cmd.Stdin = strings.NewReader(patch)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		
		if err := cmd.Run(); err != nil {
			// If git am fails due to existing files, try to reset and apply again
			if strings.Contains(out.String(), "would be overwritten") || strings.Contains(out.String(), "already exists") {
				// Reset to clean state and try again
				resetCmd := exec.Command("git", "-C", directory, "reset", "--hard", "HEAD")
				resetCmd.Run()
				
				cleanCmd := exec.Command("git", "-C", directory, "clean", "-fd")
				cleanCmd.Run()
				
				// Try git am again
				cmd2 := exec.Command("git", "-C", directory, "am", "--whitespace=nowarn", "--ignore-whitespace", "--3way")
				cmd2.Stdin = strings.NewReader(patch)
				var out2 bytes.Buffer
				cmd2.Stdout = &out2
				cmd2.Stderr = &out2
				
				if err2 := cmd2.Run(); err2 != nil {
					return true, fmt.Errorf("failed to apply format-patch after reset:\nFirst attempt: %s\nSecond attempt: %s", out.String(), out2.String())
				}
			} else {
				return true, fmt.Errorf("failed to apply format-patch: %s", out.String())
			}
		}
		// git am was successful, return true for autoCommitted since git am commits automatically
		return true, nil
	} else {
		// For regular diff patches, use git apply
		cmd := exec.Command("git", "-C", directory, "apply", "--whitespace=nowarn", "--index", "--reject", "-")
		cmd.Stdin = strings.NewReader(patch)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		
		if err := cmd.Run(); err != nil {
			// If that fails, try with --3way for better conflict resolution
			cmd2 := exec.Command("git", "-C", directory, "apply", "--whitespace=nowarn", "--3way", "-")
			cmd2.Stdin = strings.NewReader(patch)
			var out2 bytes.Buffer
			cmd2.Stdout = &out2
			cmd2.Stderr = &out2
			
			if err2 := cmd2.Run(); err2 != nil {
				return false, fmt.Errorf("failed to apply diff patch:\nMethod 1 (apply --whitespace=nowarn --index --reject): %s\nMethod 2 (apply --whitespace=nowarn --3way): %s", out.String(), out2.String())
			}
		}
		// git apply was successful, return false for autoCommitted since git apply doesn't commit
		return false, nil
	}
}
