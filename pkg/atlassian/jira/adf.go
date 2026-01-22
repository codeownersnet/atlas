package jira

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ADF (Atlassian Document Format) types
// These structures represent the ADF JSON format used by Jira Cloud API v3

// ADFDocument represents a complete ADF document
type ADFDocument struct {
	Version int       `json:"version"`
	Type    string    `json:"type"`
	Content []ADFNode `json:"content"`
}

// ADFNode represents a node in the ADF document tree
type ADFNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []ADFNode              `json:"content,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Marks   []ADFMark              `json:"marks,omitempty"`
}

// ADFMark represents text formatting marks (bold, italic, etc.)
type ADFMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}

// MarkdownToADF converts a markdown or Jira wiki markup string to an ADF document.
// It automatically detects Jira wiki markup patterns (h1., h2., etc.) and converts them.
func MarkdownToADF(markdown string) *ADFDocument {
	if markdown == "" {
		return &ADFDocument{
			Version: 1,
			Type:    "doc",
			Content: []ADFNode{},
		}
	}

	// Pre-process: Convert Jira wiki markup to markdown if detected
	markdown = convertWikiToMarkdown(markdown)

	doc := &ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []ADFNode{},
	}

	lines := strings.Split(markdown, "\n")
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Blockquote: > quoted text
		// Check before code block to ensure proper order
		if blockquoteNode := parseBlockquote(line); blockquoteNode != nil {
			doc.Content = append(doc.Content, *blockquoteNode)
			i++
			continue
		}

		// Panel: [panelType] content
		if panelNode := parsePanel(line); panelNode != nil {
			doc.Content = append(doc.Content, *panelNode)
			i++
			continue
		}

		// Code block (fenced)
		if strings.HasPrefix(line, "```") {
			lang := strings.TrimPrefix(line, "```")
			lang = strings.TrimSpace(lang)
			codeLines := []string{}
			i++
			for i < len(lines) && !strings.HasPrefix(lines[i], "```") {
				codeLines = append(codeLines, lines[i])
				i++
			}
			i++ // Skip closing ```

			attrs := map[string]interface{}{}
			if lang != "" {
				attrs["language"] = lang
			}

			doc.Content = append(doc.Content, ADFNode{
				Type:  "codeBlock",
				Attrs: attrs,
				Content: []ADFNode{
					{Type: "text", Text: strings.Join(codeLines, "\n")},
				},
			})
			continue
		}

		// Table detection (lines starting with |)
		if strings.HasPrefix(strings.TrimSpace(line), "|") && strings.HasSuffix(strings.TrimSpace(line), "|") {
			tableNode := parseTable(lines, &i)
			if tableNode != nil {
				doc.Content = append(doc.Content, *tableNode)
			}
			continue
		}

		// Expand/collapsible: <details>Title</details>
		if expandNode := parseExpand(lines, &i); expandNode != nil {
			doc.Content = append(doc.Content, *expandNode)
			continue
		}

		// Heading
		if heading := parseHeading(line); heading != nil {
			doc.Content = append(doc.Content, *heading)
			i++
			continue
		}

		// Horizontal rule
		if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "***" || strings.TrimSpace(line) == "___" {
			doc.Content = append(doc.Content, ADFNode{Type: "rule"})
			i++
			continue
		}

		// Bullet list
		if strings.HasPrefix(strings.TrimLeft(line, " \t"), "- ") || strings.HasPrefix(strings.TrimLeft(line, " \t"), "* ") {
			listItems := []ADFNode{}
			for i < len(lines) {
				trimmed := strings.TrimLeft(lines[i], " \t")
				if !strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "* ") {
					break
				}
				content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
				listItems = append(listItems, ADFNode{
					Type: "listItem",
					Content: []ADFNode{
						{
							Type:    "paragraph",
							Content: parseInlineContent(content),
						},
					},
				})
				i++
			}
			doc.Content = append(doc.Content, ADFNode{
				Type:    "bulletList",
				Content: listItems,
			})
			continue
		}

		// Ordered list
		if matched, _ := regexp.MatchString(`^\d+\.\s`, strings.TrimLeft(line, " \t")); matched {
			listItems := []ADFNode{}
			for i < len(lines) {
				trimmed := strings.TrimLeft(lines[i], " \t")
				if matched, _ := regexp.MatchString(`^\d+\.\s`, trimmed); !matched {
					break
				}
				// Remove the number and dot prefix
				content := regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(trimmed, "")
				listItems = append(listItems, ADFNode{
					Type: "listItem",
					Content: []ADFNode{
						{
							Type:    "paragraph",
							Content: parseInlineContent(content),
						},
					},
				})
				i++
			}
			doc.Content = append(doc.Content, ADFNode{
				Type:    "orderedList",
				Content: listItems,
			})
			continue
		}

		// Empty line - skip
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Regular paragraph
		doc.Content = append(doc.Content, ADFNode{
			Type:    "paragraph",
			Content: parseInlineContent(line),
		})
		i++
	}

	return doc
}

