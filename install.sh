#!/bin/bash

# Bui CLI Installation Script
# This script installs the bui binary to ~/.base/bin or /usr/local/bin
#
# IMPORTANT SAFETY NOTES:
# - This script NEVER deletes the .base directory or .base/bin directory
# - ~/.base/bin/ is shared by all Base Framework CLIs (base, bui, etc.)
# - This script ONLY manages the 'bui' binary, all other binaries are preserved
# - Never run commands that remove directories or other binaries

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

# Interactive installation prompt (only for non-Windows and TTY)
GLOBAL_INSTALL=false
if [ "$OS" != "windows" ] && [ -t 0 ]; then
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║              Bui CLI Installation                          ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "Choose installation type:"
    echo ""
    echo "  1) Local install (Recommended)"
    echo "     • Location: ~/.base/bin"
    echo "     • No sudo required"
    echo "     • You'll need to add ~/.base/bin to your PATH"
    echo ""
    echo "  2) Global install (sudo)"
    echo "     • Location: /usr/local/bin"
    echo "     • Requires sudo password"
    echo "     • Available system-wide immediately"
    echo ""
    read -p "Enter your choice [1/2] (default: 1): " INSTALL_CHOICE

    if [ "$INSTALL_CHOICE" = "2" ]; then
        GLOBAL_INSTALL=true
        echo ""
        echo "Global installation requires sudo access..."
        if ! sudo -v; then
            echo "❌ Error: sudo access required for global installation"
            echo "Please run the script again and choose option 1 for local installation."
            exit 1
        fi
        echo "✓ Sudo access granted"
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

# Create installation directories (never delete existing directories)
if [ "$GLOBAL_INSTALL" = true ]; then
    sudo mkdir -p "$INSTALL_DIR"
    sudo mkdir -p "$BIN_DIR"
else
    # Create directories if they don't exist, but never remove existing ones
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$BIN_DIR"
fi

# Check if bui is already installed (only if running interactively)
if [ -f "$BIN_DIR/$BINARY_NAME" ] && [ -t 0 ]; then
    EXISTING_VERSION="unknown"
    if [ "$GLOBAL_INSTALL" = false ]; then
        EXISTING_VERSION=$("$BIN_DIR/$BINARY_NAME" version 2>/dev/null | grep -o 'v[0-9]*\.[0-9]*\.[0-9]*' || echo "unknown")
    fi
    echo ""
    echo "⚠️  Bui CLI is already installed"
    if [ "$EXISTING_VERSION" != "unknown" ]; then
        echo "   Current version: $EXISTING_VERSION"
    fi
    echo ""
    read -p "Do you want to reinstall/update? [y/N]: " REINSTALL
    if [[ ! $REINSTALL =~ ^[Yy]$ ]]; then
        echo "Installation cancelled."
        exit 0
    fi
    echo ""
fi

echo ""
echo "════════════════════════════════════════════════════════════"
echo "  Installing Bui CLI"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "Platform: $OS $ARCH"

# Get the latest release version
echo "→ Fetching latest release information..."
API_RESPONSE=$(curl -s https://api.github.com/repos/base-al/bui/releases/latest)
if [ $? -ne 0 ]; then
    echo "❌ Error: Failed to fetch release information"
    exit 1
fi

LATEST_RELEASE=$(echo "$API_RESPONSE" | grep '"tag_name"' | head -n1 | cut -d '"' -f 4)
if [ -z "$LATEST_RELEASE" ]; then
    echo "❌ Error: Could not determine latest version"
    echo "API Response debug info:"
    echo "$API_RESPONSE" | head -n 10
    echo "Please check if the repository exists and has releases"
    echo "Repository: https://github.com/base-al/bui"
    exit 1
fi

echo "✓ Latest version: $LATEST_RELEASE"

# Download the appropriate binary
DOWNLOAD_URL="https://github.com/base-al/bui/releases/download/$LATEST_RELEASE/bui_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="https://github.com/base-al/bui/releases/download/$LATEST_RELEASE/bui_${OS}_${ARCH}.zip"
fi

echo ""
echo "→ Downloading Bui CLI binary..."
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if [ "$OS" = "windows" ]; then
    if ! curl -sL "$DOWNLOAD_URL" -o bui.zip; then
        echo "❌ Error: Download failed"
        exit 1
    fi
    unzip -q bui.zip
else
    if ! curl -sL "$DOWNLOAD_URL" | tar xz; then
        echo "❌ Error: Download failed"
        exit 1
    fi
fi

echo "✓ Download complete"

# Install the binary
echo ""
echo "→ Installing to $BIN_DIR..."
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

echo "✓ Installation complete"

# Install Go dependencies
echo ""
echo "════════════════════════════════════════════════════════════"
echo "  Installing Dependencies"
echo "════════════════════════════════════════════════════════════"
echo ""

# Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo "→ Installing swag (API documentation generator)..."
    if ! go install github.com/swaggo/swag/cmd/swag@latest 2>/dev/null; then
        echo "⚠️  Warning: Failed to install swag"
        echo "   You can install it manually later with:"
        echo "   go install github.com/swaggo/swag/cmd/swag@latest"
    else
        echo "✓ swag installed successfully"
    fi
else
    echo "⚠️  Go is not installed or not in PATH"
    echo "   Bui CLI dependencies (swag) will be installed automatically when needed."
    echo "   To install Go, visit: https://golang.org/dl/"
fi

echo ""
echo "════════════════════════════════════════════════════════════"
echo "  Installation Summary"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "✓ Bui CLI $LATEST_RELEASE installed successfully!"
echo "  Location: $BIN_DIR/bui"

# Only show PATH instructions for local installations
if [ "$GLOBAL_INSTALL" = false ]; then
    # Check if ~/.base/bin is in PATH
    if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        echo ""
        echo "════════════════════════════════════════════════════════════"
        echo "  ⚠️  Action Required: Add to PATH"
        echo "════════════════════════════════════════════════════════════"
        echo ""
        if [ "$OS" = "darwin" ] || [ "$OS" = "linux" ]; then
            # Detect shell
            SHELL_CONFIG=""
            if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
                SHELL_CONFIG="~/.zshrc"
            elif [ -f "$HOME/.bashrc" ]; then
                SHELL_CONFIG="~/.bashrc"
            else
                SHELL_CONFIG="your shell configuration file"
            fi

            echo "Add this line to $SHELL_CONFIG:"
            echo ""
            echo '    export PATH="$HOME/.base/bin:$PATH"'
            echo ""
            echo "Then reload your shell:"
            echo ""
            echo "    source $SHELL_CONFIG"
        elif [ "$OS" = "windows" ]; then
            echo "Run this command:"
            echo ""
            echo "    setx PATH \"%PATH%;$BIN_DIR\""
        fi
        echo ""
    else
        echo ""
        echo "✓ ~/.base/bin is already in your PATH"
    fi
fi

echo ""
echo "════════════════════════════════════════════════════════════"
echo "  🎉 Ready to Go!"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "Get started:"
echo "  bui --help              Show all commands"
echo "  bui new my-project      Create a new project"
echo ""