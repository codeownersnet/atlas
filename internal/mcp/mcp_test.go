package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	if registry == nil {
		t.Fatal("NewToolRegistry() returned nil")
	}
	if len(registry.tools) != 0 {
		t.Errorf("NewToolRegistry() should start with 0 tools, got %d", len(registry.tools))
	}
}

func TestRegisterTool(t *testing.T) {
	registry := NewToolRegistry()

	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		return NewSuccessResult("test"), nil
	}

	tool := NewTool(
		"test_tool",
		"A test tool",
		NewInputSchema(map[string]Property{
			"param1": NewStringProperty("First parameter"),
		}, "param1"),
		handler,
		"test",
	)

	err := registry.RegisterTool(tool)
	if err != nil {
		t.Fatalf("RegisterTool() error = %v", err)
	}

	// Try to register the same tool again
	err = registry.RegisterTool(tool)
	if err == nil {
		t.Error("RegisterTool() should return error when registering duplicate tool")
	}
}

func TestRegisterToolErrors(t *testing.T) {
	registry := NewToolRegistry()

	tests := []struct {
		name    string
		tool    *ToolDefinition
		wantErr bool
	}{
		{
			name: "empty name",
			tool: &ToolDefinition{
				Tool: Tool{
					Name:        "",
					Description: "test",
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
					return NewSuccessResult("test"), nil
				},
			},
			wantErr: true,
		},
		{
			name: "nil handler",
			tool: &ToolDefinition{
				Tool: Tool{
					Name:        "test",
					Description: "test",
				},
				Handler: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterTool(tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterTool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListTools(t *testing.T) {
	registry := NewToolRegistry()

	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		return NewSuccessResult("test"), nil
	}

	tool1 := NewTool("tool1", "First tool", NewInputSchema(nil), handler, "test")
	tool2 := NewTool("tool2", "Second tool", NewInputSchema(nil), handler, "test")

	registry.RegisterTool(tool1)
	registry.RegisterTool(tool2)

	tools := registry.ListTools()
	if len(tools) != 2 {
		t.Errorf("ListTools() returned %d tools, want 2", len(tools))
	}
}

func TestListToolsFiltered(t *testing.T) {
	registry := NewToolRegistry()

	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		return NewSuccessResult("test"), nil
	}

	readTool := NewTool("read_tool", "Read tool", NewInputSchema(nil), handler, "read")
	writeTool := NewTool("write_tool", "Write tool", NewInputSchema(nil), handler, "write")
	bothTool := NewTool("both_tool", "Both tool", NewInputSchema(nil), handler, "read", "write")

	registry.RegisterTool(readTool)
	registry.RegisterTool(writeTool)
	registry.RegisterTool(bothTool)

	tests := []struct {
		name         string
		enabledTools []string
		readOnlyMode bool
		want         int
	}{
		{
			name:         "all tools",
			enabledTools: []string{},
			readOnlyMode: false,
			want:         3,
		},
		{
			name:         "read-only mode",
			enabledTools: []string{},
			readOnlyMode: true,
			want:         1, // Only read_tool
		},
		{
			name:         "enabled tools filter",
			enabledTools: []string{"read_tool", "write_tool"},
			readOnlyMode: false,
			want:         2,
		},
		{
			name:         "enabled tools filter with read-only",
			enabledTools: []string{"read_tool", "write_tool"},
			readOnlyMode: true,
			want:         1, // Only read_tool
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := registry.ListToolsFiltered(tt.enabledTools, tt.readOnlyMode)
			if len(tools) != tt.want {
				t.Errorf("ListToolsFiltered() returned %d tools, want %d", len(tools), tt.want)
			}
		})
	}
}

func TestCallTool(t *testing.T) {
	registry := NewToolRegistry()

	called := false
	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		called = true
		return NewSuccessResult("success"), nil
	}

	tool := NewTool("test_tool", "Test tool", NewInputSchema(nil), handler, "test")
	registry.RegisterTool(tool)

	ctx := context.Background()
	result, err := registry.CallTool(ctx, "test_tool", nil)
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}

	if !called {
		t.Error("CallTool() did not call the handler")
	}

	if result.IsError {
		t.Error("CallTool() result should not be an error")
	}

	if len(result.Content) != 1 || result.Content[0].Text != "success" {
		t.Errorf("CallTool() result content = %v, want 'success'", result.Content)
	}
}

func TestCallToolNotFound(t *testing.T) {
	registry := NewToolRegistry()
	ctx := context.Background()

	_, err := registry.CallTool(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("CallTool() should return error for nonexistent tool")
	}
}

func TestNewServer(t *testing.T) {
	logger := zerolog.Nop()
	server := NewServer(&ServerConfig{
		Logger:       &logger,
		ReadOnlyMode: false,
	})

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
}

