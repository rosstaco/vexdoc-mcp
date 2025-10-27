# Phase 6: Streaming & Performance
**Story Points**: 5 | **Prerequisites**: [Phase 5](./phase5-tools.md) | **Next**: [Phase 7](./phase7-testing.md)

## Overview
Optimize streaming capabilities and overall performance to achieve the 50%+ improvement target over the Node.js implementation.

## Objectives
- [ ] Optimize streaming HTTP transport for real-time operations
- [ ] Implement memory-efficient processing for large document sets
- [ ] Add performance monitoring and metrics
- [ ] Optimize VEX operations for speed and memory usage
- [ ] Add benchmarking and performance validation

## Tasks

### Task 6.1: Streaming HTTP Transport Optimization (2 points)

**Goal**: Enhance HTTP transport with efficient streaming capabilities

**Checklist:**
- [ ] Implement Server-Sent Events (SSE) for streaming responses
- [ ] Add WebSocket support for bidirectional streaming
- [ ] Optimize memory usage for concurrent streaming operations
- [ ] Add streaming response chunking and flow control
- [ ] Test with high-concurrency scenarios

**Code Example - Streaming HTTP Transport:**
```go
// internal/mcp/streaming_transport.go
package mcp

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
    "github.com/rosstaco/vexdoc-mcp-go/internal/logging"
)

type StreamingHTTPTransport struct {
    server    *http.Server
    logger    logging.Logger
    requestCh chan *StreamingRequest
    
    // Active streaming connections
    streams   map[string]*StreamingConnection
    streamsMu sync.RWMutex
}

type StreamingRequest struct {
    Request *api.Request
    StreamID string
    ResponseWriter http.ResponseWriter
}

type StreamingConnection struct {
    ID       string
    Writer   http.ResponseWriter
    Flusher  http.Flusher
    Done     chan struct{}
    LastSeen time.Time
}

func NewStreamingHTTPTransport(port int, logger logging.Logger) *StreamingHTTPTransport {
    transport := &StreamingHTTPTransport{
        logger:    logger,
        requestCh: make(chan *StreamingRequest, 100),
        streams:   make(map[string]*StreamingConnection),
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/mcp/stream", transport.handleStreaming)
    mux.HandleFunc("/mcp/sse", transport.handleSSE)
    mux.HandleFunc("/mcp", transport.handleStandardMCP)
    
    transport.server = &http.Server{
        Addr:         fmt.Sprintf(":%d", port),
        Handler:      mux,
        WriteTimeout: 30 * time.Second,
        ReadTimeout:  30 * time.Second,
    }
    
    return transport
}

func (t *StreamingHTTPTransport) Start(ctx context.Context) error {
    t.logger.Info("Starting streaming HTTP transport", "addr", t.server.Addr)
    
    // Start cleanup goroutine for stale connections
    go t.cleanupConnections(ctx)
    
    go func() {
        if err := t.server.ListenAndServe(); err != http.ErrServerClosed {
            t.logger.Error("Streaming HTTP server error", "error", err)
        }
    }()
    
    return nil
}

func (t *StreamingHTTPTransport) Stop() error {
    t.logger.Info("Stopping streaming HTTP transport")
    
    // Close all active streams
    t.streamsMu.Lock()
    for _, stream := range t.streams {
        close(stream.Done)
    }
    t.streamsMu.Unlock()
    
    return t.server.Shutdown(context.Background())
}

func (t *StreamingHTTPTransport) Receive() <-chan *api.Request {
    // Convert streaming requests to regular requests
    requestCh := make(chan *api.Request, 100)
    
    go func() {
        defer close(requestCh)
        for streamReq := range t.requestCh {
            requestCh <- streamReq.Request
        }
    }()
    
    return requestCh
}

func (t *StreamingHTTPTransport) Send(response *api.Response) error {
    // This will be handled by streaming connections directly
    return nil
}

func (t *StreamingHTTPTransport) SendToStream(streamID string, response *api.Response) error {
    t.streamsMu.RLock()
    stream, exists := t.streams[streamID]
    t.streamsMu.RUnlock()
    
    if !exists {
        return fmt.Errorf("stream not found: %s", streamID)
    }
    
    // Send as Server-Sent Event
    data, err := json.Marshal(response)
    if err != nil {
        return fmt.Errorf("marshaling response: %w", err)
    }
    
    _, err = fmt.Fprintf(stream.Writer, "data: %s\n\n", string(data))
    if err != nil {
        return fmt.Errorf("writing to stream: %w", err)
    }
    
    stream.Flusher.Flush()
    stream.LastSeen = time.Now()
    
    return nil
}

func (t *StreamingHTTPTransport) handleSSE(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    
    streamID := r.URL.Query().Get("stream_id")
    if streamID == "" {
        streamID = fmt.Sprintf("stream_%d", time.Now().UnixNano())
    }
    
    // Create streaming connection
    stream := &StreamingConnection{
        ID:       streamID,
        Writer:   w,
        Flusher:  flusher,
        Done:     make(chan struct{}),
        LastSeen: time.Now(),
    }
    
    t.streamsMu.Lock()
    t.streams[streamID] = stream
    t.streamsMu.Unlock()
    
    // Clean up on disconnect
    defer func() {
        t.streamsMu.Lock()
        delete(t.streams, streamID)
        t.streamsMu.Unlock()
    }()
    
    t.logger.Info("SSE connection established", "stream_id", streamID)
    
    // Send initial connection event
    fmt.Fprintf(w, "data: {\"type\":\"connected\",\"stream_id\":\"%s\"}\n\n", streamID)
    flusher.Flush()
    
    // Keep connection alive
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-r.Context().Done():
            return
        case <-stream.Done:
            return
        case <-ticker.C:
            // Send keep-alive ping
            fmt.Fprintf(w, "data: {\"type\":\"ping\"}\n\n")
            flusher.Flush()
        }
    }
}

func (t *StreamingHTTPTransport) handleStreaming(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req api.Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    streamID := r.Header.Get("X-Stream-ID")
    if streamID == "" {
        streamID = fmt.Sprintf("req_%v", req.ID)
    }
    
    // Send request for processing
    streamingReq := &StreamingRequest{
        Request:        &req,
        StreamID:       streamID,
        ResponseWriter: w,
    }
    
    select {
    case t.requestCh <- streamingReq:
        // Response will be sent via streaming
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, "{\"stream_id\":\"%s\",\"status\":\"processing\"}\n", streamID)
    case <-r.Context().Done():
        http.Error(w, "Request cancelled", http.StatusRequestTimeout)
    }
}

func (t *StreamingHTTPTransport) handleStandardMCP(w http.ResponseWriter, r *http.Request) {
    // Fallback to standard HTTP transport for non-streaming requests
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req api.Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // For standard requests, use synchronous processing
    // This would integrate with the main protocol handler
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, "{\"error\":\"Use streaming endpoint for this request\"}\n")
}

func (t *StreamingHTTPTransport) cleanupConnections(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            t.streamsMu.Lock()
            for id, stream := range t.streams {
                if time.Since(stream.LastSeen) > 10*time.Minute {
                    close(stream.Done)
                    delete(t.streams, id)
                    t.logger.Info("Cleaned up stale stream", "stream_id", id)
                }
            }
            t.streamsMu.Unlock()
        }
    }
}
```

