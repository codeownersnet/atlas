# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go MCP Atlassian is a high-performance Go implementation of the Model Context Protocol (MCP) server for Atlassian products (Jira and Confluence). It provides 40 tools that enable AI assistants to interact with Atlassian Cloud and Server/Data Center deployments.

## Project Structure

```
go-mcp-atlassian/
├── cmd/atlas-mcp/           # Main application entry point
│   └── main.go              # CLI and server initialization
├── internal/                # Private application code
│   ├── auth/                # Authentication providers (Basic, PAT, Bearer)
│   ├── client/              # HTTP client with retry, proxy, SSL support
│   ├── config/              # Configuration loading and validation
│   ├── mcp/                 # MCP protocol implementation (JSON-RPC 2.0)
│   └── tools/               # MCP tool implementations
│       ├── jira/            # 29 Jira tools (14 read, 15 write)
│       └── confluence/      # 11 Confluence tools (6 read, 5 write)
├── pkg/atlassian/           # Public Atlassian API clients
│   ├── jira/                # Jira REST API client
│   └── confluence/          # Confluence REST API client
└── docs/                    # Documentation
    ├── product.md           # Product specification
    └── implementation-plan.md  # Detailed implementation plan
```

## Common Development Commands

The project uses a Makefile for common tasks. Run `make help` to see all available commands.

### Quick Reference

```bash
# Build and test
make build              # Build binary
make test               # Run all tests
make test-coverage      # Generate coverage report
make check              # Run all quality checks (fmt, vet, lint, test)

# Code quality
make fmt                # Format code
make vet                # Run go vet
make lint               # Run golangci-lint
make tidy               # Tidy go.mod

# Running
make run                # Build and run
make dev                # Run with verbose logging
./bin/mcp-atlassian     # Run built binary

# Cross-platform builds
make build-all          # Build for Linux, macOS, Windows

# Information
make info               # Show project statistics
make size               # Show binary size
make version            # Show version info

# Utilities
make clean              # Remove build artifacts
make install            # Install to $GOPATH/bin
make tools              # Install development tools
```

### Manual Commands (without Make)

```bash
# Build
go build -o atlas-mcp ./cmd/atlas-mcp

# Test
go test ./...
go test -cover ./...
go test -coverprofile=coverage.out ./...

# Quality
go fmt ./...
go vet ./...
golangci-lint run

# Dependencies
go mod tidy
go get -u ./...
```

## Architecture

### Authentication Flow
1. Configuration loaded from environment variables or .env file
2. Auth provider created based on config (Basic Auth for Cloud, PAT for Server/DC, Bearer for BYO OAuth)
3. HTTP client created with auth provider
4. API clients (Jira/Confluence) created with HTTP client
5. Clients stored in context for tool access

### Tool Execution Flow
1. MCP client sends JSON-RPC request via stdio
2. Server deserializes request and routes to tool handler
3. Tool handler extracts parameters from args map
4. Tool retrieves API client from context
5. API client makes authenticated HTTP request to Atlassian
6. Response deserialized and returned as JSON to MCP client

### Key Design Patterns
- **Dependency Injection**: Clients passed via context, not globals
- **Interface-based Auth**: `auth.Provider` interface for multiple auth types
- **Builder Pattern**: Options structs for API calls (e.g., `SearchOptions`, `GetIssueOptions`)
- **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
- **Table-driven Tests**: All test files use table-driven test patterns

## Important Implementation Details

### Jira API Client (`pkg/atlassian/jira/`)
- Auto-detects Cloud vs Server/DC from URL pattern
- Field filtering supports "essential", "*all", or comma-separated field list
- Epic linking handled differently for Cloud vs Server (different field IDs)
- Pagination uses `startAt` and `maxResults` parameters
- JQL queries must be URL-encoded
- Time tracking uses Jira time format: "2h 30m", "1d", "3w"

### Confluence API Client (`pkg/atlassian/confluence/`)
- Auto-detects Cloud vs Server/DC from URL pattern
- Search supports both CQL and plain text (client auto-detects)
- Content formats: storage (default), view, wiki, markdown
- Pagination uses `start` and `limit` parameters
- Space filtering applied via CQL: `space in (KEY1,KEY2)`
- Version management required for updates (optimistic locking)

### MCP Protocol (`internal/mcp/`)
- Implements JSON-RPC 2.0 specification
- Tools registered with name, description, input schema, and handler function
- Handler signature: `func(context.Context, map[string]interface{}) (*CallToolResult, error)`
- Tool filtering by tags: `{jira}`, `{confluence}`, `{read}`, `{write}`
- Three-level filtering: service config → read-only mode → enabled tools list

### Configuration (`internal/config/`)
- Loads from environment variables or .env file
- CLI flags override environment variables
- Validation ensures required fields present
- Auth detection: Bearer if OAuth token set, PAT if personal token set, otherwise Basic Auth
- Proxy config: service-specific overrides global settings
- Custom headers: comma-separated key=value pairs

## Testing Guidelines

### Unit Tests
- Test files colocated with implementation: `foo.go` → `foo_test.go`
- Use table-driven tests for multiple cases
- Mock external dependencies (HTTP responses, API clients)
- Test both success and error cases
- Aim for >80% coverage for new code

### Integration Tests
- Place in `tests/integration/` directory (if created)
- Use environment variables for real API endpoints
- Skip if credentials not provided: `t.Skip("integration test requires credentials")`
- Clean up created resources in test teardown

