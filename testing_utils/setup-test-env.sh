#!/bin/bash

# Script to automatically setup test environment for Axle
# Clears alice_project and bob_project folders and initializes them with Axle

set -e  # Exit on any error

echo "Setting up Axle test environment..."

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Define project directories
ALICE_DIR="alice_project"
BOB_DIR="bob_project"
AXLE_EXE="./axle.exe"

# Check if axle.exe exists
if [ ! -f "$AXLE_EXE" ]; then
    echo "Error: axle.exe not found in tests directory"
    echo "Please copy axle.exe to the tests folder first"
    exit 1
fi

echo "Cleaning up existing directories..."

# Function to forcefully remove directory
remove_dir_force() {
    local dir="$1"
    if [ -d "$dir" ]; then
        echo "Removing existing $dir directory..."
        # Try normal removal first
        if ! rm -rf "$dir" 2>/dev/null; then
            echo "Warning: Normal removal failed. Directory may be in use."
            echo "Please ensure no Axle processes are running in $dir"
            echo "You may need to:"
            echo "  1. Stop any running axle.exe processes"
            echo "  2. Close any terminals/editors in $dir"
            echo "  3. Wait a moment and try again"
            echo ""
            read -p "Press Enter to continue trying, or Ctrl+C to exit..."
            
            # Try again with more aggressive approach
            if command -v taskkill >/dev/null 2>&1; then
                echo "Attempting to stop any axle processes..."
                taskkill //F //IM axle.exe 2>/dev/null || true
                sleep 2
            fi
            
            # Final attempt
            if ! rm -rf "$dir" 2>/dev/null; then
                echo "Error: Could not remove $dir. Please manually delete it and run this script again."
                exit 1
            fi
        fi
    fi
}

# Remove existing directories if they exist
remove_dir_force "$ALICE_DIR"
remove_dir_force "$BOB_DIR"

# Create fresh directories
echo "Creating fresh directories..."
mkdir -p "$ALICE_DIR"
mkdir -p "$BOB_DIR"

# Initialize Alice's project
echo "Initializing Alice's project..."
cd "$ALICE_DIR"
../"$AXLE_EXE" init --team "dev-team" --username "alice"

echo "Alice's project initialized successfully!"

# Initialize Bob's project
echo "Initializing Bob's project..."
cd "../$BOB_DIR"
../"$AXLE_EXE" init --team "dev-team" --username "bob"

echo "Bob's project initialized successfully!"

# Go back to tests directory
cd "$SCRIPT_DIR"

echo ""
echo "✅ Test environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Copy axle.exe to this tests directory if not already done"
echo "2. Start Alice: cd alice_project && ../axle.exe start"
echo "3. Start Bob: cd bob_project && ../axle.exe start"
echo ""
echo "Directory structure:"
echo "├── tests/"
echo "│   ├── axle.exe"
echo "│   ├── setup-test-env.sh (this script)"
echo "│   ├── alice_project/"
echo "│   │   ├── .git/"
echo "│   │   └── axle_config.json"
echo "│   └── bob_project/"
echo "│       ├── .git/"
echo "│       └── axle_config.json"
echo ""
