package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/rosstaco/vexdoc-mcp-go/pkg/api"
)

// StdioTransport implements the Transport interface using stdin/stdout
type StdioTransport struct {
	reader *bufio.Scanner
	writer io.Writer
	mu     sync.Mutex
	closed bool
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer size for large messages
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max token size

	return &StdioTransport{
		reader: scanner,
		writer: os.Stdout,
	}
}

// Read reads a request from stdin
func (t *StdioTransport) Read() (*api.Request, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, io.EOF
	}

	if !t.reader.Scan() {
		if err := t.reader.Err(); err != nil {
			return nil, fmt.Errorf("error reading from stdin: %w", err)
		}
		return nil, io.EOF
	}

	line := t.reader.Bytes()
	if len(line) == 0 {
		return nil, fmt.Errorf("empty line received")
	}

	var req api.Request
	if err := json.Unmarshal(line, &req); err != nil {
		return nil, fmt.Errorf("error parsing JSON request: %w", err)
	}

	// Log to stderr for debugging (stdout is reserved for JSON-RPC)
	fmt.Fprintf(os.Stderr, "[DEBUG] Received request: method=%s id=%v\n", req.Method, req.ID)

	return &req, nil
}

// Write writes a response to stdout
func (t *StdioTransport) Write(resp *api.Response) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	// Write JSON followed by newline
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}
	if _, err := t.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("error writing newline: %w", err)
	}

	// Log to stderr for debugging
	fmt.Fprintf(os.Stderr, "[DEBUG] Sent response: id=%v error=%v\n", resp.ID, resp.Error != nil)

	return nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	return nil
}