**Deliverable**: Enhanced streaming transport in `internal/mcp/streaming_transport.go`

---

### Task 6.2: Memory and Performance Optimization (2 points)

**Goal**: Optimize memory usage and processing speed for large operations

**Checklist:**
- [ ] Implement object pooling for frequently allocated objects
- [ ] Add memory-efficient JSON processing for large documents
- [ ] Optimize VEX document parsing and serialization
- [ ] Add goroutine pools for concurrent processing
- [ ] Implement memory usage monitoring

**Code Example - Performance Optimizations:**
```go
// internal/performance/pools.go
package performance

import (
    "encoding/json"
    "sync"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

var (
    // Object pools for reusing allocations
    responsePool = sync.Pool{
        New: func() interface{} {
            return &api.ToolResponse{
                Content: make([]api.Content, 0, 2),
            }
        },
    }
    
    bufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 0, 1024)
        },
    }
    
    jsonEncoderPool = sync.Pool{
        New: func() interface{} {
            return json.NewEncoder(nil)
        },
    }
)

// GetResponse returns a response from the pool
func GetResponse() *api.ToolResponse {
    resp := responsePool.Get().(*api.ToolResponse)
    resp.Content = resp.Content[:0] // Reset slice
    resp.IsError = false
    return resp
}

// PutResponse returns a response to the pool
func PutResponse(resp *api.ToolResponse) {
    responsePool.Put(resp)
}

// GetBuffer returns a buffer from the pool
func GetBuffer() []byte {
    return bufferPool.Get().([]byte)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf []byte) {
    bufferPool.Put(buf[:0]) // Reset slice
}

// GetJSONEncoder returns a JSON encoder from the pool
func GetJSONEncoder() *json.Encoder {
    return jsonEncoderPool.Get().(*json.Encoder)
}

// PutJSONEncoder returns a JSON encoder to the pool
func PutJSONEncoder(encoder *json.Encoder) {
    jsonEncoderPool.Put(encoder)
}
```