// convertWikiToMarkdown converts Jira wiki markup to markdown
func convertWikiToMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))

	inCodeBlock := false
	codeBlockLang := ""

	for _, line := range lines {
		// Handle {code} blocks
		if strings.HasPrefix(line, "{code") {
			if !inCodeBlock {
				// Opening code block
				inCodeBlock = true
				// Extract language if specified: {code:java}
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						codeBlockLang = strings.TrimSuffix(parts[1], "}")
					}
				}
				result = append(result, "```"+codeBlockLang)
				continue
			}
		}
		if line == "{code}" && inCodeBlock {
			// Closing code block
			inCodeBlock = false
			codeBlockLang = ""
			result = append(result, "```")
			continue
		}

		// Don't process content inside code blocks
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Convert wiki headings: h1. Title -> # Title
		if wikiHeading := regexp.MustCompile(`^h([1-6])\.\s*(.*)$`); wikiHeading.MatchString(line) {
			matches := wikiHeading.FindStringSubmatch(line)
			if len(matches) == 3 {
				level := matches[1]
				text := matches[2]
				hashes := strings.Repeat("#", int(level[0]-'0'))
				result = append(result, hashes+" "+text)
				continue
			}
		}

		// Convert wiki horizontal rule: ---- -> ---
		if strings.TrimSpace(line) == "----" {
			result = append(result, "---")
			continue
		}

		// Convert wiki bold: *text* -> **text** (only if not already markdown bold)
		// Be careful: wiki uses single *, markdown uses double **
		// We need to detect wiki-style bold which uses single *
		line = convertWikiInlineFormatting(line)

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// convertWikiInlineFormatting converts Jira wiki inline formatting to markdown
func convertWikiInlineFormatting(line string) string {
	// Convert wiki bold *text* to markdown **text**
	// But be careful not to convert markdown's * for italic or existing **
	// Wiki bold: *text* (single asterisk)
	// Wiki italic: _text_ (underscore - same as markdown)
	// Markdown bold: **text** (double asterisk)
	// Markdown italic: *text* (single asterisk)

	// Since wiki and markdown conflict on * usage, we need a heuristic:
	// If the text has wiki-style patterns like h1. or {code}, assume wiki format
	// In that case, convert *text* to **text**

	// Convert {{monospace}} to `monospace`
	line = regexp.MustCompile(`\{\{([^}]+)\}\}`).ReplaceAllString(line, "`$1`")

	// Convert wiki links: [text|url] to markdown [text](url)
	line = regexp.MustCompile(`\[([^|\]]+)\|([^\]]+)\]`).ReplaceAllString(line, "[$1]($2)")

	// Convert simple wiki links: [url] to markdown [url](url)
	line = regexp.MustCompile(`\[([^\]|]+)\]`).ReplaceAllStringFunc(line, func(match string) string {
		// Don't convert if it's already markdown format [text](url)
		if strings.Contains(match, "](") {
			return match
		}
		url := strings.Trim(match, "[]")
		// Only convert if it looks like a URL
		if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "www.") {
			return "[" + url + "](" + url + ")"
		}
		return match
	})

	return line
}

