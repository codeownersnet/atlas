# Go MCP Atlassian

A high-performance Go implementation of the Model Context Protocol (MCP) server for Atlassian products (Jira, Confluence, and Opsgenie). This server enables AI assistants like Claude to seamlessly interact with your Atlassian instances.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Features

- **65 Tools Total**: 29 Jira tools + 11 Confluence tools + 25 Opsgenie tools
- **Multi-Platform Support**: Cloud and Server/Data Center deployments
- **Multiple Auth Methods**: API Token, Personal Access Token, Bearer Token (BYO OAuth)
- **Production Ready**: Built-in retry logic, proxy support, SSL verification
- **Flexible Configuration**: Environment variables, .env files, CLI flags
- **Security**: Read-only mode, tool filtering, credential masking
- **High Performance**: Native Go implementation with efficient HTTP client

## Supported Platforms

| Product | Deployment | Status |
|---------|-----------|--------|
| **Jira** | Cloud | ✅ Supported |
| **Jira** | Server/Data Center (8.14+) | ✅ Supported |
| **Confluence** | Cloud | ✅ Supported |
| **Confluence** | Server/Data Center (6.0+) | ✅ Supported |
| **Opsgenie** | Cloud | ✅ Supported |

## Quick Start

### 1. Installation

#### Option A: Homebrew (macOS/Linux)
```bash
brew tap codeownersnet/atlas
brew install atlas-mcp
```

