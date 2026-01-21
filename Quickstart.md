# Quickstart Guide

Get started with atlas-mcp in under 5 minutes.

## 1. Install via Homebrew

```bash
brew tap codeownersnet/atlas
brew install atlas-mcp
```

## 2. Get Your Jira Cloud API Token

1. Go to [https://id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**
3. Give it a label (e.g., "Claude Code")
4. Click **Create**
5. Copy the token (you won't be able to see it again)

## 3. Configure Claude Code

You can add the MCP server using the CLI wizard or by editing the configuration file directly:

### Option A: Using the CLI (Recommended)

```bash
claude mcp add --transport stdio atlas \
  --env JIRA_URL=https://your-domain.atlassian.net \
  --env JIRA_USERNAME=your-email@example.com \
  --env JIRA_API_TOKEN=your-api-token-here \
  -- atlas-mcp
```

> **Tip:** Add `--scope user` to install globally (available in all projects):
>
> ```bash
> claude mcp add --scope user --transport stdio atlas \
>   --env JIRA_URL=https://your-domain.atlassian.net \
>   --env JIRA_USERNAME=your-email@example.com \
>   --env JIRA_API_TOKEN=your-api-token-here \
>   -- atlas-mcp
> ```
>
> Without `--scope`, the server is only available in the current project.

Replace the values with your actual Jira URL, email, and API token.

### Option B: Manual Configuration

Edit `~/.claude.json` (for user/global scope) or `.mcp.json` (for project scope):

```json
{
  "mcpServers": {
    "atlas": {
      "type": "stdio",
      "command": "atlas-mcp",
      "env": {
        "JIRA_URL": "https://your-domain.atlassian.net",
        "JIRA_USERNAME": "your-email@example.com",
        "JIRA_API_TOKEN": "your-api-token-here"
      }
    }
  }
}
```

## 4. Restart Claude Code (if manually configured)

Close and reopen Claude Code for the changes to take effect.

## 5. Verify Installation

In Claude Code, try asking:

```
List my recent Jira issues
```

or

```
Search for issues assigned to me
```

You should see results from your Jira instance!

## What's Next?

- **Explore Available Tools**: Ask Claude "What Jira tools are available?"
- **Add Confluence**: Add the following environment variables to your config:
  ```json
  "CONFLUENCE_URL": "https://your-domain.atlassian.net/wiki",
  "CONFLUENCE_USERNAME": "your-email@example.com",
  "CONFLUENCE_API_TOKEN": "your-api-token-here"
  ```
- **Configure for Server/DC**: For on-premise Atlassian deployments, use `JIRA_PERSONAL_TOKEN` or `CONFLUENCE_PERSONAL_TOKEN` instead of username/API token
- **Add Opsgenie**: Add `OPSGENIE_URL` (default: `https://api.opsgenie.com`) and `OPSGENIE_API_KEY` to access Opsgenie
- **Enable Specific Tools**: Use `ENABLED_TOOLS` to control which tools are available

## Troubleshooting

**Can't connect to Jira?**
- Verify your Jira URL is correct (include `https://`)
- Check your email and API token are correct
- Ensure your Jira user has appropriate permissions

**Tools not showing up?**
- Restart Claude Code completely
- Check the MCP logs: `~/Library/Logs/Claude/mcp*.log` (macOS)
- Verify JSON syntax in config file

**Need more help?**
- Check [README.md](../README.md) for full documentation
- See [Product Specification](product.md) for feature details
- Open an issue on [GitHub](https://github.com/codeownersnet/atlas/issues)