// parseTable parses a markdown/wiki table starting at the current line
func parseTable(lines []string, i *int) *ADFNode {
	tableRows := []ADFNode{}
	isFirstRow := true

	for *i < len(lines) {
		line := strings.TrimSpace(lines[*i])
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			break
		}

		// Skip separator rows (|---|---|)
		if regexp.MustCompile(`^\|[\s\-:|]+\|$`).MatchString(line) {
			*i++
			continue
		}

		// Parse cells
		cells := strings.Split(strings.Trim(line, "|"), "|")
		rowCells := []ADFNode{}

		for _, cell := range cells {
			cellText := strings.TrimSpace(cell)
			cellType := "tableCell"
			if isFirstRow {
				cellType = "tableHeader"
			}

			rowCells = append(rowCells, ADFNode{
				Type: cellType,
				Content: []ADFNode{
					{
						Type:    "paragraph",
						Content: parseInlineContent(cellText),
					},
				},
			})
		}

		tableRows = append(tableRows, ADFNode{
			Type:    "tableRow",
			Content: rowCells,
		})

		isFirstRow = false
		*i++
	}

	if len(tableRows) == 0 {
		return nil
	}

	return &ADFNode{
		Type:    "table",
		Content: tableRows,
	}
}

// parseHeading parses a markdown heading line
func parseHeading(line string) *ADFNode {
	for level := 6; level >= 1; level-- {
		prefix := strings.Repeat("#", level) + " "
		if strings.HasPrefix(line, prefix) {
			text := strings.TrimPrefix(line, prefix)
			return &ADFNode{
				Type:    "heading",
				Attrs:   map[string]interface{}{"level": level},
				Content: parseInlineContent(text),
			}
		}
	}
	return nil
}

// parseBlockquote parses a blockquote line starting with ">"
func parseBlockquote(line string) *ADFNode {
	trimmed := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(trimmed, "> ") {
		quoteContent := strings.TrimPrefix(trimmed, "> ")
		return &ADFNode{
			Type: "blockquote",
			Content: []ADFNode{
				{
					Type:    "paragraph",
					Content: parseInlineContent(quoteContent),
				},
			},
		}
	}
	return nil
}

// parsePanel parses a panel line with syntax [panelType] content
// Supported panel types: info, warning, error, success, note, tip
func parsePanel(line string) *ADFNode {
	panelTypes := map[string]string{
		"info":     "info",
		"warning":  "warning",
		"error":    "error",
		"success":  "success",
		"note":     "info",
		"tip":      "success",
	}

	trimmed := strings.TrimSpace(line)
	// Match pattern: [panelType] content
	re := regexp.MustCompile(`^\[([a-zA-Z]+)\]\s*(.*)$`)
	matches := re.FindStringSubmatch(trimmed)
	if len(matches) == 3 {
		panelType := matches[1]
		content := matches[2]
		if adfType, ok := panelTypes[panelType]; ok {
			return &ADFNode{
				Type: "panel",
				Attrs: map[string]interface{}{
					"panelType": adfType,
				},
				Content: []ADFNode{
					{
						Type:    "paragraph",
						Content: parseInlineContent(content),
					},
				},
			}
		}
	}
	return nil
}

// parseExpand parses a collapsible section using HTML-style <details>Title</details> syntax
// The content after the opening tag until the closing tag is the title
// Any subsequent lines until an empty line become the body content
func parseExpand(lines []string, i *int) *ADFNode {
	if *i >= len(lines) {
		return nil
	}

	line := strings.TrimSpace(lines[*i])

	// Match opening <details>Title</details>
	openDetailsRe := regexp.MustCompile(`^<details>(.*?)</details>$`)
	matches := openDetailsRe.FindStringSubmatch(line)
	if len(matches) != 2 {
		return nil
	}

	title := matches[1]
	bodyContent := []ADFNode{}

	// Collect body content until empty line or end
	*i++
	for *i < len(lines) {
		nextLine := strings.TrimSpace(lines[*i])
		if nextLine == "" {
			// Empty line ends the expand section
			*i++
			break
		}
		bodyContent = append(bodyContent, ADFNode{
			Type:    "paragraph",
			Content: parseInlineContent(nextLine),
		})
		*i++
	}

	return &ADFNode{
		Type: "expand",
		Content: append([]ADFNode{
			{
				Type:    "paragraph",
				Content: parseInlineContent(title),
			},
		}, bodyContent...),
	}
}

