#!/bin/bash

# Post-create script for VexDoc MCP devcontainer
set -e

echo "🚀 Setting up development environment..."

# Install vexctl@latest
echo "📦 Installing vexctl@latest..."
if command -v go &> /dev/null; then
    # Install vexctl using go install
    go install github.com/openvex/vexctl@latest
    echo "✅ vexctl installed successfully via go install"
else
    echo "❌ Go not found, cannot install vexctl"
    exit 1
fi

# Verify installations
echo "🔍 Verifying installations..."

# Check Node.js
if command -v node &> /dev/null; then
    echo "✅ Node.js $(node --version) is installed"
else
    echo "❌ Node.js not found"
fi

# Check npm
if command -v npm &> /dev/null; then
    echo "✅ npm $(npm --version) is installed"
else
    echo "❌ npm not found"
fi

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go $(go version | cut -d' ' -f3) is installed"
else
    echo "❌ Go not found"
fi

# Check vexctl
if command -v vexctl &> /dev/null; then
    echo "✅ vexctl $(vexctl version) is installed"
else
    echo "❌ vexctl not found in PATH"
    echo "🔧 Checking if vexctl is in GOPATH/bin..."
    if [ -f "$HOME/go/bin/vexctl" ]; then
        echo "✅ vexctl found in $HOME/go/bin/"
        echo "💡 Make sure $HOME/go/bin is in your PATH"
    fi
fi

# Set up git if not configured
if [ -z "$(git config --global user.name)" ]; then
    echo "⚙️  Git user not configured. You may want to run:"
    echo "   git config --global user.name 'Your Name'"
    echo "   git config --global user.email 'your.email@example.com'"
fi

echo "🎉 Development environment setup complete!"
echo ""
echo "🛠️  Available tools:"
echo "   - Node.js LTS with npm"
echo "   - Go latest"
echo "   - vexctl latest"
echo "   - Git and GitHub CLI"
echo ""
echo "📁 Workspace: /workspaces/vexdoc-mcp"
echo ""