**Code Example - Optimized VEX Processing:**
```go
// internal/vex/optimized.go
package vex

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "runtime"
    "sync"
    
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
    "github.com/rosstaco/vexdoc-mcp-go/internal/performance"
)

// OptimizedClient provides high-performance VEX operations
type OptimizedClient struct {
    *Client
    workerPool chan struct{}
}

func NewOptimizedClient(config *api.VEXConfig, logger logging.Logger) *OptimizedClient {
    // Create worker pool with size based on CPU count
    poolSize := runtime.NumCPU() * 2
    workerPool := make(chan struct{}, poolSize)
    
    return &OptimizedClient{
        Client:     NewClient(config, logger),
        workerPool: workerPool,
    }
}

func (c *OptimizedClient) StreamMergeOptimized(ctx context.Context, opts *api.StreamMergeOptions) (<-chan *api.MergeResult, error) {
    resultCh := make(chan *api.MergeResult, 10)
    
    go func() {
        defer close(resultCh)
        c.performOptimizedStreamingMerge(ctx, opts, resultCh)
    }()
    
    return resultCh, nil
}

func (c *OptimizedClient) performOptimizedStreamingMerge(ctx context.Context, opts *api.StreamMergeOptions, resultCh chan<- *api.MergeResult) {
    // Process documents in parallel using worker pool
    docCount := len(opts.Documents)
    chunkSize := opts.ChunkSize
    if chunkSize <= 0 {
        chunkSize = 50 // Larger default for optimization
    }
    
    // Channel for parsed documents
    parsedCh := make(chan *ParsedChunk, 10)
    
    // Start parser workers
    var wg sync.WaitGroup
    numWorkers := runtime.NumCPU()
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.parseWorker(ctx, parsedCh)
        }()
    }
    
    // Send parsing jobs
    go func() {
        defer close(parsedCh)
        
        for i := 0; i < docCount; i += chunkSize {
            select {
            case <-ctx.Done():
                return
            default:
            }
            
            end := i + chunkSize
            if end > docCount {
                end = docCount
            }
            
            chunk := &ParsedChunk{
                Documents: opts.Documents[i:end],
                StartIdx:  i,
                EndIdx:    end,
            }
            
            parsedCh <- chunk
        }
    }()
    
    // Collect parsed results
    statementMap := make(map[string]*StatementGroup)
    processedCount := 0
    
    wg.Wait() // Wait for all parsing to complete
    
    // Continue with merging logic...
    c.mergeOptimized(ctx, statementMap, opts, resultCh, docCount)
}

type ParsedChunk struct {
    Documents []string
    StartIdx  int
    EndIdx    int
}

func (c *OptimizedClient) parseWorker(ctx context.Context, chunkCh <-chan *ParsedChunk) {
    for chunk := range chunkCh {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // Parse documents in this chunk
        for _, docStr := range chunk.Documents {
            // Use streaming JSON decoder for large documents
            if err := c.parseDocumentOptimized(docStr); err != nil {
                c.logger.Error("Failed to parse document", "error", err)
                continue
            }
        }
    }
}

func (c *OptimizedClient) parseDocumentOptimized(docStr string) error {
    // Use buffer pool for parsing
    buf := performance.GetBuffer()
    defer performance.PutBuffer(buf)
    
    reader := bytes.NewReader([]byte(docStr))
    decoder := json.NewDecoder(reader)
    
    // Use streaming decoder to avoid loading entire document into memory
    var doc map[string]interface{}
    if err := decoder.Decode(&doc); err != nil {
        return err
    }
    
    // Process document incrementally
    return nil
}

func (c *OptimizedClient) mergeOptimized(ctx context.Context, statementMap map[string]*StatementGroup, opts *api.StreamMergeOptions, resultCh chan<- *api.MergeResult, docCount int) {
    // Optimized merging with parallel processing
    // Implementation details...
}
```

