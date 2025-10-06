#!/bin/bash

set -e

# Detect OS and architecture
detect_os() {
    case "$(uname -s)" in
        Darwin*)   echo "darwin" ;;
        Linux*)    echo "linux" ;;
        MINGW64*|MSYS*|CYGWIN*) echo "windows" ;;
        *)         echo "unknown" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)            echo "unknown" ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
    echo "Unsupported operating system or architecture"
    exit 1
fi

# Binary name based on OS
if [ "$OS" = "windows" ]; then
    BINARY_NAME="bui.exe"
else
    BINARY_NAME="bui"
fi

# Ask for installation type (only for non-Windows)
GLOBAL_INSTALL=false
if [ "$OS" != "windows" ] && [ -t 0 ]; then
    echo ""
    echo "Choose installation type:"
    echo "  1) Local install (~/.base/bin) - No sudo required"
    echo "  2) Global install (/usr/local/bin) - Requires sudo"
    echo ""
    read -p "Enter your choice [1/2] (default: 1): " INSTALL_CHOICE

    if [ "$INSTALL_CHOICE" = "2" ]; then
        GLOBAL_INSTALL=true
        # Request sudo upfront
        echo ""
        echo "Requesting sudo access for global installation..."
        if ! sudo -v; then
            echo "Error: sudo access required for global installation"
            exit 1
        fi
    fi
fi

# Set installation directories based on choice
if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$USERPROFILE/.base"
    BIN_DIR="$USERPROFILE/bin"
elif [ "$GLOBAL_INSTALL" = true ]; then
    INSTALL_DIR="/usr/local/lib/bui"
    BIN_DIR="/usr/local/bin"
else
    INSTALL_DIR="$HOME/.base"
    BIN_DIR="$HOME/.base/bin"
fi

# Create installation directories
if [ "$GLOBAL_INSTALL" = true ]; then
    sudo mkdir -p "$INSTALL_DIR"
    sudo mkdir -p "$BIN_DIR"
else
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$BIN_DIR"
fi

echo "Installing Bui CLI..."
echo "OS: $OS"
echo "Architecture: $ARCH"

# Get the latest release version
echo "Fetching latest release information..."
API_RESPONSE=$(curl -s https://api.github.com/repos/base-al/bui/releases/latest)
if [ $? -ne 0 ]; then
    echo "Error: Failed to fetch release information"
    exit 1
fi

LATEST_RELEASE=$(echo "$API_RESPONSE" | grep '"tag_name"' | head -n1 | cut -d '"' -f 4)
if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not determine latest version"
    echo "API Response debug info:"
    echo "$API_RESPONSE" | head -n 10
    echo "Please check if the repository exists and has releases"
    echo "Repository: https://github.com/base-al/bui"
    exit 1
fi

echo "Latest version: $LATEST_RELEASE"

# Download the appropriate binary
DOWNLOAD_URL="https://github.com/base-al/bui/releases/download/$LATEST_RELEASE/bui_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="https://github.com/base-al/bui/releases/download/$LATEST_RELEASE/bui_${OS}_${ARCH}.zip"
fi

echo "Downloading from: $DOWNLOAD_URL"
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if [ "$OS" = "windows" ]; then
    curl -sL "$DOWNLOAD_URL" -o bui.zip
    unzip bui.zip
else
    curl -sL "$DOWNLOAD_URL" | tar xz
fi

# Install the binary
if [ "$GLOBAL_INSTALL" = true ]; then
    sudo mv "$BINARY_NAME" "$BIN_DIR/"
    sudo chmod +x "$BIN_DIR/$BINARY_NAME"
else
    mv "$BINARY_NAME" "$BIN_DIR/"
    chmod +x "$BIN_DIR/$BINARY_NAME"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

echo "Bui CLI has been installed successfully!"

# Install Go dependencies
echo ""
echo "Installing Bui CLI dependencies..."

# Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo "Installing swag (API documentation generator)..."
    if ! go install github.com/swaggo/swag/cmd/swag@latest 2>/dev/null; then
        echo "Warning: Failed to install swag. You can install it manually later with:"
        echo "   go install github.com/swaggo/swag/cmd/swag@latest"
    else
        echo "✓ swag installed successfully"
    fi
else
    echo "Warning: Go is not installed or not in PATH."
    echo "Bui CLI dependencies (swag) will be installed automatically when needed."
    echo "To install Go, visit: https://golang.org/dl/"
fi

echo ""
echo "Bui CLI is installed in: $BIN_DIR"

# Only show PATH instructions for local installations
if [ "$GLOBAL_INSTALL" = false ]; then
    # Check if ~/.base/bin is in PATH
    if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        echo ""
        echo "⚠️  Please add ~/.base/bin to your PATH:"
        echo ""
        if [ "$OS" = "darwin" ] || [ "$OS" = "linux" ]; then
            # Detect shell
            if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
                echo "Add this line to your ~/.zshrc:"
                echo '    export PATH="$HOME/.base/bin:$PATH"'
            elif [ -f "$HOME/.bashrc" ]; then
                echo "Add this line to your ~/.bashrc:"
                echo '    export PATH="$HOME/.base/bin:$PATH"'
            else
                echo "Add this line to your shell configuration file:"
                echo '    export PATH="$HOME/.base/bin:$PATH"'
            fi
            echo ""
            echo "Then reload your shell:"
            echo "    source ~/.zshrc  # or ~/.bashrc"
        elif [ "$OS" = "windows" ]; then
            echo "Run this command:"
            echo "    setx PATH \"%PATH%;$BIN_DIR\""
        fi
        echo ""
    else
        echo ""
        echo "✓ ~/.base/bin is already in your PATH"
        echo ""
    fi
else
    echo ""
    echo "✓ Installed globally - bui is available system-wide"
    echo ""
fi

echo "Run 'bui --help' to get started"
echo "Create a new project with: bui new my-project"