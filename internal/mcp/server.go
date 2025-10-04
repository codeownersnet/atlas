package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
)

const (
	ProtocolVersion = "2024-11-05"
	ServerName      = "go-mcp-atlassian"
	ServerVersion   = "0.1.0"
)

// Server represents the MCP server
type Server struct {
	registry     *ToolRegistry
	logger       *zerolog.Logger
	readOnlyMode bool
	enabledTools []string
}

// ServerConfig holds the configuration for the MCP server
type ServerConfig struct {
	Logger       *zerolog.Logger
	ReadOnlyMode bool
	EnabledTools []string
}

// NewServer creates a new MCP server
func NewServer(cfg *ServerConfig) *Server {
	return &Server{
		registry:     NewToolRegistry(),
		logger:       cfg.Logger,
		readOnlyMode: cfg.ReadOnlyMode,
		enabledTools: cfg.EnabledTools,
	}
}

// RegisterTool registers a new tool
func (s *Server) RegisterTool(def *ToolDefinition) error {
	return s.registry.RegisterTool(def)
}

// HandleMessage handles an incoming JSON-RPC message
func (s *Server) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	// Parse the message
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		s.logError("failed to parse message", err)
		response := NewErrorResponse(nil, ParseError, "Parse error", err.Error())
		return json.Marshal(response)
	}

	s.logDebug("received message", map[string]interface{}{
		"method": msg.Method,
		"id":     msg.ID,
	})

	// Handle different message types
	if msg.IsRequest() {
		return s.handleRequest(ctx, msg.ToRequest())
	} else if msg.IsNotification() {
		return s.handleNotification(ctx, msg.ToNotification())
	} else if msg.IsResponse() {
		// Responses are not expected in this server (we're not making requests)
		s.logDebug("ignoring response message", nil)
		return nil, nil
	}

	// Unknown message type
	response := NewErrorResponse(msg.ID, InvalidRequest, "Invalid request", nil)
	return json.Marshal(response)
}

// handleRequest handles a JSON-RPC request
func (s *Server) handleRequest(ctx context.Context, req *Request) ([]byte, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req)
	case "tools/list":
		return s.handleToolsList(ctx, req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	default:
		response := NewErrorResponse(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
		return json.Marshal(response)
	}
}

// handleNotification handles a JSON-RPC notification
func (s *Server) handleNotification(ctx context.Context, notif *Notification) ([]byte, error) {
	switch notif.Method {
	case "initialized":
		s.logDebug("client initialized", nil)
		return nil, nil
	case "notifications/cancelled":
		s.logDebug("notification cancelled", nil)
		return nil, nil
	default:
		s.logDebug("unknown notification", map[string]interface{}{
			"method": notif.Method,
		})
		return nil, nil
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(ctx context.Context, req *Request) ([]byte, error) {
	var params InitializeParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			response := NewErrorResponse(req.ID, InvalidParams, "Invalid parameters", err.Error())
			return json.Marshal(response)
		}
	}

	s.logDebug("initialize request", map[string]interface{}{
		"protocol_version": params.ProtocolVersion,
		"client_name":      params.ClientInfo.Name,
		"client_version":   params.ClientInfo.Version,
	})

	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
		Instructions: "MCP server for Atlassian products (Jira and Confluence). Use the available tools to interact with Jira issues and Confluence pages.",
	}

	response := NewResponse(req.ID, result)
	return json.Marshal(response)
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(ctx context.Context, req *Request) ([]byte, error) {
	var params ListToolsParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			response := NewErrorResponse(req.ID, InvalidParams, "Invalid parameters", err.Error())
			return json.Marshal(response)
		}
	}

	s.logDebug("tools/list request", nil)

	// Get filtered tools based on configuration
	tools := s.registry.ListToolsFiltered(s.enabledTools, s.readOnlyMode)

	result := ListToolsResult{
		Tools: tools,
	}

	s.logDebug("returning tools", map[string]interface{}{
		"count": len(tools),
	})

	response := NewResponse(req.ID, result)
	return json.Marshal(response)
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(ctx context.Context, req *Request) ([]byte, error) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		response := NewErrorResponse(req.ID, InvalidParams, "Invalid parameters", err.Error())
		return json.Marshal(response)
	}

	s.logDebug("tools/call request", map[string]interface{}{
		"tool": params.Name,
	})

	// Check if tool exists
	if _, ok := s.registry.GetTool(params.Name); !ok {
		response := NewErrorResponse(req.ID, MethodNotFound, fmt.Sprintf("Tool not found: %s", params.Name), nil)
		return json.Marshal(response)
	}

	// Check if tool is allowed in read-only mode
	if s.readOnlyMode {
		tool, _ := s.registry.GetTool(params.Name)
		if s.registry.hasWriteTag(tool.Tags) {
			response := NewErrorResponse(req.ID, InvalidRequest, "Write operations are disabled in read-only mode", nil)
			return json.Marshal(response)
		}
	}

	// Execute the tool
	result, err := s.registry.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		s.logError("tool execution failed", err)
		response := NewErrorResponse(req.ID, InternalError, "Tool execution failed", err.Error())
		return json.Marshal(response)
	}

	response := NewResponse(req.ID, result)
	return json.Marshal(response)
}

// Logging helpers

func (s *Server) logDebug(msg string, fields map[string]interface{}) {
	if s.logger == nil {
		return
	}

	event := s.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (s *Server) logError(msg string, err error) {
	if s.logger == nil {
		return
	}

	s.logger.Error().Err(err).Msg(msg)
}

// Helper function for JSON marshaling
func marshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
