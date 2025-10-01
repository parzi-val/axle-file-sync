#!/bin/bash
# Axle File Sync Installer for macOS/Linux
# Run with: curl -sSL https://raw.githubusercontent.com/parzi-val/axle-file-sync/main/scripts/install.sh | bash

set -e

VERSION="0.1.0"
REPO="parzi-val/axle-file-sync"
INSTALL_DIR="$HOME/.local/bin"

echo -e "\033[36mInstalling Axle File Sync...\033[0m"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64) ARCH="arm64" ;;
    aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

ARCHIVE_NAME="axle-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/v$VERSION/$ARCHIVE_NAME"
TEMP_DIR=$(mktemp -d)

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download and extract
echo -e "\033[33mDownloading Axle v$VERSION for $OS/$ARCH...\033[0m"
curl -sSL "$DOWNLOAD_URL" -o "$TEMP_DIR/$ARCHIVE_NAME"

echo -e "\033[33mExtracting...\033[0m"
tar -xzf "$TEMP_DIR/$ARCHIVE_NAME" -C "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/axle"

# Clean up
rm -rf "$TEMP_DIR"

echo -e "\033[32mInstalled to $INSTALL_DIR/axle\033[0m"

# Add to PATH if needed
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "\033[33mAdding to PATH...\033[0m"

    # Detect shell
    SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
        bash)
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.bashrc
            echo "Added to ~/.bashrc. Run 'source ~/.bashrc' or restart terminal."
            ;;
        zsh)
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.zshrc
            echo "Added to ~/.zshrc. Run 'source ~/.zshrc' or restart terminal."
            ;;
        *)
            echo "Add $INSTALL_DIR to your PATH manually"
            ;;
    esac
fi

echo -e "\n\033[32mAxle File Sync v$VERSION installed successfully!\033[0m"
echo -e "\033[36mRun 'axle version' to verify installation\033[0m"
echo -e "\033[36mRun 'axle help' to get started\033[0m"