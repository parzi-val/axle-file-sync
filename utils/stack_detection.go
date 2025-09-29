package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// StackType represents different technology stacks
type StackType string

const (
	StackNode   StackType = "node"
	StackPython StackType = "python"
	StackGo     StackType = "go"
	StackRust   StackType = "rust"
	StackJava   StackType = "java"
	StackDotNet StackType = "dotnet"
	StackRuby   StackType = "ruby"
	StackPHP    StackType = "php"
	StackUnknown StackType = "unknown"
)

// DetectStack analyzes the project directory to identify the technology stack
func DetectStack(rootDir string) []StackType {
	detectedStacks := []StackType{}

	// Check for various stack indicators
	if fileExists(filepath.Join(rootDir, "package.json")) {
		detectedStacks = append(detectedStacks, StackNode)
	}
	if fileExists(filepath.Join(rootDir, "requirements.txt")) ||
	   fileExists(filepath.Join(rootDir, "setup.py")) ||
	   fileExists(filepath.Join(rootDir, "Pipfile")) ||
	   fileExists(filepath.Join(rootDir, "pyproject.toml")) {
		detectedStacks = append(detectedStacks, StackPython)
	}
	if fileExists(filepath.Join(rootDir, "go.mod")) {
		detectedStacks = append(detectedStacks, StackGo)
	}
	if fileExists(filepath.Join(rootDir, "Cargo.toml")) {
		detectedStacks = append(detectedStacks, StackRust)
	}
	if fileExists(filepath.Join(rootDir, "pom.xml")) ||
	   fileExists(filepath.Join(rootDir, "build.gradle")) {
		detectedStacks = append(detectedStacks, StackJava)
	}
	if fileExists(filepath.Join(rootDir, "*.csproj")) ||
	   fileExists(filepath.Join(rootDir, "*.sln")) {
		detectedStacks = append(detectedStacks, StackDotNet)
	}
	if fileExists(filepath.Join(rootDir, "Gemfile")) {
		detectedStacks = append(detectedStacks, StackRuby)
	}
	if fileExists(filepath.Join(rootDir, "composer.json")) {
		detectedStacks = append(detectedStacks, StackPHP)
	}

	if len(detectedStacks) == 0 {
		detectedStacks = append(detectedStacks, StackUnknown)
	}

	return detectedStacks
}

// GetIgnorePatternsForStack returns common ignore patterns for a given stack
func GetIgnorePatternsForStack(stack StackType) []string {
	basePatterns := []string{
		// Version control
		".git",
		".svn",
		".hg",

		// OS files
		".DS_Store",
		"Thumbs.db",
		"desktop.ini",

		// Editor files
		"*.swp",
		"*.swo",
		"*~",
		".idea",
		".vscode",
		"*.sublime-*",

		// Logs and temp files
		"*.log",
		"*.tmp",
		"*.temp",
		"*.cache",
	}

	stackSpecific := map[StackType][]string{
		StackNode: {
			"node_modules",
			"dist",
			"build",
			".next",
			".nuxt",
			"coverage",
			"*.tsbuildinfo",
			".npm",
			".yarn",
			".pnp.*",
			".env.local",
			".env.*.local",
		},
		StackPython: {
			"__pycache__",
			"*.py[cod]",
			"*$py.class",
			"*.so",
			".Python",
			"env",
			"venv",
			"ENV",
			".venv",
			"pip-log.txt",
			"*.egg-info",
			".pytest_cache",
			".mypy_cache",
			".coverage",
			"htmlcov",
		},
		StackGo: {
			"*.exe",
			"*.exe~",
			"*.dll",
			"*.so",
			"*.dylib",
			"*.test",
			"*.out",
			"vendor",
			"go.sum",
		},
		StackRust: {
			"target",
			"Cargo.lock",
			"**/*.rs.bk",
		},
		StackJava: {
			"target",
			"*.class",
			"*.jar",
			"*.war",
			"*.ear",
			".gradle",
			"build",
			"out",
		},
		StackDotNet: {
			"bin",
			"obj",
			"*.user",
			"*.suo",
			".vs",
			"packages",
		},
		StackRuby: {
			"*.gem",
			"*.rbc",
			".bundle",
			".config",
			"coverage",
			"InstalledFiles",
			"pkg",
			"spec/reports",
			"test/tmp",
			"test/version_tmp",
			"tmp",
		},
		StackPHP: {
			"vendor",
			"composer.lock",
			".phpunit.result.cache",
		},
	}

	patterns := append([]string{}, basePatterns...)
	if specific, ok := stackSpecific[stack]; ok {
		patterns = append(patterns, specific...)
	}

	return patterns
}

// AutoConfigureGitignore detects the project stack and returns appropriate ignore patterns
func AutoConfigureGitignore(rootDir string) []string {
	stacks := DetectStack(rootDir)

	// Log detected stacks
	stackNames := []string{}
	for _, stack := range stacks {
		if stack != StackUnknown {
			stackNames = append(stackNames, string(stack))
		}
	}

	if len(stackNames) > 0 {
		log.Printf("[INIT] Detected project stacks: %s", strings.Join(stackNames, ", "))
	}

	// Combine patterns from all detected stacks
	patternSet := make(map[string]bool)
	for _, stack := range stacks {
		for _, pattern := range GetIgnorePatternsForStack(stack) {
			patternSet[pattern] = true
		}
	}

	// Convert set to slice
	patterns := []string{}
	for pattern := range patternSet {
		patterns = append(patterns, pattern)
	}

	return patterns
}

// WriteGitignore writes the ignore patterns to .gitignore file
func WriteGitignore(rootDir string, patterns []string) error {
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	// Check if .gitignore already exists
	existingPatterns := []string{}
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingPatterns = strings.Split(string(data), "\n")
	}

	// Merge with existing patterns
	patternSet := make(map[string]bool)
	for _, p := range existingPatterns {
		if p = strings.TrimSpace(p); p != "" && !strings.HasPrefix(p, "#") {
			patternSet[p] = true
		}
	}
	for _, p := range patterns {
		patternSet[p] = true
	}

	// Build the final content
	var content strings.Builder
	content.WriteString("# Axle auto-generated ignore patterns\n")
	content.WriteString("# Generated based on detected project stack\n\n")

	// Add axle-specific ignores first
	content.WriteString("# Axle configuration\n")
	content.WriteString("axle_config.json\n")
	content.WriteString(".axle\n\n")

	// Add stack-specific patterns
	for pattern := range patternSet {
		content.WriteString(pattern + "\n")
	}

	return os.WriteFile(gitignorePath, []byte(content.String()), 0644)
}

// fileExists checks if a file exists (supports wildcards)
func fileExists(path string) bool {
	// Check for wildcards
	if strings.Contains(path, "*") {
		dir := filepath.Dir(path)
		pattern := filepath.Base(path)

		entries, err := os.ReadDir(dir)
		if err != nil {
			return false
		}

		for _, entry := range entries {
			if matched, _ := filepath.Match(pattern, entry.Name()); matched {
				return true
			}
		}
		return false
	}

	// Direct file check
	_, err := os.Stat(path)
	return err == nil
}

// UpdateConfigWithIgnorePatterns adds detected ignore patterns to the local config
func UpdateConfigWithIgnorePatterns(configPath string, patterns []string) error {
	// Read existing config as a map to preserve all fields
	var configData map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &configData)
	} else {
		configData = make(map[string]interface{})
	}

	// Update ignore patterns
	configData["ignorePatterns"] = patterns

	// Write back
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, jsonData, 0644)
}