package mcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// StdioTransport implements the stdio transport for MCP
type StdioTransport struct {
	server *Server
	reader *bufio.Reader
	writer io.Writer
	logger *zerolog.Logger
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(server *Server, logger *zerolog.Logger) *StdioTransport {
	return &StdioTransport{
		server: server,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		logger: logger,
	}
}

// Start starts the stdio transport loop
func (t *StdioTransport) Start(ctx context.Context) error {
	t.logDebug("starting stdio transport")

	// Channel to receive lines from stdin
	lineChan := make(chan []byte)
	errChan := make(chan error, 1)

	// Start goroutine to read from stdin
	go func() {
		for {
			line, err := t.reader.ReadBytes('\n')
			if err != nil {
				errChan <- err
				return
			}
			lineChan <- line
		}
	}()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			t.logDebug("stdio transport stopping due to context cancellation")
			return ctx.Err()

		case err := <-errChan:
			if err == io.EOF {
				t.logDebug("stdin closed, stopping transport")
				return nil
			}
			t.logError("error reading from stdin", err)
			return err

		case line := <-lineChan:
			if err := t.handleMessage(ctx, line); err != nil {
				t.logError("error handling message", err)
				// Continue processing despite errors
			}
		}
	}
}

// handleMessage processes a single message
func (t *StdioTransport) handleMessage(ctx context.Context, line []byte) error {
	// Skip empty lines
	if len(line) <= 1 {
		return nil
	}

	t.logDebug("received message from stdin")

	// Process the message
	response, err := t.server.HandleMessage(ctx, line)
	if err != nil {
		t.logError("failed to handle message", err)
		return nil // Don't stop the transport on handler errors
	}

	// Send response if we have one (notifications don't have responses)
	if response != nil {
		if err := t.sendResponse(response); err != nil {
			t.logError("failed to send response", err)
			return err
		}
	}

	return nil
}

// sendResponse sends a response to stdout
func (t *StdioTransport) sendResponse(data []byte) error {
	// Write the response followed by a newline
	if _, err := t.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	if _, err := t.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	t.logDebug("sent response to stdout")
	return nil
}

// Logging helpers

func (t *StdioTransport) logDebug(msg string) {
	if t.logger == nil {
		return
	}
	t.logger.Debug().Msg(msg)
}

func (t *StdioTransport) logError(msg string, err error) {
	if t.logger == nil {
		return
	}
	t.logger.Error().Err(err).Msg(msg)
}
