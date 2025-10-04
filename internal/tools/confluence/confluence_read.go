package confluence

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/codeownersnet/atlas/internal/mcp"
	"github.com/codeownersnet/atlas/pkg/atlassian/confluence"
)

// ConfluenceSearchTool creates the confluence_search tool
func ConfluenceSearchTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_search",
		"Search Confluence content using CQL (Confluence Query Language) or simple text search. Automatically detects query type.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"query":  mcp.NewStringProperty("Search query. Can be CQL (e.g., 'type=page AND space=DOCS') or simple text for full-text search"),
				"expand": mcp.NewStringProperty("Resources to expand (e.g., 'body.storage,version,space'). Comma-separated."),
				"limit": mcp.NewIntegerProperty("Maximum number of results to return (default 25)").
					WithDefault(25),
				"start": mcp.NewIntegerProperty("Starting index for pagination (0-based)").
					WithDefault(0),
			},
			"query",
		),
		confluenceSearchHandler,
		"confluence", "read",
	)
}

func confluenceSearchHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	opts := &confluence.SearchOptions{
		Start: getIntArg(args, "start", 0),
		Limit: getIntArg(args, "limit", 25),
	}

	if expand, ok := args["expand"].(string); ok && expand != "" {
		opts.Expand = strings.Split(expand, ",")
	}

	result, err := client.Search(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// ConfluenceGetPageTool creates the confluence_get_page tool
func ConfluenceGetPageTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_get_page",
		"Get a Confluence page by ID or by title and space key. Returns page content and metadata.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"page_id":   mcp.NewStringProperty("Page ID (use this OR title+space_key)"),
				"title":     mcp.NewStringProperty("Page title (requires space_key)"),
				"space_key": mcp.NewStringProperty("Space key (required when using title)"),
				"expand":    mcp.NewStringProperty("Resources to expand (e.g., 'body.storage,version,space'). Comma-separated."),
			},
		),
		confluenceGetPageHandler,
		"confluence", "read",
	)
}

func confluenceGetPageHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	var expand []string
	if expandStr, ok := args["expand"].(string); ok && expandStr != "" {
		expand = strings.Split(expandStr, ",")
	}

	var page *confluence.Content
	var err error

	// Check if we're getting by ID or by title+space
	if pageID, ok := args["page_id"].(string); ok && pageID != "" {
		page, err = client.GetPage(ctx, pageID, expand)
	} else if title, ok := args["title"].(string); ok && title != "" {
		spaceKey, ok := args["space_key"].(string)
		if !ok || spaceKey == "" {
			return nil, fmt.Errorf("space_key is required when using title")
		}
		page, err = client.GetPageByTitle(ctx, spaceKey, title, expand)
	} else {
		return nil, fmt.Errorf("either page_id or (title and space_key) must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	return mcp.NewJSONResult(page)
}

// ConfluenceGetPageChildrenTool creates the confluence_get_page_children tool
func ConfluenceGetPageChildrenTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_get_page_children",
		"Get child pages of a Confluence page. Useful for navigating page hierarchies.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"page_id": mcp.NewStringProperty("Parent page ID"),
				"expand":  mcp.NewStringProperty("Resources to expand (e.g., 'body.storage,version'). Comma-separated."),
				"limit": mcp.NewIntegerProperty("Maximum number of children to return (default 25)").
					WithDefault(25),
			},
			"page_id",
		),
		confluenceGetPageChildrenHandler,
		"confluence", "read",
	)
}

func confluenceGetPageChildrenHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	var expand []string
	if expandStr, ok := args["expand"].(string); ok && expandStr != "" {
		expand = strings.Split(expandStr, ",")
	}

	limit := getIntArg(args, "limit", 25)

	children, err := client.GetPageChildren(ctx, pageID, expand, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get page children: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"children": children,
		"total":    len(children),
	})
}

// ConfluenceGetCommentsTool creates the confluence_get_comments tool
func ConfluenceGetCommentsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_get_comments",
		"Get comments for a Confluence page.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"page_id": mcp.NewStringProperty("Page ID"),
				"expand":  mcp.NewStringProperty("Resources to expand (e.g., 'body.storage,version'). Comma-separated."),
				"limit": mcp.NewIntegerProperty("Maximum number of comments to return (default 25)").
					WithDefault(25),
			},
			"page_id",
		),
		confluenceGetCommentsHandler,
		"confluence", "read",
	)
}

func confluenceGetCommentsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	var expand []string
	if expandStr, ok := args["expand"].(string); ok && expandStr != "" {
		expand = strings.Split(expandStr, ",")
	}

	limit := getIntArg(args, "limit", 25)

	comments, err := client.GetComments(ctx, pageID, expand, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"comments": comments,
		"total":    len(comments),
	})
}

// ConfluenceGetLabelsTool creates the confluence_get_labels tool
func ConfluenceGetLabelsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_get_labels",
		"Get labels for a Confluence page or other content.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"content_id": mcp.NewStringProperty("Content ID (page ID, blogpost ID, etc.)"),
				"prefix":     mcp.NewStringProperty("Filter labels by prefix (e.g., 'global', 'my', 'team')"),
				"limit": mcp.NewIntegerProperty("Maximum number of labels to return (default 100)").
					WithDefault(100),
			},
			"content_id",
		),
		confluenceGetLabelsHandler,
		"confluence", "read",
	)
}

func confluenceGetLabelsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	contentID, ok := args["content_id"].(string)
	if !ok || contentID == "" {
		return nil, fmt.Errorf("content_id is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	prefix := ""
	if p, ok := args["prefix"].(string); ok {
		prefix = p
	}

	limit := getIntArg(args, "limit", 100)

	labels, err := client.GetLabels(ctx, contentID, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"labels": labels,
		"total":  len(labels),
	})
}

// ConfluenceSearchUserTool creates the confluence_search_user tool
func ConfluenceSearchUserTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_search_user",
		"Search for Confluence users. Returns user information including account ID, display name, and email.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"query": mcp.NewStringProperty("Search query (username, display name, or email)"),
				"limit": mcp.NewIntegerProperty("Maximum number of users to return (default 25)").
					WithDefault(25),
			},
			"query",
		),
		confluenceSearchUserHandler,
		"confluence", "read",
	)
}

func confluenceSearchUserHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	limit := getIntArg(args, "limit", 25)

	users, err := client.SearchUsersByName(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

// Helper function to get integer argument with default
func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultVal
}
