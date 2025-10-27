# Phase 8: Deployment & Migration
**Story Points**: 3 | **Prerequisites**: [Phase 7](./phase7-testing.md) | **Next**: Project Complete

## Overview
Execute production deployment with gradual migration strategy, monitoring, and rollback capabilities to ensure smooth transition from Node.js to Go implementation.

## Objectives
- [ ] Deploy Go implementation to production safely
- [ ] Migrate users with zero downtime
- [ ] Establish monitoring and alerting
- [ ] Validate production performance
- [ ] Complete documentation and handover

## Tasks

### Task 8.1: Build & Deployment System (1 point)

**Goal**: Production-ready build, packaging, and deployment automation

**Checklist:**
- [ ] Multi-platform binary builds (Linux, macOS, Windows)
- [ ] Docker containerization with optimal image size
- [ ] CI/CD pipeline with automated testing
- [ ] Release management and versioning
- [ ] Package distribution (npm replacement)

**Code Example - Docker Multi-Stage Build:**
```dockerfile
# Dockerfile.production
# Multi-stage build for optimal production image
FROM golang:1.21-alpine AS builder

# Install git for module dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
    -a -installsuffix cgo \
    -o vexdoc-mcp \
    ./cmd/server

# Production image
FROM alpine:3.18

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Create directory for app
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/vexdoc-mcp .

# Copy configuration files
COPY --from=builder /build/configs ./configs

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./vexdoc-mcp health || exit 1

# Expose port (if using HTTP mode)
EXPOSE 8080

# Default to stdio mode
ENTRYPOINT ["./vexdoc-mcp"]
CMD ["stdio"]
```

