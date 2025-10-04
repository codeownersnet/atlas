package mcp

import (
	"context"
	"fmt"
)

// ToolHandler is a function that handles a tool call
type ToolHandler func(ctx context.Context, arguments map[string]interface{}) (*CallToolResult, error)

// ToolRegistry manages tool registration and execution
type ToolRegistry struct {
	tools    map[string]*ToolDefinition
	handlers map[string]ToolHandler
}

// ToolDefinition represents a complete tool definition
type ToolDefinition struct {
	Tool
	Handler ToolHandler
	Tags    []string // For filtering (e.g., "jira", "confluence", "read", "write")
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]*ToolDefinition),
		handlers: make(map[string]ToolHandler),
	}
}

// RegisterTool registers a new tool with its handler
func (r *ToolRegistry) RegisterTool(def *ToolDefinition) error {
	if def.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if def.Handler == nil {
		return fmt.Errorf("tool handler cannot be nil for %s", def.Name)
	}

	if _, exists := r.tools[def.Name]; exists {
		return fmt.Errorf("tool %s is already registered", def.Name)
	}

	r.tools[def.Name] = def
	r.handlers[def.Name] = def.Handler

	return nil
}

// GetTool returns a tool definition by name
func (r *ToolRegistry) GetTool(name string) (*ToolDefinition, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// ListTools returns all registered tools
func (r *ToolRegistry) ListTools() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, def := range r.tools {
		tools = append(tools, def.Tool)
	}
	return tools
}

// ListToolsFiltered returns tools filtered by tags and enabled list
func (r *ToolRegistry) ListToolsFiltered(enabledTools []string, readOnlyMode bool) []Tool {
	tools := make([]Tool, 0)

	// Create a map of enabled tools for quick lookup
	enabledMap := make(map[string]bool)
	if len(enabledTools) > 0 {
		for _, name := range enabledTools {
			enabledMap[name] = true
		}
	}

	for _, def := range r.tools {
		// Check if tool is in enabled list (if list is provided)
		if len(enabledTools) > 0 && !enabledMap[def.Name] {
			continue
		}

		// Filter out write tools in read-only mode
		if readOnlyMode && r.hasWriteTag(def.Tags) {
			continue
		}

		tools = append(tools, def.Tool)
	}

	return tools
}

// CallTool executes a tool by name with the given arguments
func (r *ToolRegistry) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*CallToolResult, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// Execute the tool handler
	result, err := handler(ctx, arguments)
	if err != nil {
		// Return error as a tool result with isError flag
		return &CallToolResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error executing tool: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return result, nil
}

// hasWriteTag checks if a tool has a "write" tag
func (r *ToolRegistry) hasWriteTag(tags []string) bool {
	for _, tag := range tags {
		if tag == "write" {
			return true
		}
	}
	return false
}

// Helper functions for creating tool definitions

// NewTool creates a new tool definition
func NewTool(name, description string, schema InputSchema, handler ToolHandler, tags ...string) *ToolDefinition {
	return &ToolDefinition{
		Tool: Tool{
			Name:        name,
			Description: description,
			InputSchema: schema,
		},
		Handler: handler,
		Tags:    tags,
	}
}

// NewInputSchema creates a new input schema
func NewInputSchema(properties map[string]Property, required ...string) InputSchema {
	// Ensure properties is never nil to avoid "properties": null in JSON
	if properties == nil {
		properties = make(map[string]Property)
	}

	return InputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// NewProperty creates a new property definition
func NewProperty(propType, description string) Property {
	return Property{
		Type:        propType,
		Description: description,
	}
}

// NewStringProperty creates a new string property
func NewStringProperty(description string) Property {
	return Property{
		Type:        "string",
		Description: description,
	}
}

// NewIntegerProperty creates a new integer property
func NewIntegerProperty(description string) Property {
	return Property{
		Type:        "integer",
		Description: description,
	}
}

// NewBooleanProperty creates a new boolean property
func NewBooleanProperty(description string) Property {
	return Property{
		Type:        "boolean",
		Description: description,
	}
}

// NewArrayProperty creates a new array property
func NewArrayProperty(description string, items Property) Property {
	return Property{
		Type:        "array",
		Description: description,
		Items:       &items,
	}
}

// NewEnumProperty creates a new enum property
func NewEnumProperty(description string, values ...string) Property {
	return Property{
		Type:        "string",
		Description: description,
		Enum:        values,
	}
}

// WithDefault adds a default value to a property
func (p Property) WithDefault(value interface{}) Property {
	p.Default = value
	return p
}

// Helper functions for creating tool results

// NewTextContent creates a new text content item
func NewTextContent(text string) Content {
	return Content{
		Type: "text",
		Text: text,
	}
}

// NewSuccessResult creates a successful tool result with text content
func NewSuccessResult(text string) *CallToolResult {
	return &CallToolResult{
		Content: []Content{NewTextContent(text)},
		IsError: false,
	}
}

// NewErrorResult creates an error tool result
func NewErrorResult(err error) *CallToolResult {
	return &CallToolResult{
		Content: []Content{NewTextContent(err.Error())},
		IsError: true,
	}
}

// NewJSONResult creates a tool result with JSON-formatted text
func NewJSONResult(data interface{}) (*CallToolResult, error) {
	jsonBytes, err := marshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result to JSON: %w", err)
	}

	return &CallToolResult{
		Content: []Content{NewTextContent(string(jsonBytes))},
		IsError: false,
	}, nil
}
