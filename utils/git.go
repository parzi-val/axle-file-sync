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

// validatePatch checks if a patch contains any dangerous path traversal attempts
func validatePatch(patch string) error {
	// Check for path traversal patterns in the patch
	dangerousPatterns := []string{
		"../",           // Parent directory traversal
		"..\\",          // Windows parent directory traversal
		"/etc/",         // System configuration files
		"\\etc\\",       // Windows system files
		"/root/",        // Root user directory
		"C:\\Windows\\", // Windows system directory
		"/usr/bin/",     // System binaries
		"~/.ssh/",       // SSH keys
		"~/.aws/",       // AWS credentials
	}

	lines := strings.Split(patch, "\n")
	for _, line := range lines {
		// Check file paths in diff headers (+++, ---, diff --git)
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "diff --git") || strings.Contains(line, "rename to") ||
			strings.Contains(line, "rename from") {
			for _, pattern := range dangerousPatterns {
				if strings.Contains(line, pattern) {
					return fmt.Errorf("patch contains dangerous path pattern: %s", pattern)
				}
			}

			// Check for absolute paths (security risk)
			if strings.Contains(line, " /") && !strings.Contains(line, " a/") && !strings.Contains(line, " b/") {
				// Allow git's a/ and b/ prefixes but block other absolute paths
				if !strings.Contains(line, "/dev/null") { // Allow /dev/null for deletions
					return fmt.Errorf("patch contains absolute path which is not allowed")
				}
			}
		}
	}

	// Check patch size (prevent DOS with huge patches)
	maxPatchSize := 10 * 1024 * 1024 // 10MB limit
	if len(patch) > maxPatchSize {
		return fmt.Errorf("patch size exceeds maximum allowed size of 10MB")
	}

	return nil
}

// ApplyPatch applies a patch to the repository.
// Returns (autoCommitted bool, error) where autoCommitted indicates if the patch was auto-committed by git am.
func ApplyPatch(directory, patch string) (bool, error) {
	// Validate the patch for security issues
	if err := validatePatch(patch); err != nil {
		return false, fmt.Errorf("patch validation failed: %w", err)
	}

	// First clean up any previous git am/rebase state
	abortCmd := exec.Command("git", "-C", directory, "am", "--abort")
	abortCmd.Run() // Ignore errors - this is cleanup

	rebaseAbortCmd := exec.Command("git", "-C", directory, "rebase", "--abort")
	rebaseAbortCmd.Run() // Ignore errors - this is cleanup
	
	// Check if this is a format-patch style patch (has "From" header)
	isFormatPatch := strings.Contains(patch, "From ") && strings.Contains(patch, "Subject:")
	
	if isFormatPatch {
		// First try with --3way for repos with shared history
		cmd := exec.Command("git", "-C", directory, "am", "--whitespace=nowarn", "--ignore-whitespace", "--3way")
		cmd.Stdin = strings.NewReader(patch)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			// Check if it failed due to missing ancestor (independent repos)
			if strings.Contains(out.String(), "could not build fake ancestor") || strings.Contains(out.String(), "sha1 information is lacking") {
				// Extract the diff from the format-patch and apply it as a regular patch
				// This handles independent repositories without shared history

				// Extract the diff portion (everything after the first "---" line)
				diffStart := strings.Index(patch, "\n---")
				if diffStart == -1 {
					return false, fmt.Errorf("failed to extract diff from format-patch")
				}

				diffPatch := patch[diffStart:]

				// Apply as a regular diff without --3way
				applyCmd := exec.Command("git", "-C", directory, "apply", "--whitespace=nowarn", "--ignore-whitespace", "-")
				applyCmd.Stdin = strings.NewReader(diffPatch)
				var applyOut bytes.Buffer
				applyCmd.Stdout = &applyOut
				applyCmd.Stderr = &applyOut

				if applyErr := applyCmd.Run(); applyErr != nil {
					// If regular apply also fails, return the error
					return false, fmt.Errorf("failed to apply patch from independent repo: %s", applyOut.String())
				}

				// Successfully applied the diff, now commit with the original message
				// Extract commit message from the format-patch
				subjectMatch := strings.Index(patch, "Subject: ")
				if subjectMatch != -1 {
					messageStart := subjectMatch + 9
					messageEnd := strings.Index(patch[messageStart:], "\n")
					if messageEnd != -1 {
						commitMessage := patch[messageStart : messageStart+messageEnd]
						// Remove [PATCH] prefix if present
						commitMessage = strings.TrimPrefix(commitMessage, "[PATCH] ")

						// Stage all changes
						addCmd := exec.Command("git", "-C", directory, "add", ".")
						addCmd.Run()

						// Commit with the extracted message
						commitCmd := exec.Command("git", "-C", directory, "commit", "-m", commitMessage)
						commitCmd.Run()
					}
				}

				return true, nil
			} else if strings.Contains(out.String(), "would be overwritten") || strings.Contains(out.String(), "already exists") {
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