func TestServerHandleInitialize(t *testing.T) {
	logger := zerolog.Nop()
	server := NewServer(&ServerConfig{
		Logger: &logger,
	})

	request := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		}`),
	}

	reqData, _ := json.Marshal(request)
	ctx := context.Background()

	respData, err := server.HandleMessage(ctx, reqData)
	if err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(respData, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Response contains error: %v", response.Error)
	}

	if response.Result == nil {
		t.Error("Response should contain result")
	}
}

func TestServerHandleToolsList(t *testing.T) {
	logger := zerolog.Nop()
	server := NewServer(&ServerConfig{
		Logger: &logger,
	})

	// Register a test tool
	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		return NewSuccessResult("test"), nil
	}
	tool := NewTool("test_tool", "Test tool", NewInputSchema(nil), handler, "test")
	server.RegisterTool(tool)

	request := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	reqData, _ := json.Marshal(request)
	ctx := context.Background()

	respData, err := server.HandleMessage(ctx, reqData)
	if err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(respData, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Response contains error: %v", response.Error)
	}

	if response.Result == nil {
		t.Error("Response should contain result")
	}
}

func TestServerHandleToolsCall(t *testing.T) {
	logger := zerolog.Nop()
	server := NewServer(&ServerConfig{
		Logger: &logger,
	})

	// Register a test tool
	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		return NewSuccessResult("tool executed"), nil
	}
	tool := NewTool("test_tool", "Test tool", NewInputSchema(nil), handler, "test")
	server.RegisterTool(tool)

	request := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "test_tool", "arguments": {}}`),
	}

	reqData, _ := json.Marshal(request)
	ctx := context.Background()

	respData, err := server.HandleMessage(ctx, reqData)
	if err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(respData, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Response contains error: %v", response.Error)
	}

	if response.Result == nil {
		t.Error("Response should contain result")
	}
}

func TestServerReadOnlyMode(t *testing.T) {
	logger := zerolog.Nop()
	server := NewServer(&ServerConfig{
		Logger:       &logger,
		ReadOnlyMode: true,
	})

	// Register a write tool
	handler := func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
		return NewSuccessResult("should not execute"), nil
	}
	tool := NewTool("write_tool", "Write tool", NewInputSchema(nil), handler, "write")
	server.RegisterTool(tool)

	request := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name": "write_tool", "arguments": {}}`),
	}

	reqData, _ := json.Marshal(request)
	ctx := context.Background()

	respData, err := server.HandleMessage(ctx, reqData)
	if err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}

	var response Response
	if err := json.Unmarshal(respData, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error == nil {
		t.Error("Response should contain error in read-only mode")
	}
}

func TestPropertyHelpers(t *testing.T) {
	stringProp := NewStringProperty("test string")
	if stringProp.Type != "string" {
		t.Errorf("NewStringProperty() type = %s, want 'string'", stringProp.Type)
	}

	intProp := NewIntegerProperty("test int")
	if intProp.Type != "integer" {
		t.Errorf("NewIntegerProperty() type = %s, want 'integer'", intProp.Type)
	}

	boolProp := NewBooleanProperty("test bool")
	if boolProp.Type != "boolean" {
		t.Errorf("NewBooleanProperty() type = %s, want 'boolean'", boolProp.Type)
	}

	arrayProp := NewArrayProperty("test array", NewStringProperty("item"))
	if arrayProp.Type != "array" {
		t.Errorf("NewArrayProperty() type = %s, want 'array'", arrayProp.Type)
	}
	if arrayProp.Items == nil {
		t.Error("NewArrayProperty() should have items")
	}

	enumProp := NewEnumProperty("test enum", "value1", "value2", "value3")
	if enumProp.Type != "string" {
		t.Errorf("NewEnumProperty() type = %s, want 'string'", enumProp.Type)
	}
	if len(enumProp.Enum) != 3 {
		t.Errorf("NewEnumProperty() enum length = %d, want 3", len(enumProp.Enum))
	}

	propWithDefault := NewStringProperty("test").WithDefault("default_value")
	if propWithDefault.Default != "default_value" {
		t.Errorf("WithDefault() default = %v, want 'default_value'", propWithDefault.Default)
	}
}

func TestContentHelpers(t *testing.T) {
	textContent := NewTextContent("test text")
	if textContent.Type != "text" {
		t.Errorf("NewTextContent() type = %s, want 'text'", textContent.Type)
	}
	if textContent.Text != "test text" {
		t.Errorf("NewTextContent() text = %s, want 'test text'", textContent.Text)
	}

	successResult := NewSuccessResult("success")
	if successResult.IsError {
		t.Error("NewSuccessResult() should not be an error")
	}
	if len(successResult.Content) != 1 {
		t.Errorf("NewSuccessResult() content length = %d, want 1", len(successResult.Content))
	}
}

func TestMessageTypes(t *testing.T) {
	request := Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "test",
	}
	if !request.IsRequest() {
		t.Error("Message should be identified as request")
	}

	notification := Message{
		JSONRPC: "2.0",
		Method:  "test",
	}
	if !notification.IsNotification() {
		t.Error("Message should be identified as notification")
	}

	response := Message{
		JSONRPC: "2.0",
		ID:      1,
		Result:  "test",
	}
	if !response.IsResponse() {
		t.Error("Message should be identified as response")
	}
}