// parseInlineContent parses inline markdown formatting (bold, italic, code, links)
func parseInlineContent(text string) []ADFNode {
	if text == "" {
		return []ADFNode{}
	}

	nodes := []ADFNode{}

	// Regular expressions for inline formatting
	// Order matters: process more complex patterns first
	patterns := []struct {
		re      *regexp.Regexp
		process func(match []string) ([]ADFNode, int)
	}{
		// Links: [text](url)
		{
			re: regexp.MustCompile(`^\[([^\]]+)\]\(([^)]+)\)`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type: "text",
					Text: match[1],
					Marks: []ADFMark{{
						Type:  "link",
						Attrs: map[string]interface{}{"href": match[2]},
					}},
				}}, len(match[0])
			},
		},
		// Inline code: `code`
		{
			re: regexp.MustCompile("^`([^`]+)`"),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "code"}},
				}}, len(match[0])
			},
		},
		// Bold and italic: ***text*** or ___text___
		{
			re: regexp.MustCompile(`^\*\*\*([^*]+)\*\*\*`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "strong"}, {Type: "em"}},
				}}, len(match[0])
			},
		},
		// Bold: **text** or __text__
		{
			re: regexp.MustCompile(`^\*\*([^*]+)\*\*`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "strong"}},
				}}, len(match[0])
			},
		},
		{
			re: regexp.MustCompile(`^__([^_]+)__`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "strong"}},
				}}, len(match[0])
			},
		},
		// Italic: *text* or _text_
		{
			re: regexp.MustCompile(`^\*([^*]+)\*`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "em"}},
				}}, len(match[0])
			},
		},
		{
			re: regexp.MustCompile(`^_([^_]+)_`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "em"}},
				}}, len(match[0])
			},
		},
		// Strikethrough: ~~text~~
		{
			re: regexp.MustCompile(`^~~([^~]+)~~`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "strike"}},
				}}, len(match[0])
			},
		},
		// Underline: ++text++
		{
			re: regexp.MustCompile(`^\+\+([^+]+)\+\+`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "text",
					Text:  match[1],
					Marks: []ADFMark{{Type: "underline"}},
				}}, len(match[0])
			},
		},
		// Status: [status:StatusName] - inline status node
		{
			re: regexp.MustCompile(`^\[status:([^\]]+)\]`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:    "status",
					Attrs:   map[string]interface{}{"text": match[1]},
					Content: []ADFNode{{Type: "text", Text: match[1]}},
				}}, len(match[0])
			},
		},
		// Emoji: :emoji_name:
		{
			re: regexp.MustCompile(`^:([a-zA-Z0-9_]+):`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "emoji",
					Attrs: map[string]interface{}{"shortName": match[1]},
				}}, len(match[0])
			},
		},
		// Mention: @username - creates a mention node with placeholder id
		{
			re: regexp.MustCompile(`^@([a-zA-Z0-9_.-]+)`),
			process: func(match []string) ([]ADFNode, int) {
				return []ADFNode{{
					Type:  "mention",
					Attrs: map[string]interface{}{
						"id":   match[1],
						"text": match[1],
					},
				}}, len(match[0])
			},
		},
	}

	i := 0
	var accumulated []byte

	for i < len(text) {
		// Find any pattern match at the current position (using ^ anchor)
		matchedPattern := -1
		matchedEnd := 0

		for patternIdx, p := range patterns {
			if match := p.re.FindStringSubmatchIndex(text[i:]); match != nil && match[0] == 0 {
				if matchedPattern == -1 || match[1] > matchedEnd {
					matchedPattern = patternIdx
					matchedEnd = match[1]
				}
			}
		}

		if matchedPattern >= 0 {
			// Found a pattern match at current position
			// First, flush any accumulated text as a single UTF-8 string
			if len(accumulated) > 0 {
				nodes = append(nodes, ADFNode{Type: "text", Text: string(accumulated)})
				accumulated = accumulated[:0] // Reset
			}

			// Process the pattern match
			p := patterns[matchedPattern]
			patternsMatch := p.re.FindStringSubmatch(text[i:])
			if patternsMatch != nil {
				newNodes, advance := p.process(patternsMatch)
				nodes = append(nodes, newNodes...)
				i += advance
			} else {
				i++
			}
		} else {
			// No pattern matches at current position - accumulate this byte
			// and move to the next byte position
			accumulated = append(accumulated, text[i])
			i++
		}
	}

	// Flush any remaining accumulated text
	if len(accumulated) > 0 {
		nodes = append(nodes, ADFNode{Type: "text", Text: string(accumulated)})
	}

	return nodes
}

