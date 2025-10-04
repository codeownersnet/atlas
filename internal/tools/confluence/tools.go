package confluence

import (
	"context"
	"fmt"

	"github.com/codeownersnet/atlas/internal/mcp"
	"github.com/codeownersnet/atlas/pkg/atlassian/confluence"
)

// Context key for storing Confluence client
type contextKey string

const confluenceClientKey contextKey = "confluence_client"

// WithConfluenceClient adds a Confluence client to the context
func WithConfluenceClient(ctx context.Context, client *confluence.Client) context.Context {
	return context.WithValue(ctx, confluenceClientKey, client)
}

// GetConfluenceClient retrieves the Confluence client from the context
func GetConfluenceClient(ctx context.Context) *confluence.Client {
	client, ok := ctx.Value(confluenceClientKey).(*confluence.Client)
	if !ok {
		return nil
	}
	return client
}

// RegisterConfluenceTools registers all Confluence tools with the MCP server
func RegisterConfluenceTools(server *mcp.Server) error {
	tools := []struct {
		name string
		tool *mcp.ToolDefinition
	}{
		// Read operations
		{"confluence_search", ConfluenceSearchTool()},
		{"confluence_get_page", ConfluenceGetPageTool()},
		{"confluence_get_page_children", ConfluenceGetPageChildrenTool()},
		{"confluence_get_comments", ConfluenceGetCommentsTool()},
		{"confluence_get_labels", ConfluenceGetLabelsTool()},
		{"confluence_search_user", ConfluenceSearchUserTool()},

		// Write operations
		{"confluence_create_page", ConfluenceCreatePageTool()},
		{"confluence_update_page", ConfluenceUpdatePageTool()},
		{"confluence_delete_page", ConfluenceDeletePageTool()},
		{"confluence_add_label", ConfluenceAddLabelTool()},
		{"confluence_add_comment", ConfluenceAddCommentTool()},
	}

	for _, t := range tools {
		if err := server.RegisterTool(t.tool); err != nil {
			return fmt.Errorf("failed to register %s: %w", t.name, err)
		}
	}

	return nil
}
