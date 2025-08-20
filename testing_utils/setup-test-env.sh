#!/bin/bash

# Script to automatically setup test environment for Axle
# Clears alice_project and bob_project folders

set -e # Exit on any error

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

# Remove existing directories if they exist
if [ -d "$ALICE_DIR" ]; then
    echo "Removing existing $ALICE_DIR directory..."
    rm -rf "$ALICE_DIR"
fi

if [ -d "$BOB_DIR" ]; then
    echo "Removing existing $BOB_DIR directory..."
    rm -rf "$BOB_DIR"
fi

# Create fresh directories
echo "Creating fresh directories..."
mkdir -p "$ALICE_DIR"
mkdir -p "$BOB_DIR"

# Go back to tests directory
cd "$SCRIPT_DIR"

echo ""
echo "✅ Test environment cleaned up!"
echo ""
echo "Next steps:"
echo "1. Team Lead (Alice) initializes the team:"
echo "   cd testing_utils/alice_project"
echo "   ../axle.exe init --team \"dev-team\" --username \"alice\""
echo ""
echo "2. Team Member (Bob) joins the team:"
echo "   cd testing_utils/bob_project"
echo "   ../axle.exe join --team \"dev-team\" --username \"bob\""
echo ""
echo "3. Start both clients:"
echo "   - In one terminal, from alice_project: ../axle.exe start"
echo "   - In another terminal, from bob_project: ../axle.exe start"
echo ""