**Code Example - Memory Monitoring:**
```go
// internal/performance/monitor.go
package performance

import (
    "context"
    "runtime"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/internal/logging"
)

type MemoryMonitor struct {
    logger   logging.Logger
    interval time.Duration
}

func NewMemoryMonitor(logger logging.Logger, interval time.Duration) *MemoryMonitor {
    return &MemoryMonitor{
        logger:   logger,
        interval: interval,
    }
}

func (m *MemoryMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.logMemoryStats()
        }
    }
}

func (m *MemoryMonitor) logMemoryStats() {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    m.logger.Info("Memory stats",
        "alloc_mb", memStats.Alloc/1024/1024,
        "sys_mb", memStats.Sys/1024/1024,
        "heap_objects", memStats.HeapObjects,
        "gc_cycles", memStats.NumGC,
    )
    
    // Force GC if memory usage is high
    if memStats.Alloc > 500*1024*1024 { // 500MB threshold
        runtime.GC()
        m.logger.Info("Forced garbage collection due to high memory usage")
    }
}

func (m *MemoryMonitor) GetMemoryUsage() (uint64, uint64) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    return memStats.Alloc, memStats.Sys
}
```

**Deliverable**: Performance optimizations in `internal/performance/`

---

### Task 6.3: Benchmarking and Validation (1 point)

**Goal**: Create comprehensive benchmarks and validate performance improvements

**Checklist:**
- [ ] Create benchmark suite for all major operations
- [ ] Add performance regression tests
- [ ] Compare against Node.js implementation
- [ ] Create load testing scenarios
- [ ] Document performance characteristics