**Code Example - GitHub Actions CI/CD:**
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Run tests
        run: |
          go test -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
          
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage.html

  build:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64
            
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Build binary
        run: |
          export VERSION=${GITHUB_REF#refs/tags/}
          export BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
          export BINARY_NAME=vexdoc-mcp-${{ matrix.goos }}-${{ matrix.goarch }}
          
          if [ "${{ matrix.goos }}" = "windows" ]; then
            export BINARY_NAME=${BINARY_NAME}.exe
          fi
          
          CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build \
            -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
            -o ${BINARY_NAME} \
            ./cmd/server
            
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: vexdoc-mcp-${{ matrix.goos }}-${{ matrix.goarch }}
          path: vexdoc-mcp-*

  docker:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        
      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
          
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=tag
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: Dockerfile.production
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          build-args: |
            VERSION=${{ github.ref_name }}
            BUILD_TIME=${{ github.event.head_commit.timestamp }}

  npm-package:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: ./dist
          
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          registry-url: 'https://registry.npmjs.org'
          
      - name: Create npm package
        run: |
          mkdir -p npm-package/bin
          
          # Copy binaries to bin directory
          for dir in dist/*/; do
            if [ -d "$dir" ]; then
              cp "$dir"/* npm-package/bin/
            fi
          done
          
          # Create package.json for npm distribution
          cat > npm-package/package.json << EOF
          {
            "name": "vexdoc-mcp-go",
            "version": "${GITHUB_REF#refs/tags/v}",
            "description": "Go implementation of VEX Document MCP Server",
            "main": "index.js",
            "bin": {
              "vexdoc-mcp": "./bin/vexdoc-mcp"
            },
            "scripts": {
              "postinstall": "node postinstall.js"
            },
            "keywords": ["vex", "mcp", "security", "vulnerability"],
            "author": "Your Name",
            "license": "MIT",
            "repository": {
              "type": "git",
              "url": "git+https://github.com/${{ github.repository }}.git"
            }
          }
          EOF
          
          # Create postinstall script for platform detection
          cat > npm-package/postinstall.js << 'EOF'
          #!/usr/bin/env node
          const fs = require('fs');
          const path = require('path');
          const os = require('os');
          
          const platform = os.platform();
          const arch = os.arch();
          
          let binaryName = 'vexdoc-mcp';
          
          // Map Node.js platform/arch to Go's GOOS/GOARCH
          const platformMap = {
            'linux': 'linux',
            'darwin': 'darwin', 
            'win32': 'windows'
          };
          
          const archMap = {
            'x64': 'amd64',
            'arm64': 'arm64'
          };
          
          const goos = platformMap[platform];
          const goarch = archMap[arch];
          
          if (!goos || !goarch) {
            console.error(`Unsupported platform: ${platform}-${arch}`);
            process.exit(1);
          }
          
          const sourceBinary = `vexdoc-mcp-${goos}-${goarch}${platform === 'win32' ? '.exe' : ''}`;
          const targetBinary = `vexdoc-mcp${platform === 'win32' ? '.exe' : ''}`;
          
          const sourcePath = path.join(__dirname, 'bin', sourceBinary);
          const targetPath = path.join(__dirname, 'bin', targetBinary);
          
          if (fs.existsSync(sourcePath)) {
            fs.copyFileSync(sourcePath, targetPath);
            fs.chmodSync(targetPath, 0o755);
            console.log(`Installed ${sourceBinary} as ${targetBinary}`);
          } else {
            console.error(`Binary not found: ${sourcePath}`);
            process.exit(1);
          }
          EOF
          
          chmod +x npm-package/postinstall.js
          
      - name: Publish to npm
        working-directory: npm-package
        run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}

  release:
    needs: [build, docker, npm-package]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: ./dist
          
      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*/*
          generate_release_notes: true
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Code Example - Makefile for Local Development:**
```makefile
# Makefile
.PHONY: build test clean docker npm-package

# Variables
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -w -s -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'

# Default target
all: build

# Build for current platform
build:
	go build -ldflags="$(LDFLAGS)" -o bin/vexdoc-mcp ./cmd/server

# Build for all platforms
build-all: clean
	mkdir -p dist
	
	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o dist/vexdoc-mcp-linux-amd64 \
		./cmd/server
	
	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-ldflags="$(LDFLAGS)" \
		-o dist/vexdoc-mcp-linux-arm64 \
		./cmd/server
	
	# macOS AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o dist/vexdoc-mcp-darwin-amd64 \
		./cmd/server
	
	# macOS ARM64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags="$(LDFLAGS)" \
		-o dist/vexdoc-mcp-darwin-arm64 \
		./cmd/server
	
	# Windows AMD64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o dist/vexdoc-mcp-windows-amd64.exe \
		./cmd/server

# Test targets
test:
	go test -race ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	go test -tags=integration ./test/integration/...

test-compatibility:
	go test -tags=compatibility ./test/compatibility/...

# Docker targets
docker:
	docker build -f Dockerfile.production -t vexdoc-mcp:$(VERSION) .

docker-dev:
	docker build -f Dockerfile.dev -t vexdoc-mcp:dev .

# NPM package
npm-package: build-all
	./scripts/create-npm-package.sh $(VERSION)

# Clean
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

# Development
dev: build
	./bin/vexdoc-mcp stdio

# Install (requires sudo for system install)
install: build
	sudo cp bin/vexdoc-mcp /usr/local/bin/
	sudo chmod +x /usr/local/bin/vexdoc-mcp

# Uninstall
uninstall:
	sudo rm -f /usr/local/bin/vexdoc-mcp
```

**Deliverable**: Complete build and deployment automation

---

### Task 8.2: Migration Strategy & Monitoring (2 points)

**Goal**: Safe production migration with comprehensive monitoring and rollback capabilities

**Checklist:**
- [ ] Gradual rollout strategy (canary deployment)
- [ ] Feature flags for controlled rollout
- [ ] Comprehensive monitoring and alerting
- [ ] Performance and error rate tracking
- [ ] Automated rollback triggers

**Code Example - Migration Strategy:**
```go
// internal/migration/manager.go
package migration

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/internal/logging"
    "github.com/rosstaco/vexdoc-mcp-go/internal/monitoring"
)

type MigrationManager struct {
    logger     logging.Logger
    monitor    monitoring.Monitor
    config     *MigrationConfig
    
    // State management
    mu               sync.RWMutex
    currentStage     MigrationStage
    rolloutPercent   int
    healthStatus     HealthStatus
    
    // Metrics
    errorCount       int
    successCount     int
    lastHealthCheck  time.Time
}

type MigrationConfig struct {
    // Rollout configuration
    InitialPercent      int           `json:"initial_percent"`
    IncrementPercent    int           `json:"increment_percent"`
    StageInterval       time.Duration `json:"stage_interval"`
    
    // Health thresholds
    MaxErrorRate        float64       `json:"max_error_rate"`
    MinSuccessRate      float64       `json:"min_success_rate"`
    HealthCheckInterval time.Duration `json:"health_check_interval"`
    
    // Rollback triggers
    ConsecutiveErrors   int           `json:"consecutive_errors"`
    ErrorRateWindow     time.Duration `json:"error_rate_window"`
    AutoRollbackEnabled bool          `json:"auto_rollback_enabled"`
}

type MigrationStage int

const (
    StagePreparing MigrationStage = iota
    StageCanary
    StageGradual
    StageComplete
    StageRolledBack
)

type HealthStatus int

const (
    HealthUnknown HealthStatus = iota
    HealthHealthy
    HealthDegraded
    HealthUnhealthy
)

func NewMigrationManager(config *MigrationConfig, logger logging.Logger, monitor monitoring.Monitor) *MigrationManager {
    return &MigrationManager{
        logger:          logger,
        monitor:         monitor,
        config:          config,
        currentStage:    StagePreparing,
        rolloutPercent:  0,
        healthStatus:    HealthUnknown,
        lastHealthCheck: time.Now(),
    }
}

func (mm *MigrationManager) StartMigration(ctx context.Context) error {
    mm.logger.Info("Starting migration to Go implementation")
    
    // Stage 1: Canary deployment (5% traffic)
    if err := mm.progressToStage(ctx, StageCanary, mm.config.InitialPercent); err != nil {
        return fmt.Errorf("canary stage failed: %w", err)
    }
    
    // Stage 2: Gradual rollout
    if err := mm.progressToStage(ctx, StageGradual, 100); err != nil {
        return fmt.Errorf("gradual rollout failed: %w", err)
    }
    
    // Stage 3: Complete migration
    if err := mm.progressToStage(ctx, StageComplete, 100); err != nil {
        return fmt.Errorf("final stage failed: %w", err)
    }
    
    mm.logger.Info("Migration completed successfully")
    return nil
}

func (mm *MigrationManager) progressToStage(ctx context.Context, stage MigrationStage, targetPercent int) error {
    mm.mu.Lock()
    mm.currentStage = stage
    mm.mu.Unlock()
    
    // Gradual traffic increase
    for percent := mm.rolloutPercent; percent < targetPercent; percent += mm.config.IncrementPercent {
        if percent > targetPercent {
            percent = targetPercent
        }
        
        mm.logger.Info("Increasing traffic to Go implementation", 
            "stage", stage, "percent", percent)
        
        mm.updateRolloutPercent(percent)
        
        // Wait for stage interval
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(mm.config.StageInterval):
        }
        
        // Health check
        if err := mm.performHealthCheck(ctx); err != nil {
            mm.logger.Error("Health check failed", "error", err)
            
            if mm.config.AutoRollbackEnabled {
                return mm.initiateRollback(ctx, err)
            }
            return err
        }
        
        mm.logger.Info("Stage health check passed", "percent", percent)
    }
    
    return nil
}

func (mm *MigrationManager) updateRolloutPercent(percent int) {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    mm.rolloutPercent = percent
    mm.monitor.SetGauge("migration_rollout_percent", float64(percent))
}

func (mm *MigrationManager) performHealthCheck(ctx context.Context) error {
    mm.mu.Lock()
    mm.lastHealthCheck = time.Now()
    mm.mu.Unlock()
    
    // Check error rate
    errorRate := mm.calculateErrorRate()
    if errorRate > mm.config.MaxErrorRate {
        mm.updateHealthStatus(HealthUnhealthy)
        return fmt.Errorf("error rate too high: %.2f%% > %.2f%%", 
            errorRate*100, mm.config.MaxErrorRate*100)
    }
    
    // Check success rate
    successRate := mm.calculateSuccessRate()
    if successRate < mm.config.MinSuccessRate {
        mm.updateHealthStatus(HealthDegraded)
        return fmt.Errorf("success rate too low: %.2f%% < %.2f%%", 
            successRate*100, mm.config.MinSuccessRate*100)
    }
    
    // Check consecutive errors
    if mm.errorCount >= mm.config.ConsecutiveErrors {
        mm.updateHealthStatus(HealthUnhealthy)
        return fmt.Errorf("too many consecutive errors: %d >= %d", 
            mm.errorCount, mm.config.ConsecutiveErrors)
    }
    
    mm.updateHealthStatus(HealthHealthy)
    mm.resetErrorCount()
    
    return nil
}

func (mm *MigrationManager) calculateErrorRate() float64 {
    mm.mu.RLock()
    defer mm.mu.RUnlock()
    
    total := mm.errorCount + mm.successCount
    if total == 0 {
        return 0
    }
    
    return float64(mm.errorCount) / float64(total)
}

func (mm *MigrationManager) calculateSuccessRate() float64 {
    mm.mu.RLock()
    defer mm.mu.RUnlock()
    
    total := mm.errorCount + mm.successCount
    if total == 0 {
        return 1.0 // Assume healthy if no data
    }
    
    return float64(mm.successCount) / float64(total)
}

func (mm *MigrationManager) updateHealthStatus(status HealthStatus) {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    mm.healthStatus = status
    mm.monitor.SetGauge("migration_health_status", float64(status))
}

func (mm *MigrationManager) resetErrorCount() {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    mm.errorCount = 0
}

func (mm *MigrationManager) initiateRollback(ctx context.Context, reason error) error {
    mm.logger.Error("Initiating automatic rollback", "reason", reason)
    
    mm.mu.Lock()
    mm.currentStage = StageRolledBack
    mm.rolloutPercent = 0
    mm.mu.Unlock()
    
    // Update routing to send all traffic to Node.js
    mm.monitor.SetGauge("migration_rollout_percent", 0)
    mm.monitor.IncrementCounter("migration_rollbacks_total")
    
    // Alert operations team
    mm.monitor.Alert("migration_rollback", "severity", "critical", 
        "reason", reason.Error())
    
    return fmt.Errorf("migration rolled back: %w", reason)
}

func (mm *MigrationManager) RecordSuccess() {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    mm.successCount++
    mm.monitor.IncrementCounter("migration_requests_success_total")
}

func (mm *MigrationManager) RecordError() {
    mm.mu.Lock()
    defer mm.mu.Unlock()
    
    mm.errorCount++
    mm.monitor.IncrementCounter("migration_requests_error_total")
}

func (mm *MigrationManager) GetStatus() MigrationStatus {
    mm.mu.RLock()
    defer mm.mu.RUnlock()
    
    return MigrationStatus{
        Stage:            mm.currentStage,
        RolloutPercent:   mm.rolloutPercent,
        HealthStatus:     mm.healthStatus,
        ErrorRate:        mm.calculateErrorRate(),
        SuccessRate:      mm.calculateSuccessRate(),
        LastHealthCheck:  mm.lastHealthCheck,
    }
}

type MigrationStatus struct {
    Stage           MigrationStage `json:"stage"`
    RolloutPercent  int           `json:"rollout_percent"`
    HealthStatus    HealthStatus  `json:"health_status"`
    ErrorRate       float64       `json:"error_rate"`
    SuccessRate     float64       `json:"success_rate"`
    LastHealthCheck time.Time     `json:"last_health_check"`
}
```

**Code Example - Feature Flag Integration:**
```go
// internal/features/flags.go
package features

import (
    "context"
    "hash/fnv"
    "os"
    "strconv"
    "strings"
)

type FeatureFlags struct {
    flags map[string]*Flag
}

type Flag struct {
    Name         string  `json:"name"`
    Enabled      bool    `json:"enabled"`
    Percentage   int     `json:"percentage"`   // 0-100
    UserSegments []string `json:"user_segments"`
}

func NewFeatureFlags() *FeatureFlags {
    flags := &FeatureFlags{
        flags: make(map[string]*Flag),
    }
    
    // Load from environment variables
    flags.loadFromEnv()
    
    return flags
}

func (ff *FeatureFlags) loadFromEnv() {
    // Migration flags
    ff.flags["use_go_implementation"] = &Flag{
        Name:       "use_go_implementation",
        Enabled:    getEnvBool("USE_GO_IMPLEMENTATION", false),
        Percentage: getEnvInt("GO_IMPL_PERCENTAGE", 0),
    }
    
    ff.flags["enable_streaming"] = &Flag{
        Name:       "enable_streaming",
        Enabled:    getEnvBool("ENABLE_STREAMING", false),
        Percentage: getEnvInt("STREAMING_PERCENTAGE", 0),
    }
    
    ff.flags["performance_monitoring"] = &Flag{
        Name:    "performance_monitoring",
        Enabled: getEnvBool("PERFORMANCE_MONITORING", true),
    }
}

func (ff *FeatureFlags) IsEnabled(ctx context.Context, flagName string, userID string) bool {
    flag, exists := ff.flags[flagName]
    if !exists {
        return false
    }
    
    if !flag.Enabled {
        return false
    }
    
    // If percentage is 100, always enabled
    if flag.Percentage >= 100 {
        return true
    }
    
    // If percentage is 0, never enabled
    if flag.Percentage <= 0 {
        return false
    }
    
    // Use consistent hashing for user-based rollout
    return ff.isInPercentage(userID, flag.Percentage)
}

func (ff *FeatureFlags) isInPercentage(userID string, percentage int) bool {
    if userID == "" {
        return false
    }
    
    hash := fnv.New32a()
    hash.Write([]byte(userID))
    hashValue := hash.Sum32()
    
    bucket := int(hashValue % 100)
    return bucket < percentage
}

func (ff *FeatureFlags) UpdateFlag(flagName string, enabled bool, percentage int) {
    if flag, exists := ff.flags[flagName]; exists {
        flag.Enabled = enabled
        flag.Percentage = percentage
    }
}

func getEnvBool(key string, defaultValue bool) bool {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    
    result, err := strconv.ParseBool(value)
    if err != nil {
        return defaultValue
    }
    
    return result
}

func getEnvInt(key string, defaultValue int) int {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    
    result, err := strconv.Atoi(value)
    if err != nil {
        return defaultValue
    }
    
    return result
}
```

**Code Example - Monitoring & Alerting:**
```go
// internal/monitoring/prometheus.go
package monitoring

import (
    "net/http"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMonitor struct {
    // Request metrics
    requestDuration    *prometheus.HistogramVec
    requestsTotal      *prometheus.CounterVec
    errorsTotal        *prometheus.CounterVec
    
    // Migration metrics
    migrationStage     prometheus.Gauge
    rolloutPercent     prometheus.Gauge
    rollbacksTotal     prometheus.Counter
    
    // Performance metrics
    memoryUsage        prometheus.Gauge
    goroutineCount     prometheus.Gauge
    vexOperationsTotal *prometheus.CounterVec
    streamingActive    prometheus.Gauge
}

func NewPrometheusMonitor() *PrometheusMonitor {
    pm := &PrometheusMonitor{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "vexdoc_mcp_request_duration_seconds",
                Help: "Request duration in seconds",
                Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
            },
            []string{"method", "status", "implementation"},
        ),
        
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "vexdoc_mcp_requests_total",
                Help: "Total number of requests",
            },
            []string{"method", "status", "implementation"},
        ),
        
        errorsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "vexdoc_mcp_errors_total",
                Help: "Total number of errors",
            },
            []string{"type", "implementation"},
        ),
        
        migrationStage: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "vexdoc_mcp_migration_stage",
                Help: "Current migration stage",
            },
        ),
        
        rolloutPercent: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "vexdoc_mcp_rollout_percent",
                Help: "Go implementation rollout percentage",
            },
        ),
        
        rollbacksTotal: prometheus.NewCounter(
            prometheus.CounterOpts{
                Name: "vexdoc_mcp_rollbacks_total",
                Help: "Total number of rollbacks",
            },
        ),
        
        memoryUsage: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "vexdoc_mcp_memory_bytes",
                Help: "Memory usage in bytes",
            },
        ),
        
        goroutineCount: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "vexdoc_mcp_goroutines",
                Help: "Number of goroutines",
            },
        ),
        
        vexOperationsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "vexdoc_mcp_vex_operations_total",
                Help: "Total VEX operations",
            },
            []string{"operation", "status"},
        ),
        
        streamingActive: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "vexdoc_mcp_streaming_active",
                Help: "Number of active streaming operations",
            },
        ),
    }
    
    // Register metrics
    prometheus.MustRegister(
        pm.requestDuration,
        pm.requestsTotal,
        pm.errorsTotal,
        pm.migrationStage,
        pm.rolloutPercent,
        pm.rollbacksTotal,
        pm.memoryUsage,
        pm.goroutineCount,
        pm.vexOperationsTotal,
        pm.streamingActive,
    )
    
    return pm
}

func (pm *PrometheusMonitor) RecordRequest(method, status, implementation string, duration time.Duration) {
    pm.requestDuration.WithLabelValues(method, status, implementation).Observe(duration.Seconds())
    pm.requestsTotal.WithLabelValues(method, status, implementation).Inc()
}

func (pm *PrometheusMonitor) RecordError(errorType, implementation string) {
    pm.errorsTotal.WithLabelValues(errorType, implementation).Inc()
}

func (pm *PrometheusMonitor) SetMigrationStage(stage float64) {
    pm.migrationStage.Set(stage)
}

func (pm *PrometheusMonitor) SetRolloutPercent(percent float64) {
    pm.rolloutPercent.Set(percent)
}

func (pm *PrometheusMonitor) IncrementRollbacks() {
    pm.rollbacksTotal.Inc()
}

func (pm *PrometheusMonitor) UpdateMemoryUsage(bytes float64) {
    pm.memoryUsage.Set(bytes)
}

func (pm *PrometheusMonitor) UpdateGoroutineCount(count float64) {
    pm.goroutineCount.Set(count)
}

func (pm *PrometheusMonitor) RecordVEXOperation(operation, status string) {
    pm.vexOperationsTotal.WithLabelValues(operation, status).Inc()
}

func (pm *PrometheusMonitor) SetStreamingActive(count float64) {
    pm.streamingActive.Set(count)
}

func (pm *PrometheusMonitor) Handler() http.Handler {
    return promhttp.Handler()
}
```

**Deliverable**: Production migration system with monitoring and automated rollback

---

## Phase 8 Deliverables

### 1. Build System (`scripts/`, `.github/workflows/`)
- [ ] Multi-platform binary builds (Linux, macOS, Windows)
- [ ] Docker containerization with optimized images
- [ ] CI/CD pipeline with automated testing and deployment
- [ ] NPM package distribution for easy installation

### 2. Migration Infrastructure (`internal/migration/`)
- [ ] Gradual rollout manager with configurable stages
- [ ] Feature flag system for controlled traffic routing
- [ ] Health monitoring and automatic rollback triggers
- [ ] Migration status dashboard and reporting

### 3. Monitoring & Observability (`internal/monitoring/`)
- [ ] Prometheus metrics for all key operations
- [ ] Performance comparison dashboards
- [ ] Error rate and success rate tracking
- [ ] Alert rules for critical migration events

### 4. Documentation
- [ ] Production deployment guide
- [ ] Migration runbook and procedures
- [ ] Monitoring and alerting documentation
- [ ] Rollback procedures and troubleshooting

### 5. Package Distribution
- [ ] GitHub releases with binaries
- [ ] Docker images in container registry
- [ ] NPM package for Node.js ecosystem compatibility
- [ ] Installation and upgrade instructions

## Success Criteria
- [ ] Zero-downtime migration completed successfully
- [ ] All monitoring and alerting operational
- [ ] Performance improvements validated in production
- [ ] Rollback procedures tested and documented
- [ ] User migration completed with minimal friction
- [ ] Documentation complete and accessible

## Migration Timeline
```
Week 1: Build system and CI/CD setup
Week 2: Migration infrastructure and monitoring
Week 3: Canary deployment (5% traffic)
Week 4: Gradual rollout (25%, 50%, 75%, 100%)
```

## Quality Gates
- **Build System**: All platforms build successfully, tests pass
- **Migration System**: Rollout controls work, health checks functional
- **Monitoring**: All metrics collected, alerts configured
- **Documentation**: Complete deployment and migration guides
- **Production**: Performance targets met, zero critical issues

## Dependencies
- **Input**: Phase 7 tested and validated Go implementation
- **Output**: Production Go MCP server fully deployed

## Risks & Mitigation
- **Risk**: Migration causes user disruption
  - **Mitigation**: Gradual rollout with automatic rollback triggers
- **Risk**: Performance doesn't meet expectations in production
  - **Mitigation**: Comprehensive monitoring and quick rollback capability
- **Risk**: Compatibility issues with existing clients
  - **Mitigation**: Extensive testing and feature flag controls

## Time Estimate
**3 Story Points** ≈ 1-2 days of focused deployment work

---

## 🎉 Migration Complete!

Upon completion of Phase 8, you will have successfully:

- ✅ **Native Go Implementation**: Eliminated subprocess overhead with direct VEX library integration
- ✅ **Streaming Capabilities**: Enabled native streaming for large document processing
- ✅ **Performance Gains**: Achieved 50%+ latency reduction and improved memory efficiency
- ✅ **Production Ready**: Deployed with monitoring, rollback capabilities, and zero-downtime migration
- ✅ **Maintained Compatibility**: Preserved API compatibility with existing MCP clients

**Total Effort**: 45 Story Points across 8 phases

**Key Benefits Realized**:
- Native streaming eliminates CLI subprocess overhead
- Direct VEX library integration provides better error handling
- Go's performance characteristics deliver measurable improvements
- Production deployment maintains compatibility and reliability

**Next Steps**: Monitor production performance, gather user feedback, and iterate on streaming optimizations based on real-world usage patterns.
