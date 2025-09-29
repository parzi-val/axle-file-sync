package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConflictStrategy defines how to handle conflicts when applying patches
type ConflictStrategy string

const (
	ConflictStrategyTheirs     ConflictStrategy = "theirs"     // Accept incoming changes
	ConflictStrategyMine       ConflictStrategy = "mine"       // Keep local changes
	ConflictStrategyMerge      ConflictStrategy = "merge"      // Create merge conflict markers
	ConflictStrategyBackup     ConflictStrategy = "backup"     // Create .backup files
	ConflictStrategyInteractive ConflictStrategy = "interactive" // Open in IDE for resolution
)

// ApplyPatchWithStrategy applies a patch with a specified conflict resolution strategy
func ApplyPatchWithStrategy(directory, patch string, strategy ConflictStrategy) (bool, error) {
	// Validate the patch for security issues
	if err := validatePatch(patch); err != nil {
		return false, fmt.Errorf("patch validation failed: %w", err)
	}

	// Clean up any previous git am/rebase state
	cleanupGitState(directory)

	isFormatPatch := strings.Contains(patch, "From ") && strings.Contains(patch, "Subject:")

	switch strategy {
	case ConflictStrategyTheirs:
		return applyPatchTheirs(directory, patch, isFormatPatch)
	case ConflictStrategyMine:
		return applyPatchMine(directory, patch, isFormatPatch)
	case ConflictStrategyMerge:
		return applyPatchMerge(directory, patch, isFormatPatch)
	case ConflictStrategyBackup:
		return applyPatchBackup(directory, patch, isFormatPatch)
	case ConflictStrategyInteractive:
		return applyPatchInteractive(directory, patch, isFormatPatch)
	default:
		// Fallback to default behavior
		return ApplyPatch(directory, patch)
	}
}

// applyPatchTheirs accepts all incoming changes, discarding local changes
func applyPatchTheirs(directory, patch string, isFormatPatch bool) (bool, error) {
	// Create a stash to save current work
	stashCmd := exec.Command("git", "-C", directory, "stash", "push", "-m", "Axle: Saving local changes before accepting incoming patch")
	stashCmd.Run()

	// Reset to clean state
	resetCmd := exec.Command("git", "-C", directory, "reset", "--hard", "HEAD")
	if err := resetCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to reset repository: %w", err)
	}

	// Apply the patch normally
	autoCommitted, err := ApplyPatch(directory, patch)
	if err == nil {
		log.Printf("[CONFLICT] Applied patch using 'theirs' strategy - local changes were stashed")
	}
	return autoCommitted, err
}

// applyPatchMine keeps local changes, ignoring the incoming patch
func applyPatchMine(directory, patch string, isFormatPatch bool) (bool, error) {
	log.Printf("[CONFLICT] Skipping patch using 'mine' strategy - keeping local changes")
	return false, nil // Don't apply the patch at all
}

// applyPatchMerge attempts to merge and creates conflict markers
func applyPatchMerge(directory, patch string, isFormatPatch bool) (bool, error) {
	if isFormatPatch {
		// Use git am with 3way merge to create conflict markers
		cmd := exec.Command("git", "-C", directory, "am", "--3way", "--no-commit")
		cmd.Stdin = strings.NewReader(patch)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		if err != nil {
			// Check if we have conflicts
			statusCmd := exec.Command("git", "-C", directory, "status", "--porcelain")
			statusOut, _ := statusCmd.Output()

			if strings.Contains(string(statusOut), "UU") || strings.Contains(out.String(), "Applying") {
				// We have merge conflicts - this is expected
				log.Printf("[CONFLICT] Merge conflicts detected - conflict markers added to files")

				// Add conflicted files to index
				addCmd := exec.Command("git", "-C", directory, "add", "-A")
				addCmd.Run()

				// List conflicted files for the user
				conflictedFiles := findConflictedFiles(directory)
				if len(conflictedFiles) > 0 {
					log.Printf("[CONFLICT] Files with conflicts: %v", conflictedFiles)
					log.Printf("[CONFLICT] Open these files in your IDE to resolve conflicts")

					// Optionally open in VS Code if available
					openInIDE(directory, conflictedFiles)
				}

				return false, nil // Don't auto-commit when there are conflicts
			}
			return false, fmt.Errorf("failed to apply patch with merge: %s", out.String())
		}
		return true, nil
	} else {
		// For diff patches, try 3way merge
		cmd := exec.Command("git", "-C", directory, "apply", "--3way", "--no-index")
		cmd.Stdin = strings.NewReader(patch)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			// Fall back to creating .rej files for conflicts
			cmd2 := exec.Command("git", "-C", directory, "apply", "--reject", "-")
			cmd2.Stdin = strings.NewReader(patch)
			var out2 bytes.Buffer
			cmd2.Stdout = &out2
			cmd2.Stderr = &out2

			if err2 := cmd2.Run(); err2 == nil || strings.Contains(out2.String(), "Applied") {
				// Some parts applied, some rejected
				rejFiles := findRejectedFiles(directory)
				if len(rejFiles) > 0 {
					log.Printf("[CONFLICT] Partial application - rejected hunks saved in: %v", rejFiles)
					openInIDE(directory, rejFiles)
				}
				return false, nil
			}

			return false, fmt.Errorf("failed to apply patch with merge: %s", out.String())
		}
		return false, nil
	}
}