**Code Example - Comprehensive Benchmarks:**
```go
// test/benchmarks/vex_test.go
package benchmarks

import (
    "context"
    "fmt"
    "testing"
    "time"
    
    "github.com/rosstaco/vexdoc-mcp-go/internal/vex"
    "github.com/rosstaco/vexdoc-mcp-go/internal/logging"
    "github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

func BenchmarkVEXCreate(b *testing.B) {
    client := setupVEXClient(b)
    
    opts := &api.CreateOptions{
        Product:       "nginx:1.20",
        Vulnerability: "CVE-2023-1234",
        Status:        "not_affected",
        Justification: "component_not_present",
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := client.CreateStatement(context.Background(), opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkVEXMerge(b *testing.B) {
    client := setupVEXClient(b)
    
    // Generate test documents
    docs := generateTestDocuments(10)
    
    opts := &api.MergeOptions{
        Documents: docs,
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := client.MergeDocuments(context.Background(), opts)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkStreamingMerge(b *testing.B) {
    client := setupVEXClient(b)
    
    testCases := []struct {
        name     string
        docCount int
    }{
        {"10_docs", 10},
        {"100_docs", 100},
        {"1000_docs", 1000},
    }
    
    for _, tc := range testCases {
        b.Run(tc.name, func(b *testing.B) {
            docs := generateTestDocuments(tc.docCount)
            
            opts := &api.StreamMergeOptions{
                MergeOptions: api.MergeOptions{Documents: docs},
                ChunkSize:    50,
            }
            
            b.ResetTimer()
            b.ReportAllocs()
            
            for i := 0; i < b.N; i++ {
                resultCh, err := client.StreamMerge(context.Background(), opts)
                if err != nil {
                    b.Fatal(err)
                }
                
                // Consume all results
                for range resultCh {
                    // Process results
                }
            }
        })
    }
}

func BenchmarkMemoryUsage(b *testing.B) {
    client := setupVEXClient(b)
    
    // Test with large documents
    largeDocuments := generateLargeTestDocuments(100, 100) // 100 docs with 100 statements each
    
    opts := &api.StreamMergeOptions{
        MergeOptions: api.MergeOptions{Documents: largeDocuments},
        ChunkSize:    10,
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        resultCh, err := client.StreamMerge(context.Background(), opts)
        if err != nil {
            b.Fatal(err)
        }
        
        for range resultCh {
            // Memory usage will be measured by test framework
        }
    }
}

func TestPerformanceRegression(t *testing.T) {
    // Baseline performance targets
    targets := map[string]time.Duration{
        "create_statement": 10 * time.Millisecond,
        "merge_10_docs":    100 * time.Millisecond,
        "merge_100_docs":   1 * time.Second,
    }
    
    client := setupVEXClient(t)
    
    // Test create statement performance
    start := time.Now()
    opts := &api.CreateOptions{
        Product:       "test-product",
        Vulnerability: "CVE-2023-1234",
        Status:        "not_affected",
        Justification: "component_not_present",
    }
    
    _, err := client.CreateStatement(context.Background(), opts)
    if err != nil {
        t.Fatal(err)
    }
    
    createDuration := time.Since(start)
    if createDuration > targets["create_statement"] {
        t.Errorf("Create statement too slow: %v > %v", createDuration, targets["create_statement"])
    }
    
    // Test merge performance
    for _, docCount := range []int{10, 100} {
        testName := fmt.Sprintf("merge_%d_docs", docCount)
        
        start = time.Now()
        docs := generateTestDocuments(docCount)
        mergeOpts := &api.MergeOptions{Documents: docs}
        
        _, err = client.MergeDocuments(context.Background(), mergeOpts)
        if err != nil {
            t.Fatal(err)
        }
        
        mergeDuration := time.Since(start)
        if mergeDuration > targets[testName] {
            t.Errorf("Merge %d docs too slow: %v > %v", docCount, mergeDuration, targets[testName])
        }
    }
}

func setupVEXClient(tb testing.TB) api.VEXClient {
    config := &api.VEXConfig{
        DefaultAuthor: "test-author",
        ValidateOn:    false, // Disable validation for benchmarks
    }
    
    logger, _ := logging.New("error", "json", io.Discard)
    return vex.NewClient(config, logger)
}

func generateTestDocuments(count int) []string {
    docs := make([]string, count)
    
    for i := 0; i < count; i++ {
        doc := fmt.Sprintf(`{
            "@id": "test-doc-%d",
            "author": "test-author",
            "version": 1,
            "timestamp": "2023-01-01T00:00:00Z",
            "statements": [{
                "vulnerability": {"name": "CVE-2023-%04d"},
                "products": [{"component": {"@id": "product-%d"}}],
                "status": "not_affected",
                "justification": "component_not_present"
            }]
        }`, i, 1000+i, i)
        
        docs[i] = doc
    }
    
    return docs
}

func generateLargeTestDocuments(docCount, statementsPerDoc int) []string {
    docs := make([]string, docCount)
    
    for i := 0; i < docCount; i++ {
        statements := make([]string, statementsPerDoc)
        
        for j := 0; j < statementsPerDoc; j++ {
            stmt := fmt.Sprintf(`{
                "vulnerability": {"name": "CVE-2023-%04d"},
                "products": [{"component": {"@id": "product-%d-%d"}}],
                "status": "not_affected",
                "justification": "component_not_present"
            }`, 1000+j, i, j)
            statements[j] = stmt
        }
        
        doc := fmt.Sprintf(`{
            "@id": "large-doc-%d",
            "author": "test-author",
            "version": 1,
            "timestamp": "2023-01-01T00:00:00Z",
            "statements": [%s]
        }`, i, strings.Join(statements, ","))
        
        docs[i] = doc
    }
    
    return docs
}
```

