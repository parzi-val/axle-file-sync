#!/bin/bash
# Axle File Sync Uninstaller for macOS/Linux
# Run with: curl -sSL https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/scripts/uninstall.sh | bash

set -e

INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="axle"

echo -e "\033[33mUninstalling Axle File Sync...\033[0m"

# Check if Axle is installed
if [ ! -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "\033[31mAxle is not installed at $INSTALL_DIR/$BINARY_NAME\033[0m"
    exit 1
fi

# Remove binary
echo -e "\033[33mRemoving Axle binary...\033[0m"
rm -f "$INSTALL_DIR/$BINARY_NAME"
echo -e "\033[32mBinary removed\033[0m"

# Remove from PATH in shell config files
echo -e "\033[33mCleaning up PATH entries...\033[0m"

# Clean bashrc
if [ -f ~/.bashrc ]; then
    grep -v "export PATH=\"\$PATH:$INSTALL_DIR\"" ~/.bashrc > ~/.bashrc.tmp 2>/dev/null || true
    if [ -s ~/.bashrc.tmp ]; then
        mv ~/.bashrc.tmp ~/.bashrc
        echo "Cleaned ~/.bashrc"
    else
        rm -f ~/.bashrc.tmp
    fi
fi

# Clean zshrc
if [ -f ~/.zshrc ]; then
    grep -v "export PATH=\"\$PATH:$INSTALL_DIR\"" ~/.zshrc > ~/.zshrc.tmp 2>/dev/null || true
    if [ -s ~/.zshrc.tmp ]; then
        mv ~/.zshrc.tmp ~/.zshrc
        echo "Cleaned ~/.zshrc"
    else
        rm -f ~/.zshrc.tmp
    fi
fi

# macOS specific: Remove from LaunchAgents if exists
if [[ "$OSTYPE" == "darwin"* ]]; then
    PLIST_FILE="$HOME/Library/LaunchAgents/com.axle.filesync.plist"
    if [ -f "$PLIST_FILE" ]; then
        echo -e "\033[33mRemoving LaunchAgent...\033[0m"
        launchctl unload "$PLIST_FILE" 2>/dev/null || true
        rm -f "$PLIST_FILE"
        echo -e "\033[32mLaunchAgent removed\033[0m"
    fi
fi

# Linux specific: Remove from applications menu
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    DESKTOP_FILE="$HOME/.local/share/applications/axle.desktop"
    if [ -f "$DESKTOP_FILE" ]; then
        echo -e "\033[33mRemoving desktop entry...\033[0m"
        rm -f "$DESKTOP_FILE"
        echo -e "\033[32mDesktop entry removed\033[0m"
    fi
fi

echo -e "\n\033[32mAxle File Sync has been uninstalled successfully!\033[0m"
echo -e "\033[36mYou may need to restart your terminal or run 'source ~/.bashrc' or 'source ~/.zshrc'\033[0m"