// applyPatchBackup creates backup files before applying changes
func applyPatchBackup(directory, patch string, isFormatPatch bool) (bool, error) {
	// Find files that will be affected by the patch
	affectedFiles := extractFilesFromPatch(patch)

	// Create backups
	backupFiles := []string{}
	for _, file := range affectedFiles {
		fullPath := filepath.Join(directory, file)
		if _, err := os.Stat(fullPath); err == nil {
			backupPath := fullPath + ".backup"
			if err := copyFile(fullPath, backupPath); err == nil {
				backupFiles = append(backupFiles, backupPath)
			}
		}
	}

	if len(backupFiles) > 0 {
		log.Printf("[CONFLICT] Created backup files: %v", backupFiles)
	}

	// Apply the patch normally
	autoCommitted, err := ApplyPatch(directory, patch)
	if err != nil && len(backupFiles) > 0 {
		log.Printf("[CONFLICT] Patch failed - your original files are saved as .backup")
	}
	return autoCommitted, err
}

// applyPatchInteractive opens conflicts in the IDE for manual resolution
func applyPatchInteractive(directory, patch string, isFormatPatch bool) (bool, error) {
	// First try to apply with merge strategy to create conflict markers
	autoCommitted, _ := applyPatchMerge(directory, patch, isFormatPatch)

	// Find all conflicted files
	conflictedFiles := findConflictedFiles(directory)
	if len(conflictedFiles) == 0 {
		// No conflicts, patch applied cleanly
		return autoCommitted, nil
	}

	// Open in IDE and wait for user resolution
	log.Printf("[CONFLICT] Opening %d conflicted files in your IDE", len(conflictedFiles))
	openInIDE(directory, conflictedFiles)

	// Show instructions
	fmt.Println("\n" + RenderWarning("⚠️  Merge Conflicts Detected"))
	fmt.Println("\nConflicted files have been opened in your IDE with conflict markers:")
	for _, file := range conflictedFiles {
		fmt.Printf("  • %s\n", file)
	}
	fmt.Println("\n" + RenderInfo("How to resolve:"))
	fmt.Println("  1. Look for <<<<<<< HEAD, =======, and >>>>>>> markers")
	fmt.Println("  2. Choose which changes to keep (yours, theirs, or both)")
	fmt.Println("  3. Remove the conflict markers")
	fmt.Println("  4. Save the files")
	fmt.Println("  5. Run 'git add .' and commit when ready")
	fmt.Println("\nYour IDE should highlight these conflicts and provide merge tools.")

	return false, nil // Don't auto-commit in interactive mode
}

// cleanupGitState cleans up any git am/rebase in progress
func cleanupGitState(directory string) {
	exec.Command("git", "-C", directory, "am", "--abort").Run()
	exec.Command("git", "-C", directory, "rebase", "--abort").Run()
	exec.Command("git", "-C", directory, "merge", "--abort").Run()
}

// findConflictedFiles finds files with merge conflicts
func findConflictedFiles(directory string) []string {
	cmd := exec.Command("git", "-C", directory, "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	files := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if file := strings.TrimSpace(line); file != "" {
			files = append(files, file)
		}
	}
	return files
}

// findRejectedFiles finds .rej files created by patch --reject
func findRejectedFiles(directory string) []string {
	rejFiles := []string{}
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.HasSuffix(path, ".rej") {
			relPath, _ := filepath.Rel(directory, path)
			rejFiles = append(rejFiles, relPath)
		}
		return nil
	})
	return rejFiles
}

// extractFilesFromPatch parses a patch to find affected files
func extractFilesFromPatch(patch string) []string {
	files := []string{}
	lines := strings.Split(patch, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				file := parts[1]
				// Remove git's a/ b/ prefixes
				file = strings.TrimPrefix(file, "a/")
				file = strings.TrimPrefix(file, "b/")
				if file != "/dev/null" && !contains(files, file) {
					files = append(files, file)
				}
			}
		} else if strings.HasPrefix(line, "diff --git") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "a/") || strings.HasPrefix(part, "b/") {
					file := strings.TrimPrefix(part, "a/")
					file = strings.TrimPrefix(file, "b/")
					if !contains(files, file) {
						files = append(files, file)
					}
				}
			}
		}
	}
	return files
}

// openInIDE attempts to open files in the user's IDE
func openInIDE(directory string, files []string) {
	if len(files) == 0 {
		return
	}

	// Try VS Code first (most common)
	if tryOpenVSCode(directory, files) {
		return
	}

	// Try other IDEs
	// Could add IntelliJ, Sublime, Atom, etc.
}

// tryOpenVSCode attempts to open files in VS Code
func tryOpenVSCode(directory string, files []string) bool {
	// Check if code command exists
	checkCmd := exec.Command("code", "--version")
	if err := checkCmd.Run(); err != nil {
		return false
	}

	// Build command to open files
	args := []string{}
	for _, file := range files {
		args = append(args, filepath.Join(directory, file))
	}

	// Open VS Code with the files
	cmd := exec.Command("code", args...)
	if err := cmd.Start(); err == nil {
		log.Printf("[IDE] Opened %d files in VS Code", len(files))
		return true
	}

	return false
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}