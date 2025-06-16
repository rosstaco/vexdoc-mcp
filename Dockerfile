# Multi-stage build for VEX Document MCP Server

# Dependencies stage
FROM node:20-alpine AS deps

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# vexctl builder stage
FROM golang:1.24-alpine AS vexctl-builder

# Install vexctl using go install (works on all architectures)
RUN go install github.com/openvex/vexctl@latest

# Final production stage
FROM node:20-alpine AS production

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy vexctl from Go builder
COPY --from=vexctl-builder /go/bin/vexctl /usr/local/bin/vexctl

# Verify vexctl installation
RUN vexctl version

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S mcp -u 1001 -G nodejs

# Set working directory
WORKDIR /app

# Copy dependencies and source code
COPY --from=deps --chown=mcp:nodejs /app/node_modules ./node_modules
COPY --chown=mcp:nodejs package*.json ./
COPY --chown=mcp:nodejs src ./src

# Switch to non-root user
USER mcp

# Expose port for HTTP transport
EXPOSE 3000

# Set default environment variables
ENV NODE_ENV=production
ENV MCP_TRANSPORT=stdio
ENV MCP_PORT=3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD if [ "$MCP_TRANSPORT" = "http" ] || [ "$MCP_TRANSPORT" = "streaming" ]; then \
        wget --no-verbose --tries=1 --spider http://localhost:$MCP_PORT/ || exit 1; \
    else \
        node -e "console.log('Health check passed')" || exit 1; \
    fi

# Default command
CMD ["sh", "-c", "node src/index.js $MCP_TRANSPORT $MCP_PORT"]
