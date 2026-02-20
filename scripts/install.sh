#!/bin/bash

# skli installation script for Mac and Linux
# This script downloads the skli archive, extracts it, and installs it to /usr/local/bin.

set -e

# Configuration
REPO="Bafoj/Skli"
BINARY_NAME="skli"
INSTALL_DIR="${HOME}/.skli/bin"
VERSION="0.1.2" # Updated automatically by 'make tag'

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$OS" != "darwin" ] && [ "$OS" != "linux" ]; then
    echo "Unsupported OS: $OS"
    exit 1
fi

# GoReleaser naming convention - default is tar.gz but we handle zip too
# We try tar.gz first as it's the standard for Linux/Mac in our config
ARCHIVE_EXT="tar.gz"
ARCHIVE_NAME="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.${ARCHIVE_EXT}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/v$VERSION/$ARCHIVE_NAME"

echo "Downloading $BINARY_NAME $VERSION for $OS/$ARCH..."

# Attempt download
if ! curl -fL "$DOWNLOAD_URL" -o "/tmp/$ARCHIVE_NAME"; then
    echo "Could not find $ARCHIVE_EXT, trying zip..."
    ARCHIVE_EXT="zip"
    ARCHIVE_NAME="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.${ARCHIVE_EXT}"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/v$VERSION/$ARCHIVE_NAME"
    curl -fL "$DOWNLOAD_URL" -o "/tmp/$ARCHIVE_NAME"
fi

echo "Extracting ($ARCHIVE_EXT)..."
mkdir -p "/tmp/skli_extract"

if [ "$ARCHIVE_EXT" = "tar.gz" ]; then
    tar -xzf "/tmp/$ARCHIVE_NAME" -C "/tmp/skli_extract"
else
    if command -v unzip >/dev/null 2>&1; then
        unzip -q "/tmp/$ARCHIVE_NAME" -d "/tmp/skli_extract"
    else
        echo "Error: unzip is required to extract the .zip archive."
        exit 1
    fi
fi

echo "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
chmod +x "/tmp/skli_extract/$BINARY_NAME"
mv "/tmp/skli_extract/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

# Clean up
rm -rf "/tmp/skli_extract" "/tmp/$ARCHIVE_NAME"

echo "Successfully installed skli!"

# Check if INSTALL_DIR is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "Warning: $INSTALL_DIR is not in your PATH."
    echo "To be able to run 'skli' from anywhere, add it to your profile:"
    echo ""
    if [[ "$SHELL" == */zsh ]]; then
        echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
    elif [[ "$SHELL" == */bash ]]; then
        if [[ "$OS" == "darwin" ]]; then
            echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bash_profile"
            echo "  source ~/.bash_profile"
        else
            echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
            echo "  source ~/.bashrc"
        fi
    else
        echo "  Add $INSTALL_DIR to your shell's PATH environment variable."
    fi
    echo ""
fi

echo "Run 'skli help' to get started."