### Test Helpers
```go
// Common test patterns
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        {"success case", input1, expected1, false},
        {"error case", input2, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Adding New Tools

### 1. Add Tool Implementation
Create tool handler function in appropriate file:
- Jira read tools: `internal/tools/jira/jira_read.go`
- Jira write tools: `internal/tools/jira/jira_write.go`
- Confluence read tools: `internal/tools/confluence/confluence_read.go`
- Confluence write tools: `internal/tools/confluence/confluence_write.go`

```go
func jiraExampleTool(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
    client := GetJiraClient(ctx)
    if client == nil {
        return mcp.NewToolResultError("Jira client not available"), nil
    }

    // Extract and validate parameters
    param, ok := args["param_name"].(string)
    if !ok || param == "" {
        return mcp.NewToolResultError("param_name is required"), nil
    }

    // Call API
    result, err := client.SomeMethod(ctx, param)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("operation failed: %v", err)), nil
    }

    // Return response
    return mcp.NewToolResultText(toJSON(result)), nil
}
```

### 2. Create Tool Schema
Add schema function in `internal/tools/jira/tools.go` or `internal/tools/confluence/tools.go`:

```go
func createExampleTool() *mcp.Tool {
    return mcp.NewTool(
        "jira_example_tool",
        "Description of what this tool does",
        map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "param_name": map[string]interface{}{
                    "type": "string",
                    "description": "Parameter description",
                },
            },
            "required": []string{"param_name"},
        },
        jiraExampleTool,
    )
}
```

### 3. Register Tool
Add registration call in `RegisterJiraTools()` or `RegisterConfluenceTools()`:

```go
func RegisterJiraTools(server *mcp.Server) {
    // ... existing tools ...
    server.RegisterTool(createExampleTool())
}
```

### 4. Tag Tool Appropriately
Tools are automatically tagged based on naming:
- Service: `{jira}` or `{confluence}` (from package)
- Operation: `{read}` or `{write}` (inferred from function name or file location)

## Adding New API Methods

### 1. Define Types
Add types in `pkg/atlassian/jira/types.go` or `pkg/atlassian/confluence/types.go`:

```go
type NewResource struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type CreateNewResourceRequest struct {
    Name string `json:"name"`
}
```

### 2. Add Client Method
Add method in appropriate client file (`pkg/atlassian/jira/` or `pkg/atlassian/confluence/`):

```go
func (c *Client) CreateNewResource(ctx context.Context, req *CreateNewResourceRequest) (*NewResource, error) {
    endpoint := "/rest/api/2/newresource"

    var result NewResource
    err := c.httpClient.Post(ctx, endpoint, req, &result)
    if err != nil {
        return nil, fmt.Errorf("create new resource: %w", err)
    }

    return &result, nil
}
```

### 3. Add Tests
Add test in `pkg/atlassian/jira/client_test.go` or similar:

```go
func TestCreateNewResource(t *testing.T) {
    // Test implementation
}
```

## Code Style Guidelines

1. **Naming Conventions:**
   - Packages: lowercase, single word (e.g., `jira`, `config`)
   - Exported: PascalCase (e.g., `Client`, `GetIssue`)
   - Unexported: camelCase (e.g., `httpClient`, `parseFields`)

2. **Error Handling:**
   - Always check errors: `if err != nil { return err }`
   - Wrap errors with context: `fmt.Errorf("operation: %w", err)`
   - Return errors, don't panic (except in init functions)

3. **Comments:**
   - Document all exported functions, types, and constants
   - Use complete sentences
   - Start with the name of the thing being documented

4. **Logging:**
   - Use zerolog for structured logging
   - Include context: `log.Info().Str("issue_key", key).Msg("fetching issue")`
   - Mask sensitive data: tokens, passwords, header values

5. **Context:**
   - First parameter is always `context.Context`
   - Pass context through all function calls
   - Use `context.WithValue()` for request-scoped data (clients)

## Troubleshooting

### Build Errors
- Run `go mod tidy` to sync dependencies
- Check Go version: requires 1.21+
- Verify all imports are correct

### Test Failures
- Ensure .env.example is not being used (tests use mocks)
- Check for race conditions: `go test -race ./...`
- Verify mocks are properly configured

### Runtime Issues
- Check configuration is loaded: enable `MCP_VERY_VERBOSE=true`
- Verify auth credentials are correct
- Test API connectivity with curl before debugging code
- Check proxy settings if behind firewall

## Feature Status

✅ **Completed:**
- Core MCP protocol implementation
- Jira and Confluence API clients
- 40 tools (29 Jira + 11 Confluence)
- Basic Auth, PAT, and Bearer Token (BYO OAuth) authentication
- Configuration management
- Proxy support
- SSL verification control
- Read-only mode
- Tool filtering

🚧 **In Progress / Planned:**
- HTTP transports (SSE, streamable-http)
- Multi-user support
- Content conversion (Markdown ↔ Confluence Storage)
- Docker images
- Homebrew formula

## Useful Resources

- [MCP Specification](https://modelcontextprotocol.io/)
- [Jira REST API Docs](https://developer.atlassian.com/cloud/jira/platform/rest/v2/)
- [Confluence REST API Docs](https://developer.atlassian.com/cloud/confluence/rest/v2/)
- [Go Documentation](https://go.dev/doc/)
- [Product Specification](docs/product.md)
- [Implementation Plan](docs/implementation-plan.md)
