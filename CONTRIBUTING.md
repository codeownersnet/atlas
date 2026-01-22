# Contributing to Go MCP Atlassian

Thank you for considering contributing to this project. This document provides guidance for developers who want to contribute code, report issues, or improve documentation.

## Prerequisites

- **Go 1.25 or higher** - Required to build and run the project
- **Git** - For cloning and managing the repository
- **Make** - For running common build and test commands (optional, manual commands also work)
- **Access to Atlassian instances** - Helpful for testing (Jira, Confluence, Opsgenie)

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/codeownersnet/atlas.git
cd atlas
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build the Project

```bash
make build
# or manually:
go build -o atlas-mcp ./cmd/atlas-mcp
```

## Development Commands

The project uses a Makefile for common development tasks. Run `make help` to see all available commands.

### Build Commands

```bash
make build              # Build binary for current platform
make build-all          # Build for Linux, macOS, and Windows
make clean              # Remove build artifacts
make install            # Install to $GOPATH/bin
```

### Testing Commands

```bash
make test               # Run all tests
make test-coverage      # Run tests with coverage report
```

### Code Quality Commands

```bash
make fmt                # Format code with go fmt
make vet                # Run go vet
make lint               # Run golangci-lint
make tidy               # Tidy go.mod
make check              # Run all quality checks (fmt, vet, lint, test)
```

### Running in Development

```bash
make run                # Build and run
make dev                # Run with verbose logging for debugging
```

### Information Commands

```bash
make info               # Show project statistics
make size               # Show binary size
make version            # Show version info
make tools              # Install development tools (golangci-lint)
```

### Manual Commands (without Make)

If you prefer not to use Make:

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

## Project Structure

```
atlas/
├── cmd/atlas-mcp/           # Main application entry point
│   └── main.go              # CLI and server initialization
├── internal/                # Private application code
│   ├── auth/                # Authentication providers (Basic, PAT, Bearer)
│   ├── client/              # HTTP client with retry, proxy, SSL support
│   ├── config/              # Configuration loading and validation
│   ├── mcp/                 # MCP protocol implementation (JSON-RPC 2.0)
│   └── tools/               # MCP tool implementations
│       ├── jira/            # 29 Jira tools (14 read, 15 write)
│       ├── confluence/      # 11 Confluence tools (6 read, 5 write)
│       └── opsgenie/        # 25 Opsgenie tools (13 read, 12 write)
├── pkg/atlassian/           # Public Atlassian API clients
│   ├── jira/                # Jira REST API client
│   ├── confluence/          # Confluence REST API client
│   └── opsgenie/            # Opsgenie REST API client
└── docs/                    # Documentation
    ├── vscode/              # VS Code integration guides
    ├── product.md           # Product specification
    └── implementation-plan.md  # Detailed implementation plan
```

## Architecture

### Authentication Flow

1. Configuration loaded from environment variables or `.env` file
2. Auth provider created based on config:
   - **Basic Auth** for Cloud (API token + username)
   - **PAT** for Server/DC (Personal Access Token)
   - **Bearer Token** for BYO OAuth
3. HTTP client created with auth provider
4. API clients (Jira/Confluence/Opsgenie) created with HTTP client
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

## Code Style Guidelines

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `jira`, `config`)
- **Exported**: PascalCase (e.g., `Client`, `GetIssue`)
- **Unexported**: camelCase (e.g., `httpClient`, `parseFields`)

### Error Handling

- Always check errors: `if err != nil { return err }`
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Return errors, don't panic (except in init functions)

### Comments

- Document all exported functions, types, and constants
- Use complete sentences
- Start with the name of the thing being documented

### Logging

- Use zerolog for structured logging
- Include context: `log.Info().Str("issue_key", key).Msg("fetching issue")`
- Mask sensitive data: tokens, passwords, header values

### Context

- First parameter is always `context.Context`
- Pass context through all function calls
- Use `context.WithValue()` for request-scoped data (clients)

## Testing Guidelines

### Unit Tests

- Test files should be colocated with implementation: `foo.go` → `foo_test.go`
- Use table-driven tests for multiple cases
- Mock external dependencies (HTTP responses, API clients)
- Test both success and error cases
- Aim for >80% coverage for new code

### Table-Driven Test Pattern