#### Option B: Download Binary
Download the latest release for your platform from [GitHub Releases](https://github.com/codeownersnet/atlas/releases).

#### Option C: Using Go Install
```bash
go install github.com/codeownersnet/atlas/cmd/atlas-mcp@latest
```

#### Option D: Build from Source
```bash
git clone https://github.com/codeownersnet/atlas
cd atlas
go build -o atlas-mcp ./cmd/atlas-mcp
```

### 2. Configuration

Create a `.env` file:

```bash
# For Jira Cloud
JIRA_URL=https://your-domain.atlassian.net
JIRA_USERNAME=your.email@example.com
JIRA_API_TOKEN=your_jira_api_token

# For Confluence Cloud
CONFLUENCE_URL=https://your-domain.atlassian.net/wiki
CONFLUENCE_USERNAME=your.email@example.com
CONFLUENCE_API_TOKEN=your_confluence_api_token

# For Opsgenie
OPSGENIE_API_KEY=your_opsgenie_api_key
```

<details>
<summary>Server/Data Center Configuration</summary>

```bash
# For Jira Server/Data Center
JIRA_URL=https://jira.your-company.com
JIRA_PERSONAL_TOKEN=your_personal_access_token
JIRA_SSL_VERIFY=true

# For Confluence Server/Data Center
CONFLUENCE_URL=https://confluence.your-company.com
CONFLUENCE_PERSONAL_TOKEN=your_personal_access_token
CONFLUENCE_SSL_VERIFY=true

# For Opsgenie (Cloud only)
OPSGENIE_API_KEY=your_opsgenie_api_key
```
</details>

### 3. Authentication

#### Option A: API Tokens (Cloud - Jira/Confluence)
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Copy the token immediately
4. Set in `.env` file

#### Option B: API Key (Opsgenie)
1. Go to Opsgenie → Settings → API Key Management
2. Create a new API key
3. Copy the key immediately
4. Set in `.env` file

#### Option C: Personal Access Token (Server/Data Center)
1. Go to your profile → Personal Access Tokens
2. Create a new token
3. Copy the token immediately
4. Set in `.env` file

#### Option D: Bearer Token (BYO OAuth - Advanced)
If you manage OAuth tokens externally:

```bash
# Provide your own OAuth access token
ATLASSIAN_OAUTH_ACCESS_TOKEN=your_oauth_access_token
ATLASSIAN_OAUTH_CLOUD_ID=your_cloud_id
```

**Note:** You are responsible for token refresh and management.

### 4. Run the Server

```bash
./atlas-mcp
```

### 5. IDE Integration

#### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "atlas": {
      "command": "/path/to/atlas-mcp",
      "args": [],
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_USERNAME": "your.email@example.com",
        "JIRA_API_TOKEN": "your_api_token",
        "CONFLUENCE_URL": "https://your-domain.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "your.email@example.com",
        "CONFLUENCE_API_TOKEN": "your_api_token",
        "OPSGENIE_API_KEY": "your_opsgenie_api_key"
      }
    }
  }
}
```

#### Cursor

Open Settings → MCP → Add new global MCP server and configure similarly.

## Available Tools

### Jira Tools (29 total)

#### Read Operations (14 tools)
- `jira_get_issue` - Get issue details with field filtering
- `jira_search` - Search issues using JQL
- `jira_search_fields` - Search for field names (including custom fields)
- `jira_get_all_projects` - List all accessible projects
- `jira_get_project_issues` - Get all issues in a project
- `jira_get_project_versions` - Get fix versions for a project
- `jira_get_transitions` - Get available status transitions
- `jira_get_worklog` - Get worklog entries
- `jira_get_agile_boards` - Get agile boards (Scrum/Kanban)
- `jira_get_board_issues` - Get issues on a board
- `jira_get_sprints_from_board` - Get sprints from a board
- `jira_get_sprint_issues` - Get issues in a sprint
- `jira_get_issue_link_types` - Get available link types
- `jira_get_user_profile` - Get user information

#### Write Operations (15 tools)
- `jira_create_issue` - Create new issues
- `jira_update_issue` - Update existing issues
- `jira_delete_issue` - Delete issues
- `jira_add_comment` - Add comments to issues
- `jira_transition_issue` - Change issue status
- `jira_add_worklog` - Log time spent
- `jira_link_to_epic` - Link issues to Epics
- `jira_create_issue_link` - Link issues together
- `jira_create_remote_issue_link` - Create external links
- `jira_remove_issue_link` - Remove issue links
- `jira_create_sprint` - Create new sprints
- `jira_update_sprint` - Update sprint details
- `jira_create_version` - Create fix versions
- `jira_batch_create_issues` - Create multiple issues at once
- `jira_batch_create_versions` - Create multiple versions at once

### Confluence Tools (11 total)

#### Read Operations (6 tools)
- `confluence_search` - Search content using CQL or text
- `confluence_get_page` - Get page content by ID or title+space
- `confluence_get_page_children` - Get child pages
- `confluence_get_comments` - Get page comments
- `confluence_get_labels` - Get page labels
- `confluence_search_user` - Search for users

#### Write Operations (5 tools)
- `confluence_create_page` - Create new pages
- `confluence_update_page` - Update existing pages
- `confluence_delete_page` - Delete pages
- `confluence_add_label` - Add labels to pages
- `confluence_add_comment` - Add comments to pages

### Opsgenie Tools (25 total)

#### Read Operations (13 tools)
- `opsgenie_get_alert` - Get alert details
- `opsgenie_list_alerts` - List alerts with filtering
- `opsgenie_count_alerts` - Count alerts matching query
- `opsgenie_get_request_status` - Get async request status
- `opsgenie_get_incident` - Get incident details
- `opsgenie_list_incidents` - List incidents with filtering
- `opsgenie_get_schedule` - Get schedule details
- `opsgenie_list_schedules` - List all schedules
- `opsgenie_get_schedule_timeline` - Get schedule timeline
- `opsgenie_get_on_calls` - Get current on-call information
- `opsgenie_get_team` - Get team details
- `opsgenie_list_teams` - List all teams
- `opsgenie_get_user` - Get user information

#### Write Operations (12 tools)
- `opsgenie_create_alert` - Create new alerts
- `opsgenie_close_alert` - Close alerts
- `opsgenie_acknowledge_alert` - Acknowledge alerts
- `opsgenie_snooze_alert` - Snooze alerts
- `opsgenie_escalate_alert` - Escalate alerts
- `opsgenie_assign_alert` - Assign alerts to users/teams
- `opsgenie_add_note_to_alert` - Add notes to alerts
- `opsgenie_add_tags_to_alert` - Add tags to alerts
- `opsgenie_create_incident` - Create new incidents
- `opsgenie_close_incident` - Close incidents
- `opsgenie_add_note_to_incident` - Add notes to incidents
- `opsgenie_add_responder_to_incident` - Add responders to incidents

## Configuration Options

### Security & Access Control

```bash
# Run in read-only mode (disables all write operations)
READ_ONLY_MODE=true

# Only enable specific tools (comma-separated)
ENABLED_TOOLS=jira_get_issue,jira_search,confluence_search,opsgenie_list_alerts

# Filter to specific projects/spaces
JIRA_PROJECTS_FILTER=PROJ1,PROJ2
CONFLUENCE_SPACES_FILTER=SPACE1,SPACE2

# Disable specific services
OPSGENIE_ENABLED=false
```

### Logging

```bash
# Enable verbose logging
MCP_VERBOSE=true

# Enable debug logging
MCP_VERY_VERBOSE=true

# Log to stdout instead of stderr
MCP_LOGGING_STDOUT=true
```

### Proxy Configuration

```bash
# Global proxy settings
HTTP_PROXY=http://proxy.example.com:8080
HTTPS_PROXY=http://proxy.example.com:8080
NO_PROXY=localhost,127.0.0.1