// ADFToMarkdown converts an ADF document to markdown
func ADFToMarkdown(adf map[string]interface{}) string {
	if adf == nil {
		return ""
	}

	content, ok := adf["content"].([]interface{})
	if !ok {
		return ""
	}

	var result strings.Builder
	for i, item := range content {
		if node, ok := item.(map[string]interface{}); ok {
			text := nodeToMarkdown(node, 0)
			if text != "" {
				if i > 0 && result.Len() > 0 {
					// Add spacing between top-level blocks
					result.WriteString("\n")
				}
				result.WriteString(text)
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// nodeToMarkdown converts a single ADF node to markdown
func nodeToMarkdown(node map[string]interface{}, depth int) string {
	nodeType, _ := node["type"].(string)

	switch nodeType {
	case "paragraph":
		return contentToMarkdown(node) + "\n"

	case "blockquote":
		return "> " + contentToMarkdown(node) + "\n"

	case "panel":
		panelType := "info"
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if pt, ok := attrs["panelType"].(string); ok {
				panelType = pt
			}
		}
		return "[" + panelType + "] " + contentToMarkdown(node) + "\n"

	case "expand":
		content, ok := node["content"].([]interface{})
		if !ok || len(content) == 0 {
			return ""
		}
		// First node is the title
		title := "> "
		if titleNode, ok := content[0].(map[string]interface{}); ok {
			title = contentToMarkdown(titleNode)
		}
		var body strings.Builder
		for i := 1; i < len(content); i++ {
			if itemNode, ok := content[i].(map[string]interface{}); ok {
				body.WriteString(nodeToMarkdown(itemNode, 0))
				body.WriteString("\n")
			}
		}
		// HTML-style tags for expand
		bodyText := strings.TrimSpace(body.String())
		if bodyText != "" {
			return "<details>" + title + "</details>\n" + bodyText + "\n"
		}
		return "<details>" + title + "</details>\n"

	case "heading":
		level := 1
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if l, ok := attrs["level"].(float64); ok {
				level = int(l)
			}
		}
		prefix := strings.Repeat("#", level) + " "
		return prefix + contentToMarkdown(node) + "\n"

	case "bulletList":
		return listToMarkdown(node, "- ", depth)

	case "orderedList":
		return orderedListToMarkdown(node, depth)

	case "listItem":
		var result strings.Builder
		if content, ok := node["content"].([]interface{}); ok {
			for _, item := range content {
				if itemNode, ok := item.(map[string]interface{}); ok {
					text := nodeToMarkdown(itemNode, depth)
					// Remove trailing newline for list items
					text = strings.TrimSuffix(text, "\n")
					result.WriteString(text)
				}
			}
		}
		return result.String()

	case "codeBlock":
		lang := ""
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if l, ok := attrs["language"].(string); ok {
				lang = l
			}
		}
		code := contentToMarkdown(node)
		return "```" + lang + "\n" + code + "\n```\n"

	case "rule":
		return "---\n"

	case "text":
		return textNodeToMarkdown(node)

	case "hardBreak":
		return "\n"

	case "mention":
		// Jira Cloud mentions use accountId, Server/DC use username
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if text, ok := attrs["text"].(string); ok {
				return "@" + text
			}
		}
		// Fallback to id attribute
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if id, ok := attrs["id"].(string); ok {
				return "@" + id
			}
		}
		return ""

	case "emoji":
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if shortName, ok := attrs["shortName"].(string); ok {
				return ":" + shortName + ":"
			}
		}
		return ""

	case "status":
		if attrs, ok := node["attrs"].(map[string]interface{}); ok {
			if text, ok := attrs["text"].(string); ok {
				return "[status:" + text + "]"
			}
		}
		return ""

	default:
		// For unknown types, try to extract content recursively
		return contentToMarkdown(node)
	}
}

