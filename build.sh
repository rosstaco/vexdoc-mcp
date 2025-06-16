#!/bin/bash

# Build script for VEX Document MCP Server Docker image
# Supports multi-platform builds

set -e

IMAGE_NAME="vexdoc-mcp"
TAG="${1:-latest}"
PUSH="${2:-false}"

echo "🔨 Building VEX Document MCP Server Docker image..."
echo "📦 Image: ${IMAGE_NAME}:${TAG}"
echo "🚀 Push: ${PUSH}"

# Check if buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    echo "❌ Docker buildx is required for multi-platform builds"
    echo "💡 Please install Docker Desktop or enable buildx"
    exit 1
fi

# Create builder if it doesn't exist
if ! docker buildx ls | grep -q multiarch; then
    echo "🔧 Creating multiarch builder..."
    docker buildx create --name multiarch --use --bootstrap
else
    echo "✅ Using existing multiarch builder"
    docker buildx use multiarch
fi

# Build arguments
BUILD_ARGS=""
if [ "$PUSH" = "true" ]; then
    BUILD_ARGS="--push"
else
    BUILD_ARGS="--load"
fi

echo "🏗️  Building for multiple architectures..."

# Build for multiple platforms
docker buildx build \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    --tag "${IMAGE_NAME}:${TAG}" \
    --tag "${IMAGE_NAME}:latest" \
    ${BUILD_ARGS} \
    .

if [ "$PUSH" = "true" ]; then
    echo "✅ Multi-platform image pushed successfully!"
    echo "🐳 Available architectures: linux/amd64, linux/arm64, linux/arm/v7"
else
    echo "✅ Multi-platform image built successfully!"
    echo "🏃 To run: docker run -p 3000:3000 -e MCP_TRANSPORT=http ${IMAGE_NAME}:${TAG}"
fi

echo ""
echo "📋 Usage examples:"
echo "  # stdio transport"
echo "  docker run --rm ${IMAGE_NAME}:${TAG}"
echo ""
echo "  # HTTP transport"
echo "  docker run --rm -p 3000:3000 -e MCP_TRANSPORT=http ${IMAGE_NAME}:${TAG}"
echo ""
echo "  # With custom port"
echo "  docker run --rm -p 8080:8080 -e MCP_TRANSPORT=http -e MCP_PORT=8080 ${IMAGE_NAME}:${TAG}"