# Service-specific proxy (overrides global)
JIRA_HTTP_PROXY=http://jira-proxy.example.com:8080
CONFLUENCE_HTTPS_PROXY=http://confluence-proxy.example.com:8080
OPSGENIE_HTTP_PROXY=http://opsgenie-proxy.example.com:8080
```

### Custom Headers

```bash
# Add custom headers to API requests
JIRA_CUSTOM_HEADERS=X-Custom-Header=value1,X-Another=value2
CONFLUENCE_CUSTOM_HEADERS=X-Custom-Header=value1
OPSGENIE_CUSTOM_HEADERS=X-Custom-Header=value1
```

## Use Cases

### AI-Powered Jira Management
```
User: "Update Jira from our meeting notes"
AI: [Creates/updates issues based on meeting discussion]
```

### Smart Confluence Search
```
User: "Find our OKR guide in Confluence and summarize it"
AI: [Searches Confluence, retrieves content, provides summary]
```

### Workflow Automation
```
User: "Show me all urgent bugs in PROJ assigned to me"
AI: [Uses JQL search to filter and display relevant issues]
```

### Documentation Creation
```
User: "Create a tech design doc for the new feature"
AI: [Creates structured Confluence page with proper formatting]
```

### Incident Management
```
User: "Show me all critical alerts in Opsgenie and create an incident"
AI: [Lists critical alerts and creates incident with proper responders]
```

## Development

### Prerequisites
- Go 1.25 or higher
- Access to Jira, Confluence, and/or Opsgenie instance
- API tokens or Personal Access Tokens

### Using Make

The project includes a comprehensive Makefile for common development tasks:

```bash
# See all available commands
make help

# Build the binary
make build

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Run linter (requires golangci-lint)
make lint

# Format code
make fmt

# Run all quality checks
make check

# Clean build artifacts
make clean

# Install to $GOPATH/bin
make install

# Build for all platforms
make build-all

# Run in development mode with verbose logging
make dev

# Show project info and statistics
make info
```

### Manual Build Commands

If you prefer not to use Make:

```bash
# Build binary
go build -o atlas-mcp ./cmd/atlas-mcp

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Install development tools
make tools
```

### Project Structure

```
atlas/
├── cmd/atlas-mcp/           # Main application entry point
├── internal/
│   ├── auth/                # Authentication providers
│   ├── client/              # HTTP client with retry logic
│   ├── config/              # Configuration management
│   ├── mcp/                 # MCP protocol implementation
│   └── tools/               # Tool implementations
│       ├── jira/            # Jira tools
│       ├── confluence/      # Confluence tools
│       └── opsgenie/        # Opsgenie tools
├── pkg/atlassian/           # Atlassian API clients
│   ├── jira/                # Jira API client
│   ├── confluence/          # Confluence API client
│   └── opsgenie/            # Opsgenie API client
└── docs/                    # Documentation
```

## Troubleshooting

### Authentication Failures

**Cloud:**
- Verify you're using an API token, not your password
- Check the token hasn't expired
- Ensure your email address is correct

**Server/Data Center:**
- Verify your Personal Access Token is valid
- Check token hasn't expired
- Ensure you have necessary permissions

### SSL Certificate Errors

For self-signed certificates on Server/Data Center:
```bash
JIRA_SSL_VERIFY=false
CONFLUENCE_SSL_VERIFY=false
```

### Connection Issues

Check proxy settings if behind corporate firewall:
```bash
HTTP_PROXY=http://proxy:8080
HTTPS_PROXY=http://proxy:8080
```

### Tool Not Available

If a tool isn't showing up:
1. Check service is configured (Jira, Confluence, or Opsgenie)
2. Verify service is enabled (e.g., `OPSGENIE_ENABLED=true`)
3. Verify `ENABLED_TOOLS` isn't filtering it out
4. Check if `READ_ONLY_MODE` is blocking write tools

## Security

- Never commit API tokens or credentials to version control
- Use `.env` files for local development (already in `.gitignore`)
- Enable `READ_ONLY_MODE` for untrusted environments
- Use tool filtering to limit exposed functionality
- Keep dependencies up to date

## Roadmap

- [x] Homebrew formula (automated with GoReleaser)
- [ ] HTTP transports (SSE, streamable-http)
- [ ] Multi-user support (per-request auth)
- [ ] Content conversion (Markdown ↔ Confluence Storage format)
- [ ] Attachment handling improvements
- [ ] Metrics and observability
- [ ] Docker images

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

This project is inspired by the Python [mcp-atlassian](https://github.com/sooperset/mcp-atlassian) implementation. While this is a complete rewrite in Go, we acknowledge the original project's design and feature set.

## Support

- 📖 [Product Specification](docs/product.md)
- 📋 [Implementation Plan](docs/implementation-plan.md)
- 🐛 [Report Issues](https://github.com/codeownersnet/atlas/issues)
- 💬 [Discussions](https://github.com/codeownersnet/atlas/discussions)

---

Made with ❤️ using Go and the Model Context Protocol