// contentToMarkdown extracts and converts content from a node
func contentToMarkdown(node map[string]interface{}) string {
	content, ok := node["content"].([]interface{})
	if !ok {
		// Check for direct text content
		if _, hasText := node["text"].(string); hasText {
			return textNodeToMarkdown(node)
		}
		return ""
	}

	var result strings.Builder
	for _, item := range content {
		if itemNode, ok := item.(map[string]interface{}); ok {
			nodeType, _ := itemNode["type"].(string)
			// Handle inline nodes directly
			switch nodeType {
			case "text", "mention", "emoji", "status":
				result.WriteString(nodeToMarkdown(itemNode, 0))
			default:
				// For other node types, process normally
				text := nodeToMarkdown(itemNode, 0)
				// Remove trailing newline for inline content
				text = strings.TrimSuffix(text, "\n")
				result.WriteString(text)
			}
		}
	}

	return result.String()
}

// textNodeToMarkdown converts a text node with marks to markdown
func textNodeToMarkdown(node map[string]interface{}) string {
	text, ok := node["text"].(string)
	if !ok {
		return ""
	}

	marks, _ := node["marks"].([]interface{})

	// Apply marks in order: link wraps everything, then code, then styling
	hasStrong := false
	hasEm := false
	hasCode := false
	hasStrike := false
	hasUnderline := false
	var linkHref string

	for _, mark := range marks {
		if m, ok := mark.(map[string]interface{}); ok {
			markType, _ := m["type"].(string)
			switch markType {
			case "strong":
				hasStrong = true
			case "em":
				hasEm = true
			case "code":
				hasCode = true
			case "strike":
				hasStrike = true
			case "underline":
				hasUnderline = true
			case "link":
				if attrs, ok := m["attrs"].(map[string]interface{}); ok {
					linkHref, _ = attrs["href"].(string)
				}
			}
		}
	}

	result := text

	// Apply formatting (innermost first)
	if hasCode {
		result = "`" + result + "`"
	}
	if hasStrike {
		result = "~~" + result + "~~"
	}
	if hasUnderline {
		result = "++" + result + "++"
	}
	if hasEm {
		result = "*" + result + "*"
	}
	if hasStrong {
		result = "**" + result + "**"
	}
	if linkHref != "" {
		result = "[" + result + "](" + linkHref + ")"
	}

	return result
}

// listToMarkdown converts a bullet list to markdown
func listToMarkdown(node map[string]interface{}, prefix string, depth int) string {
	content, ok := node["content"].([]interface{})
	if !ok {
		return ""
	}

	var result strings.Builder
	indent := strings.Repeat("  ", depth)

	for _, item := range content {
		if itemNode, ok := item.(map[string]interface{}); ok {
			text := nodeToMarkdown(itemNode, depth+1)
			result.WriteString(indent + prefix + text + "\n")
		}
	}

	return result.String()
}

// orderedListToMarkdown converts an ordered list to markdown
func orderedListToMarkdown(node map[string]interface{}, depth int) string {
	content, ok := node["content"].([]interface{})
	if !ok {
		return ""
	}

	var result strings.Builder
	indent := strings.Repeat("  ", depth)

	for i, item := range content {
		if itemNode, ok := item.(map[string]interface{}); ok {
			text := nodeToMarkdown(itemNode, depth+1)
			result.WriteString(indent + string(rune('1'+i)) + ". " + text + "\n")
		}
	}

	return result.String()
}

// ToJSON converts an ADF document to JSON bytes
func (doc *ADFDocument) ToJSON() ([]byte, error) {
	return json.Marshal(doc)
}

// ToMap converts an ADF document to a map[string]interface{} for embedding in requests
func (doc *ADFDocument) ToMap() map[string]interface{} {
	// Marshal and unmarshal to get a clean map representation
	data, _ := json.Marshal(doc)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}