**Code Example - Load Testing:**
```bash
#!/bin/bash
# scripts/load_test.sh

echo "Starting Go MCP server..."
go run cmd/server/main.go &
GO_PID=$!

sleep 2

echo "Running load tests..."

# Test concurrent create operations
echo "Testing concurrent VEX creation..."
for i in {1..100}; do
  (
    echo '{"jsonrpc":"2.0","id":'$i',"method":"tools/call","params":{"name":"create_vex_statement","arguments":{"product":"nginx:1.20","vulnerability":"CVE-2023-'$i'","status":"not_affected","justification":"component_not_present"}}}' | \
    curl -X POST -H "Content-Type: application/json" -d @- http://localhost:3000/mcp &
  )
done

wait

echo "Load test completed"

# Clean up
kill $GO_PID
```

**Deliverable**: Comprehensive benchmarking suite in `test/benchmarks/`

---

## Phase 6 Deliverables

### 1. Streaming HTTP Transport (`internal/mcp/streaming_transport.go`)
- [ ] Server-Sent Events implementation
- [ ] WebSocket support for bidirectional streaming
- [ ] Connection management and cleanup
- [ ] Flow control and backpressure handling

### 2. Performance Optimizations (`internal/performance/`)
- [ ] Object pooling for memory efficiency
- [ ] Optimized JSON processing
- [ ] Worker pools for concurrent operations
- [ ] Memory usage monitoring

### 3. Optimized VEX Client (`internal/vex/optimized.go`)
- [ ] Parallel document processing
- [ ] Memory-efficient streaming operations
- [ ] Incremental parsing for large documents
- [ ] CPU-aware worker pool sizing

### 4. Benchmarking Suite (`test/benchmarks/`)
- [ ] Comprehensive performance benchmarks
- [ ] Performance regression tests
- [ ] Load testing scenarios
- [ ] Memory usage validation

### 5. Performance Documentation
- [ ] Performance characteristics documentation
- [ ] Optimization guide
- [ ] Benchmarking results vs Node.js
- [ ] Tuning recommendations

## Success Criteria
- [ ] 50%+ performance improvement over Node.js implementation
- [ ] Memory usage remains stable under load
- [ ] Streaming operations handle 1000+ documents efficiently
- [ ] Sub-10ms response times for simple operations
- [ ] Concurrent operation support without performance degradation

## Performance Targets

| Operation | Target Performance | Memory Limit |
|-----------|-------------------|--------------|
| VEX Create | <10ms | <1MB |
| Merge 10 docs | <100ms | <10MB |
| Merge 100 docs | <1s | <50MB |
| Stream 1000 docs | <10s | <100MB |
| Concurrent requests | 100+ req/s | Stable |

## Dependencies
- **Input**: Phase 5 complete tool implementation
- **Output**: High-performance system ready for production testing

## Risks & Mitigation
- **Risk**: Performance optimizations add complexity
  - **Mitigation**: Extensive benchmarking and gradual optimization
- **Risk**: Memory leaks in streaming operations
  - **Mitigation**: Comprehensive memory monitoring and testing
- **Risk**: Concurrency issues with optimizations
  - **Mitigation**: Race condition testing and careful synchronization

## Time Estimate
**5 Story Points** ≈ 2-3 days of focused development

---
**Next**: [Phase 7: Testing & Validation](./phase7-testing.md)
