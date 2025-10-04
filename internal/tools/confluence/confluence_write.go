package confluence

import (
	"context"
	"fmt"

	"github.com/codeownersnet/atlas/internal/mcp"
)

// ConfluenceCreatePageTool creates the confluence_create_page tool
func ConfluenceCreatePageTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_create_page",
		"Create a new Confluence page. Supports Markdown, Wiki markup, and Confluence storage format.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"space_key": mcp.NewStringProperty("Space key where the page will be created (e.g., 'DOCS')"),
				"title":     mcp.NewStringProperty("Page title"),
				"body":      mcp.NewStringProperty("Page content/body"),
				"format": mcp.NewStringProperty("Content format: 'storage' (Confluence storage format, default), 'markdown', or 'wiki'").
					WithDefault("storage"),
				"parent_id": mcp.NewStringProperty("Parent page ID (optional, for creating child pages)"),
			},
			"space_key", "title", "body",
		),
		confluenceCreatePageHandler,
		"confluence", "write",
	)
}

func confluenceCreatePageHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	spaceKey, ok := args["space_key"].(string)
	if !ok || spaceKey == "" {
		return nil, fmt.Errorf("space_key is required")
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title is required")
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return nil, fmt.Errorf("body is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	// Get format (default to storage)
	format := "storage"
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	}

	// Convert content based on format
	var contentBody string
	var err error
	switch format {
	case "markdown":
		contentBody, err = client.ConvertMarkdownToStorage(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("failed to convert markdown to storage format: %w", err)
		}
	case "wiki":
		contentBody, err = client.ConvertWikiToStorage(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("failed to convert wiki to storage format: %w", err)
		}
	case "storage":
		contentBody = body
	default:
		return nil, fmt.Errorf("unsupported format: %s. Use 'storage', 'markdown', or 'wiki'", format)
	}

	// Get optional parent ID
	parentID := ""
	if p, ok := args["parent_id"].(string); ok {
		parentID = p
	}

	page, err := client.CreatePage(ctx, spaceKey, title, contentBody, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	result := map[string]interface{}{
		"id":      page.ID,
		"title":   page.Title,
		"message": fmt.Sprintf("Successfully created page '%s'", page.Title),
	}

	// Add web UI link if available from page metadata
	if page.Links != nil && page.Links.WebUI != "" {
		result["webui"] = page.Links.WebUI
	}

	return mcp.NewJSONResult(result)
}

// ConfluenceUpdatePageTool creates the confluence_update_page tool
func ConfluenceUpdatePageTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_update_page",
		"Update an existing Confluence page. Requires the current version number to prevent conflicts.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"page_id": mcp.NewStringProperty("Page ID to update"),
				"title":   mcp.NewStringProperty("New page title (optional, keeps existing if not provided)"),
				"body":    mcp.NewStringProperty("New page content/body"),
				"version": mcp.NewIntegerProperty("Current version number of the page (required for conflict detection)"),
				"format": mcp.NewStringProperty("Content format: 'storage' (default), 'markdown', or 'wiki'").
					WithDefault("storage"),
			},
			"page_id", "body", "version",
		),
		confluenceUpdatePageHandler,
		"confluence", "write",
	)
}

func confluenceUpdatePageHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return nil, fmt.Errorf("body is required")
	}

	version := getIntArg(args, "version", 0)
	if version == 0 {
		return nil, fmt.Errorf("version is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	// Get the current page to get the title if not provided
	currentPage, err := client.GetPage(ctx, pageID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current page: %w", err)
	}

	title := currentPage.Title
	if t, ok := args["title"].(string); ok && t != "" {
		title = t
	}

	// Get format (default to storage)
	format := "storage"
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	}

	// Convert content based on format
	var contentBody string
	switch format {
	case "markdown":
		contentBody, err = client.ConvertMarkdownToStorage(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("failed to convert markdown to storage format: %w", err)
		}
	case "wiki":
		contentBody, err = client.ConvertWikiToStorage(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("failed to convert wiki to storage format: %w", err)
		}
	case "storage":
		contentBody = body
	default:
		return nil, fmt.Errorf("unsupported format: %s. Use 'storage', 'markdown', or 'wiki'", format)
	}

	// Update the page with incremented version
	page, err := client.UpdatePage(ctx, pageID, title, contentBody, version+1)
	if err != nil {
		return nil, fmt.Errorf("failed to update page: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      page.ID,
		"title":   page.Title,
		"version": page.Version.Number,
		"message": fmt.Sprintf("Successfully updated page '%s' to version %d", page.Title, page.Version.Number),
	})
}

// ConfluenceDeletePageTool creates the confluence_delete_page tool
func ConfluenceDeletePageTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_delete_page",
		"Delete a Confluence page. Use with caution as this action may not be reversible depending on your Confluence configuration.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"page_id": mcp.NewStringProperty("Page ID to delete"),
			},
			"page_id",
		),
		confluenceDeletePageHandler,
		"confluence", "write",
	)
}

func confluenceDeletePageHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	err := client.DeletePage(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete page: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully deleted page %s", pageID)), nil
}

// ConfluenceAddLabelTool creates the confluence_add_label tool
func ConfluenceAddLabelTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_add_label",
		"Add a label to a Confluence page or other content. Labels help with organization and searchability.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"content_id": mcp.NewStringProperty("Content ID (page ID, blogpost ID, etc.)"),
				"name":       mcp.NewStringProperty("Label name"),
				"prefix": mcp.NewStringProperty("Label prefix: 'global' (default), 'my', or 'team'").
					WithDefault("global"),
			},
			"content_id", "name",
		),
		confluenceAddLabelHandler,
		"confluence", "write",
	)
}

func confluenceAddLabelHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	contentID, ok := args["content_id"].(string)
	if !ok || contentID == "" {
		return nil, fmt.Errorf("content_id is required")
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	prefix := "global"
	if p, ok := args["prefix"].(string); ok && p != "" {
		prefix = p
	}

	label, err := client.AddLabel(ctx, contentID, name, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to add label: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"name":    label.Name,
		"prefix":  label.Prefix,
		"message": fmt.Sprintf("Successfully added label '%s' to content %s", label.Name, contentID),
	})
}

// ConfluenceAddCommentTool creates the confluence_add_comment tool
func ConfluenceAddCommentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"confluence_add_comment",
		"Add a comment to a Confluence page. Comments are useful for collaboration and feedback.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"page_id": mcp.NewStringProperty("Page ID to comment on"),
				"body":    mcp.NewStringProperty("Comment text/body (in Confluence storage format or plain text)"),
			},
			"page_id", "body",
		),
		confluenceAddCommentHandler,
		"confluence", "write",
	)
}

func confluenceAddCommentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	pageID, ok := args["page_id"].(string)
	if !ok || pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return nil, fmt.Errorf("body is required")
	}

	client := GetConfluenceClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Confluence client not available")
	}

	comment, err := client.AddComment(ctx, pageID, body)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      comment.ID,
		"message": fmt.Sprintf("Successfully added comment to page %s", pageID),
	})
}