```go
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

### Integration Tests

- Place in `tests/integration/` directory
- Use environment variables for real API endpoints
- Skip if credentials not provided: `t.Skip("integration test requires credentials")`
- Clean up created resources in test teardown

## Adding New Tools

### 1. Add Tool Implementation

Create tool handler function in appropriate file:
- Jira read tools: `internal/tools/jira/jira_read.go`
- Jira write tools: `internal/tools/jira/jira_write.go`
- Confluence read tools: `internal/tools/confluence/confluence_read.go`
- Confluence write tools: `internal/tools/confluence/confluence_write.go`
- Opsgenie read tools: `internal/tools/opsgenie/opsgenie_read.go`
- Opsgenie write tools: `internal/tools/opsgenie/opsgenie_write.go`

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

Add schema function in `internal/tools/jira/tools.go`, `internal/tools/confluence/tools.go`, or `internal/tools/opsgenie/tools.go`:

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

Add registration call in `RegisterJiraTools()`, `RegisterConfluenceTools()`, or `RegisterOpsgenieTools()`:

```go
func RegisterJiraTools(server *mcp.Server) {
    // ... existing tools ...
    server.RegisterTool(createExampleTool())
}
```

### 4. Tag Tool Appropriately

Tools are automatically tagged based on naming:
- Service: `{jira}`, `{confluence}`, or `{opsgenie}` (from package)
- Operation: `{read}` or `{write}` (inferred from function name or file location)

## Adding New API Methods

### 1. Define Types

Add types in `pkg/atlassian/jira/types.go`, `pkg/atlassian/confluence/types.go`, or `pkg/atlassian/opsgenie/types.go`:

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

Add method in appropriate client file (`pkg/atlassian/jira/`, `pkg/atlassian/confluence/`, or `pkg/atlassian/opsgenie/`):

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

Add test in `pkg/atlassian/jira/client_test.go`, `pkg/atlassian/confluence/client_test.go`, or `pkg/atlassian/opsgenie/client_test.go`:

```go
func TestCreateNewResource(t *testing.T) {
    // Test implementation
}
```

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

### Opsgenie API Client (`pkg/atlassian/opsgenie/`)

- Cloud only (Server/Data Center not supported by Opsgenie)
- Authentication uses API key header
- API versioning via URL path (`/v2/`)
- Pagination uses `limit` and `offset` parameters
- Async operations require checking request status

### MCP Protocol (`internal/mcp/`)

- Implements JSON-RPC 2.0 specification
- Tools registered with name, description, input schema, and handler function
- Handler signature: `func(context.Context, map[string]interface{}) (*CallToolResult, error)`
- Tool filtering by tags: `{jira}`, `{confluence}`, `{opsgenie}`, `{read}`, `{write}`
- Three-level filtering: service config → read-only mode → enabled tools list

### Configuration (`internal/config/`)

- Loads from environment variables or `.env` file
- CLI flags override environment variables
- Validation ensures required fields present
- Auth detection: Bearer if OAuth token set, PAT if personal token set, otherwise Basic Auth
- Proxy config: service-specific overrides global settings
- Custom headers: comma-separated key=value pairs

## Pull Request Process

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Make Your Changes

- Write clean, well-documented code
- Add tests for new functionality
- Ensure all tests pass: `make check`
- Update documentation if needed

### 3. Commit Your Changes

Use descriptive commit messages:

```bash
git add .
git commit -m "feat: add new Jira API method for custom fields"
# or
git commit -m "fix: resolve timeout issue with Confluence search"
```

### 4. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub with:
- Clear description of changes
- Related issues (if any)
- Testing instructions
- Screenshots if applicable

### 5. Address Review Feedback

- Respond to review comments promptly
- Make requested changes
- Keep the PR focused and small if possible

## Troubleshooting

### Build Errors

```bash
# Sync dependencies
go mod tidy

# Check Go version
go version  # requires 1.25+

# Verify imports
go vet ./...
```

### Test Failures

```bash
# Ensure .env.example is not being used (tests use mocks)
# Check for race conditions
go test -race ./...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Runtime Issues

```bash
# Enable verbose logging
export MCP_VERY_VERBOSE=true

# Verify configuration is loaded
./atlas-mcp

# Test API connectivity with curl before debugging code
curl -u "email:token" https://your-domain.atlassian.net/rest/api/2/myself

# Check proxy settings if behind firewall
env | grep -i proxy
```

### Common Issues

**Import errors:** Run `go mod tidy` to clean up dependencies

**Linting failures:** Run `make lint` and fix reported issues

**Test flakiness:** Check for race conditions with `go test -race ./...`

## Useful Resources

- [MCP Specification](https://modelcontextprotocol.io/)
- [Jira REST API Docs](https://developer.atlassian.com/cloud/jira/platform/rest/v2/)
- [Confluence REST API Docs](https://developer.atlassian.com/cloud/confluence/rest/v2/)
- [Opsgenie API Docs](https://docs.opsgenie.com/docs/api-overview)
- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Product Specification](docs/product.md)
- [Implementation Plan](docs/implementation-plan.md)

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.
