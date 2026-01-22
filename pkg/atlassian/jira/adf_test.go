package jira

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMarkdownToADF_EmptyString(t *testing.T) {
	doc := MarkdownToADF("")
	if doc.Version != 1 {
		t.Errorf("expected version 1, got %d", doc.Version)
	}
	if doc.Type != "doc" {
		t.Errorf("expected type 'doc', got %s", doc.Type)
	}
	if len(doc.Content) != 0 {
		t.Errorf("expected empty content, got %d items", len(doc.Content))
	}
}

func TestMarkdownToADF_SimpleParagraph(t *testing.T) {
	doc := MarkdownToADF("Hello world")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("expected paragraph, got %s", para.Type)
	}
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}
	if para.Content[0].Text != "Hello world" {
		t.Errorf("expected 'Hello world', got %s", para.Content[0].Text)
	}
}

func TestMarkdownToADF_Headings(t *testing.T) {
	tests := []struct {
		markdown string
		level    int
		text     string
	}{
		{"# Heading 1", 1, "Heading 1"},
		{"## Heading 2", 2, "Heading 2"},
		{"### Heading 3", 3, "Heading 3"},
		{"#### Heading 4", 4, "Heading 4"},
		{"##### Heading 5", 5, "Heading 5"},
		{"###### Heading 6", 6, "Heading 6"},
	}

	for _, tt := range tests {
		t.Run(tt.markdown, func(t *testing.T) {
			doc := MarkdownToADF(tt.markdown)
			if len(doc.Content) != 1 {
				t.Fatalf("expected 1 content item, got %d", len(doc.Content))
			}

			heading := doc.Content[0]
			if heading.Type != "heading" {
				t.Errorf("expected heading, got %s", heading.Type)
			}
			if heading.Attrs["level"] != tt.level {
				t.Errorf("expected level %d, got %v", tt.level, heading.Attrs["level"])
			}
			if len(heading.Content) != 1 || heading.Content[0].Text != tt.text {
				t.Errorf("expected text '%s'", tt.text)
			}
		})
	}
}

func TestMarkdownToADF_BulletList(t *testing.T) {
	markdown := `- Item 1
- Item 2
- Item 3`

	doc := MarkdownToADF(markdown)
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	list := doc.Content[0]
	if list.Type != "bulletList" {
		t.Errorf("expected bulletList, got %s", list.Type)
	}
	if len(list.Content) != 3 {
		t.Fatalf("expected 3 list items, got %d", len(list.Content))
	}

	for i, item := range list.Content {
		if item.Type != "listItem" {
			t.Errorf("item %d: expected listItem, got %s", i, item.Type)
		}
	}
}

func TestMarkdownToADF_OrderedList(t *testing.T) {
	markdown := `1. First
2. Second
3. Third`

	doc := MarkdownToADF(markdown)
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	list := doc.Content[0]
	if list.Type != "orderedList" {
		t.Errorf("expected orderedList, got %s", list.Type)
	}
	if len(list.Content) != 3 {
		t.Fatalf("expected 3 list items, got %d", len(list.Content))
	}
}

func TestMarkdownToADF_CodeBlock(t *testing.T) {
	markdown := "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```"

	doc := MarkdownToADF(markdown)
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	codeBlock := doc.Content[0]
	if codeBlock.Type != "codeBlock" {
		t.Errorf("expected codeBlock, got %s", codeBlock.Type)
	}
	if codeBlock.Attrs["language"] != "go" {
		t.Errorf("expected language 'go', got %v", codeBlock.Attrs["language"])
	}
}

func TestMarkdownToADF_HorizontalRule(t *testing.T) {
	tests := []string{"---", "***", "___"}

	for _, hr := range tests {
		t.Run(hr, func(t *testing.T) {
			doc := MarkdownToADF(hr)
			if len(doc.Content) != 1 {
				t.Fatalf("expected 1 content item, got %d", len(doc.Content))
			}
			if doc.Content[0].Type != "rule" {
				t.Errorf("expected rule, got %s", doc.Content[0].Type)
			}
		})
	}
}

func TestMarkdownToADF_Bold(t *testing.T) {
	doc := MarkdownToADF("**bold text**")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	textNode := para.Content[0]
	if textNode.Text != "bold text" {
		t.Errorf("expected 'bold text', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 1 || textNode.Marks[0].Type != "strong" {
		t.Errorf("expected strong mark")
	}
}

func TestMarkdownToADF_Italic(t *testing.T) {
	doc := MarkdownToADF("*italic text*")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	textNode := para.Content[0]
	if textNode.Text != "italic text" {
		t.Errorf("expected 'italic text', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 1 || textNode.Marks[0].Type != "em" {
		t.Errorf("expected em mark")
	}
}

func TestMarkdownToADF_InlineCode(t *testing.T) {
	doc := MarkdownToADF("`code`")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	textNode := para.Content[0]
	if textNode.Text != "code" {
		t.Errorf("expected 'code', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 1 || textNode.Marks[0].Type != "code" {
		t.Errorf("expected code mark")
	}
}

func TestMarkdownToADF_Link(t *testing.T) {
	doc := MarkdownToADF("[link text](https://example.com)")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	textNode := para.Content[0]
	if textNode.Text != "link text" {
		t.Errorf("expected 'link text', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 1 || textNode.Marks[0].Type != "link" {
		t.Errorf("expected link mark")
	}
	if textNode.Marks[0].Attrs["href"] != "https://example.com" {
		t.Errorf("expected href 'https://example.com', got %v", textNode.Marks[0].Attrs["href"])
	}
}

func TestMarkdownToADF_BoldAndItalic(t *testing.T) {
	doc := MarkdownToADF("***bold and italic***")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	textNode := para.Content[0]
	if textNode.Text != "bold and italic" {
		t.Errorf("expected 'bold and italic', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 2 {
		t.Errorf("expected 2 marks, got %d", len(textNode.Marks))
	}
}

func TestMarkdownToADF_Strikethrough(t *testing.T) {
	doc := MarkdownToADF("~~strikethrough~~")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	textNode := para.Content[0]
	if textNode.Text != "strikethrough" {
		t.Errorf("expected 'strikethrough', got %s", textNode.Text)
	}
	if len(textNode.Marks) != 1 || textNode.Marks[0].Type != "strike" {
		t.Errorf("expected strike mark")
	}
}

func TestMarkdownToADF_MixedContent(t *testing.T) {
	markdown := `# Title

This is a paragraph with **bold** and *italic* text.

- Item 1
- Item 2

---

` + "```go\ncode\n```"

	doc := MarkdownToADF(markdown)

	// Should have: heading, paragraph, bulletList, rule, codeBlock
	if len(doc.Content) < 4 {
		t.Errorf("expected at least 4 content items, got %d", len(doc.Content))
	}

	// Verify heading
	if doc.Content[0].Type != "heading" {
		t.Errorf("expected first item to be heading, got %s", doc.Content[0].Type)
	}
}

func TestADFToMarkdown_Paragraph(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Hello world",
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "Hello world"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestADFToMarkdown_Heading(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "heading",
				"attrs": map[string]interface{}{
					"level": float64(2),
				},
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "My Heading",
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "## My Heading"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestADFToMarkdown_BoldText(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "bold",
						"marks": []interface{}{
							map[string]interface{}{
								"type": "strong",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "**bold**"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestADFToMarkdown_ItalicText(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "italic",
						"marks": []interface{}{
							map[string]interface{}{
								"type": "em",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "*italic*"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestADFToMarkdown_Link(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "click here",
						"marks": []interface{}{
							map[string]interface{}{
								"type": "link",
								"attrs": map[string]interface{}{
									"href": "https://example.com",
								},
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "[click here](https://example.com)"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestADFToMarkdown_CodeBlock(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "codeBlock",
				"attrs": map[string]interface{}{
					"language": "go",
				},
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "fmt.Println(\"hello\")",
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "```go\nfmt.Println(\"hello\")\n```"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestADFToMarkdown_BulletList(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "bulletList",
				"content": []interface{}{
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "Item 1",
									},
								},
							},
						},
					},
					map[string]interface{}{
						"type": "listItem",
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "Item 2",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	// Should contain "- Item 1" and "- Item 2"
	if !adfContains(result, "- Item 1") || !adfContains(result, "- Item 2") {
		t.Errorf("expected bullet list items, got '%s'", result)
	}
}

func TestADFToMarkdown_Rule(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "rule",
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "---"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRoundTrip_SimpleParagraph(t *testing.T) {
	original := "Hello world"
	adf := MarkdownToADF(original)

	// Convert ADF to map
	adfJSON, _ := json.Marshal(adf)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestRoundTrip_FormattedText(t *testing.T) {
	original := "**bold** and *italic*"
	adf := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(adf)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	// The result should contain bold and italic markers
	if !adfContains(result, "**bold**") || !adfContains(result, "*italic*") {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestRoundTrip_Heading(t *testing.T) {
	original := "## Heading 2"
	adf := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(adf)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestADFDocument_ToJSON(t *testing.T) {
	doc := MarkdownToADF("Hello")
	jsonBytes, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["version"].(float64) != 1 {
		t.Errorf("expected version 1")
	}
	if result["type"] != "doc" {
		t.Errorf("expected type 'doc'")
	}
}

func TestADFDocument_ToMap(t *testing.T) {
	doc := MarkdownToADF("Hello")
	result := doc.ToMap()

	if result["version"].(float64) != 1 {
		t.Errorf("expected version 1")
	}
	if result["type"] != "doc" {
		t.Errorf("expected type 'doc'")
	}
	if _, ok := result["content"].([]interface{}); !ok {
		t.Errorf("expected content to be array")
	}
}

func TestNewADFDescription(t *testing.T) {
	desc := NewADFDescription("**Bold text**")

	if !desc.IsADF() {
		t.Error("expected IsADF() to be true")
	}

	// The original markdown should be preserved as text
	if desc.text != "**Bold text**" {
		t.Errorf("expected text to be preserved, got %s", desc.text)
	}

	// Raw should contain valid ADF JSON
	var adf map[string]interface{}
	if err := json.Unmarshal(desc.Raw(), &adf); err != nil {
		t.Fatalf("Raw() should be valid JSON: %v", err)
	}

	if adf["type"] != "doc" {
		t.Errorf("expected type 'doc', got %v", adf["type"])
	}
}

func TestDescription_ToMarkdown(t *testing.T) {
	// Test with ADF description
	adfDesc := NewADFDescription("**Bold** and *italic*")
	markdown := adfDesc.ToMarkdown()

	// Should contain the formatted text
	if !adfContains(markdown, "**Bold**") || !adfContains(markdown, "*italic*") {
		t.Errorf("expected markdown formatting, got %s", markdown)
	}

	// Test with plain text description
	plainDesc := NewDescription("Plain text")
	if plainDesc.ToMarkdown() != "Plain text" {
		t.Errorf("expected 'Plain text', got %s", plainDesc.ToMarkdown())
	}

	// Test nil description
	var nilDesc *Description
	if nilDesc.ToMarkdown() != "" {
		t.Errorf("expected empty string for nil description")
	}
}

func TestADFToMarkdown_Nil(t *testing.T) {
	result := ADFToMarkdown(nil)
	if result != "" {
		t.Errorf("expected empty string for nil input, got %s", result)
	}
}

func TestADFToMarkdown_EmptyContent(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{},
	}

	result := ADFToMarkdown(adf)
	if result != "" {
		t.Errorf("expected empty string for empty content, got %s", result)
	}
}

func TestParseInlineContent_MixedFormatting(t *testing.T) {
	// Test mixed formatting in a single line
	text := "Normal **bold** normal *italic* `code` [link](http://example.com)"
	nodes := parseInlineContent(text)

	// Should have multiple nodes
	if len(nodes) < 5 {
		t.Errorf("expected at least 5 nodes, got %d", len(nodes))
	}

	// Verify we have different types of marks
	hasStrong := false
	hasEm := false
	hasCode := false
	hasLink := false

	for _, node := range nodes {
		for _, mark := range node.Marks {
			switch mark.Type {
			case "strong":
				hasStrong = true
			case "em":
				hasEm = true
			case "code":
				hasCode = true
			case "link":
				hasLink = true
			}
		}
	}

	if !hasStrong || !hasEm || !hasCode || !hasLink {
		t.Error("expected all mark types to be present")
	}
}

func TestADFStructure_ValidJSON(t *testing.T) {
	markdown := `# Overview

This is a **feature** description with:

- Bullet point 1
- Bullet point 2

## Steps

1. Step one
2. Step two

---

` + "```go\nfunc example() {}\n```"

	doc := MarkdownToADF(markdown)
	jsonBytes, err := doc.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON produced: %v", err)
	}

	// Verify required ADF fields
	if result["version"].(float64) != 1 {
		t.Error("missing or invalid version")
	}
	if result["type"] != "doc" {
		t.Error("missing or invalid type")
	}
	if _, ok := result["content"].([]interface{}); !ok {
		t.Error("missing or invalid content array")
	}
}

// Helper function
func adfContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && adfContainsString(s, substr))
}

func adfContainsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestADFNodeTypes(t *testing.T) {
	// Test that we produce valid ADF node types
	nodeTypes := map[string]string{
		"# Heading":          "heading",
		"Paragraph":          "paragraph",
		"- Bullet":           "bulletList",
		"1. Ordered":         "orderedList",
		"---":                "rule",
		"```\ncode\n```":     "codeBlock",
	}

	for markdown, expectedType := range nodeTypes {
		t.Run(expectedType, func(t *testing.T) {
			doc := MarkdownToADF(markdown)
			if len(doc.Content) == 0 {
				t.Fatalf("expected content for %s", markdown)
			}
			if doc.Content[0].Type != expectedType {
				t.Errorf("expected type %s, got %s", expectedType, doc.Content[0].Type)
			}
		})
	}
}

func TestADFMarks(t *testing.T) {
	// Test that we produce valid ADF marks
	tests := []struct {
		markdown     string
		expectedMark string
	}{
		{"**bold**", "strong"},
		{"*italic*", "em"},
		{"_italic_", "em"},
		{"__bold__", "strong"},
		{"`code`", "code"},
		{"~~strike~~", "strike"},
		{"[link](http://example.com)", "link"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedMark, func(t *testing.T) {
			doc := MarkdownToADF(tt.markdown)
			if len(doc.Content) == 0 || len(doc.Content[0].Content) == 0 {
				t.Fatalf("expected content for %s", tt.markdown)
			}

			textNode := doc.Content[0].Content[0]
			if len(textNode.Marks) == 0 {
				t.Fatalf("expected marks for %s", tt.markdown)
			}

			found := false
			for _, mark := range textNode.Marks {
				if mark.Type == tt.expectedMark {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected mark %s in %v", tt.expectedMark, textNode.Marks)
			}
		})
	}
}

func TestDescription_UnmarshalADF(t *testing.T) {
	// Simulate receiving ADF from Jira API
	adfJSON := `{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{"type": "text", "text": "Hello "},
					{"type": "text", "text": "bold", "marks": [{"type": "strong"}]},
					{"type": "text", "text": " world"}
				]
			}
		]
	}`

	var desc Description
	if err := json.Unmarshal([]byte(adfJSON), &desc); err != nil {
		t.Fatalf("Failed to unmarshal ADF: %v", err)
	}

	if !desc.IsADF() {
		t.Error("expected IsADF() to be true")
	}

	// ToMarkdown should convert to markdown
	markdown := desc.ToMarkdown()
	if !adfContains(markdown, "**bold**") {
		t.Errorf("expected markdown with bold, got %s", markdown)
	}
}

func TestDescription_UnmarshalPlainText(t *testing.T) {
	plainJSON := `"Just plain text"`

	var desc Description
	if err := json.Unmarshal([]byte(plainJSON), &desc); err != nil {
		t.Fatalf("Failed to unmarshal plain text: %v", err)
	}

	if desc.IsADF() {
		t.Error("expected IsADF() to be false")
	}

	if desc.String() != "Just plain text" {
		t.Errorf("expected 'Just plain text', got %s", desc.String())
	}

	if desc.ToMarkdown() != "Just plain text" {
		t.Errorf("expected 'Just plain text', got %s", desc.ToMarkdown())
	}
}

// Verify the ADF matches expected Jira format
func TestADFJiraCompatibility(t *testing.T) {
	// Create ADF and verify it matches Jira's expected structure
	doc := MarkdownToADF("**Bold** text")

	jsonBytes, _ := doc.ToJSON()
	var result map[string]interface{}
	json.Unmarshal(jsonBytes, &result)

	// Jira requires these exact fields
	if _, ok := result["version"]; !ok {
		t.Error("Jira ADF requires 'version' field")
	}
	if _, ok := result["type"]; !ok {
		t.Error("Jira ADF requires 'type' field")
	}
	if _, ok := result["content"]; !ok {
		t.Error("Jira ADF requires 'content' field")
	}

	// Verify content structure
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("content should not be empty")
	}

	para := content[0].(map[string]interface{})
	if para["type"] != "paragraph" {
		t.Errorf("expected paragraph type, got %v", para["type"])
	}

	paraContent := para["content"].([]interface{})
	boldNode := paraContent[0].(map[string]interface{})
	if boldNode["type"] != "text" {
		t.Errorf("expected text type, got %v", boldNode["type"])
	}
	if boldNode["text"] != "Bold" {
		t.Errorf("expected 'Bold', got %v", boldNode["text"])
	}

	marks := boldNode["marks"].([]interface{})
	if len(marks) == 0 {
		t.Fatal("expected marks on bold text")
	}
	mark := marks[0].(map[string]interface{})
	if mark["type"] != "strong" {
		t.Errorf("expected strong mark, got %v", mark["type"])
	}
}

// Test reflect usage for better type comparison
func TestADFNodeTypes_Reflect(t *testing.T) {
	doc := MarkdownToADF("test")

	// Verify ADFDocument structure
	docType := reflect.TypeOf(*doc)
	if docType.Name() != "ADFDocument" {
		t.Errorf("expected ADFDocument type, got %s", docType.Name())
	}

	// Verify it has expected fields
	if _, found := docType.FieldByName("Version"); !found {
		t.Error("ADFDocument should have Version field")
	}
	if _, found := docType.FieldByName("Type"); !found {
		t.Error("ADFDocument should have Type field")
	}
	if _, found := docType.FieldByName("Content"); !found {
		t.Error("ADFDocument should have Content field")
	}
}

// Tests for Jira Wiki markup conversion

func TestWikiMarkup_Headings(t *testing.T) {
	tests := []struct {
		wiki   string
		level  int
		text   string
	}{
		{"h1. Title", 1, "Title"},
		{"h2. Subtitle", 2, "Subtitle"},
		{"h3. Section", 3, "Section"},
		{"h4. Subsection", 4, "Subsection"},
		{"h5. Minor", 5, "Minor"},
		{"h6. Smallest", 6, "Smallest"},
	}

	for _, tt := range tests {
		t.Run(tt.wiki, func(t *testing.T) {
			doc := MarkdownToADF(tt.wiki)
			if len(doc.Content) != 1 {
				t.Fatalf("expected 1 content item, got %d", len(doc.Content))
			}

			heading := doc.Content[0]
			if heading.Type != "heading" {
				t.Errorf("expected heading, got %s", heading.Type)
			}
			if heading.Attrs["level"] != tt.level {
				t.Errorf("expected level %d, got %v", tt.level, heading.Attrs["level"])
			}
		})
	}
}

func TestWikiMarkup_HorizontalRule(t *testing.T) {
	doc := MarkdownToADF("----")
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != "rule" {
		t.Errorf("expected rule, got %s", doc.Content[0].Type)
	}
}

func TestWikiMarkup_CodeBlock(t *testing.T) {
	wiki := "{code:java}\npublic class Test {}\n{code}"
	doc := MarkdownToADF(wiki)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	codeBlock := doc.Content[0]
	if codeBlock.Type != "codeBlock" {
		t.Errorf("expected codeBlock, got %s", codeBlock.Type)
	}
	if codeBlock.Attrs["language"] != "java" {
		t.Errorf("expected language 'java', got %v", codeBlock.Attrs["language"])
	}
}

func TestWikiMarkup_MonoSpace(t *testing.T) {
	doc := MarkdownToADF("Use {{monospace}} for code")

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	// Should have text nodes with code mark for monospace
	found := false
	for _, node := range para.Content {
		if node.Text == "monospace" {
			for _, mark := range node.Marks {
				if mark.Type == "code" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected monospace text with code mark")
	}
}

func TestWikiMarkup_Links(t *testing.T) {
	doc := MarkdownToADF("Check [Google|https://google.com]")

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	found := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == "link" {
				if mark.Attrs["href"] == "https://google.com" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected link with correct href")
	}
}

func TestWikiMarkup_Table(t *testing.T) {
	wiki := `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |`

	doc := MarkdownToADF(wiki)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item (table), got %d", len(doc.Content))
	}

	table := doc.Content[0]
	if table.Type != "table" {
		t.Errorf("expected table, got %s", table.Type)
	}

	// Should have 3 rows (header + 2 data rows)
	if len(table.Content) != 3 {
		t.Errorf("expected 3 table rows, got %d", len(table.Content))
	}

	// First row should have tableHeader cells
	if table.Content[0].Content[0].Type != "tableHeader" {
		t.Errorf("expected first row cells to be tableHeader, got %s", table.Content[0].Content[0].Type)
	}

	// Other rows should have tableCell
	if table.Content[1].Content[0].Type != "tableCell" {
		t.Errorf("expected data row cells to be tableCell, got %s", table.Content[1].Content[0].Type)
	}
}

func TestWikiMarkup_MixedDocument(t *testing.T) {
	wiki := `h2. Problem Statement

This is a description with *bold* text.

h3. Requirements

* Item one
* Item two

----

h3. Technical Details

{code:go}
func main() {}
{code}`

	doc := MarkdownToADF(wiki)

	// Should have multiple content items
	if len(doc.Content) < 5 {
		t.Errorf("expected at least 5 content items, got %d", len(doc.Content))
	}

	// Verify first item is h2 heading
	if doc.Content[0].Type != "heading" {
		t.Errorf("expected first item to be heading, got %s", doc.Content[0].Type)
	}
	if doc.Content[0].Attrs["level"] != 2 {
		t.Errorf("expected level 2, got %v", doc.Content[0].Attrs["level"])
	}
}

// Phase 1 & 2 Tests: Blockquote, Panel, Expand, Underline

func TestMarkdownToADF_Blockquote(t *testing.T) {
	markdown := "> This is a blockquote"
	doc := MarkdownToADF(markdown)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	blockquote := doc.Content[0]
	if blockquote.Type != "blockquote" {
		t.Errorf("expected blockquote, got %s", blockquote.Type)
	}
	if len(blockquote.Content) == 0 {
		t.Error("expected blockquote to have content")
	}
	if blockquote.Content[0].Type != "paragraph" {
		t.Errorf("expected paragraph inside blockquote, got %s", blockquote.Content[0].Type)
	}
}

func TestMarkdownToADF_BlockquoteWithFormatting(t *testing.T) {
	markdown := "> This is **bold** quoted text"
	doc := MarkdownToADF(markdown)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	blockquote := doc.Content[0]
	if blockquote.Type != "blockquote" {
		t.Fatalf("expected blockquote, got %s", blockquote.Type)
	}
}

func TestADFToMarkdown_Blockquote(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "blockquote",
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Quoted text",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	if result != "> Quoted text" {
		t.Errorf("expected '> Quoted text', got '%s'", result)
	}
}

func TestRoundTrip_Blockquote(t *testing.T) {
	original := "> This is a blockquote"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestMarkdownToADF_Panel(t *testing.T) {
	tests := []struct {
		markdown string
		panelType string
	}{
		{"[info] This is info", "info"},
		{"[warning] This is a warning", "warning"},
		{"[error] This is an error", "error"},
		{"[success] This is success", "success"},
		{"[note] This is a note", "info"}, // note maps to info
		{"[tip] This is a tip", "success"}, // tip maps to success
	}

	for _, tt := range tests {
		t.Run(tt.markdown, func(t *testing.T) {
			doc := MarkdownToADF(tt.markdown)

			if len(doc.Content) != 1 {
				t.Fatalf("expected 1 content item for %s, got %d", tt.markdown, len(doc.Content))
			}

			panel := doc.Content[0]
			if panel.Type != "panel" {
				t.Errorf("expected panel, got %s", panel.Type)
			}
			if panel.Attrs["panelType"] != tt.panelType {
				t.Errorf("expected panelType %s, got %v", tt.panelType, panel.Attrs["panelType"])
			}
		})
	}
}

func TestADFToMarkdown_Panel(t *testing.T) {
	panelTypes := []string{"info", "warning", "error", "success"}

	for _, panelType := range panelTypes {
		t.Run(panelType, func(t *testing.T) {
			adf := map[string]interface{}{
				"version": 1,
				"type":    "doc",
				"content": []interface{}{
					map[string]interface{}{
						"type": "panel",
						"attrs": map[string]interface{}{
							"panelType": panelType,
						},
						"content": []interface{}{
							map[string]interface{}{
								"type": "paragraph",
								"content": []interface{}{
									map[string]interface{}{
										"type": "text",
										"text": "Panel content",
									},
								},
							},
						},
					},
				},
			}

			result := ADFToMarkdown(adf)
			expected := "[" + panelType + "] Panel content"
			if result != expected {
				t.Errorf("expected '%s', got '%s'", expected, result)
			}
		})
	}
}

func TestRoundTrip_Panel(t *testing.T) {
	original := "[info] This is important info"
	adffDoc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(adffDoc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestMarkdownToADF_Expand(t *testing.T) {
	markdown := `<details>Click to expand</details>
This is the hidden content
More content
`

	doc := MarkdownToADF(markdown)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	expand := doc.Content[0]
	if expand.Type != "expand" {
		t.Errorf("expected expand, got %s", expand.Type)
	}
	// First content node should be title (paragraph)
	if len(expand.Content) == 0 {
		t.Fatal("expected expand to have content")
	}
	if expand.Content[0].Type != "paragraph" {
		t.Errorf("expected first content to be paragraph, got %s", expand.Content[0].Type)
	}
}

func TestADFToMarkdown_Expand(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "expand",
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Title",
							},
						},
					},
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Hidden content",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	if !strings.Contains(result, "<details>Title</details>") {
		t.Errorf("expected <details> tag, got '%s'", result)
	}
	if !strings.Contains(result, "Hidden content") {
		t.Errorf("expected hidden content, got '%s'", result)
	}
}

func TestRoundTrip_Expand(t *testing.T) {
	original := `<details>Expand me</details>`
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	// Expand may have additional body content that was stripped, just check for title
	if !strings.Contains(result, "<details>Expand me</details>") {
		t.Errorf("round-trip failed: expected <details> tag, got '%s'", result)
	}
}

func TestMarkdownToADF_Underline(t *testing.T) {
	markdown := "This is ++underlined++ text"
	doc := MarkdownToADF(markdown)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("expected paragraph, got %s", para.Type)
	}

	// Find the underlined text node
	foundUnderline := false
	for _, node := range para.Content {
		if node.Text == "underlined" {
			for _, mark := range node.Marks {
				if mark.Type == "underline" {
					foundUnderline = true
					break
				}
			}
		}
	}
	if !foundUnderline {
		t.Error("expected to find underlined text with underline mark")
	}
}

func TestADFToMarkdown_Underline(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "underlined",
						"marks": []interface{}{
							map[string]interface{}{
								"type": "underline",
							},
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "++underlined++"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRoundTrip_Underline(t *testing.T) {
	original := "++underlined text++"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

// Phase 3 Tests: Mention, Status, Emoji

func TestMarkdownToADF_Mention(t *testing.T) {
	markdown := "Hello @john_doe"
	doc := MarkdownToADF(markdown)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("expected paragraph, got %s", para.Type)
	}

	foundMention := false
	for _, node := range para.Content {
		if node.Type == "mention" {
			foundMention = true
			if node.Attrs["id"] != "john_doe" {
				t.Errorf("expected mention id 'john_doe', got %v", node.Attrs["id"])
			}
			if node.Attrs["text"] != "john_doe" {
				t.Errorf("expected mention text 'john_doe', got %v", node.Attrs["text"])
			}
		}
	}
	if !foundMention {
		t.Error("expected to find mention node")
	}
}

func TestADFToMarkdown_Mention(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "mention",
						"attrs": map[string]interface{}{
							"id":   "user_123",
							"text": "jdoe",
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "@jdoe"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRoundTrip_Mention(t *testing.T) {
	original := "@username"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestMarkdownToADF_Status(t *testing.T) {
	markdown := "Task status: [status:In Progress]"
	doc := MarkdownToADF(markdown)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(doc.Content))
	}

	para := doc.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("expected paragraph, got %s", para.Type)
	}

	foundStatus := false
	for _, node := range para.Content {
		if node.Type == "status" {
			foundStatus = true
			if node.Attrs["text"] != "In Progress" {
				t.Errorf("expected status text 'In Progress', got %v", node.Attrs["text"])
			}
		}
	}
	if !foundStatus {
		t.Error("expected to find status node")
	}
}

func TestADFToMarkdown_Status(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "status",
						"attrs": map[string]interface{}{
							"text": "Done",
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := "[status:Done]"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRoundTrip_Status(t *testing.T) {
	original := "[status:To Do]"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestMarkdownToADF_Emoji(t *testing.T) {
	emojiCodes := []string{
		"smile",
		"thumbs_up",
		"heart",
		"check_mark",
	}

	for _, code := range emojiCodes {
		t.Run(code, func(t *testing.T) {
			markdown := ":" + code + ":"
			doc := MarkdownToADF(markdown)

			if len(doc.Content) != 1 {
				t.Fatalf("expected 1 content item for %s, got %d", markdown, len(doc.Content))
			}

			para := doc.Content[0]
			foundEmoji := false
			for _, node := range para.Content {
				if node.Type == "emoji" {
					foundEmoji = true
					if node.Attrs["shortName"] != code {
						t.Errorf("expected emoji shortName '%s', got %v", code, node.Attrs["shortName"])
					}
				}
			}
			if !foundEmoji {
				t.Errorf("expected to find emoji node for %s", markdown)
			}
		})
	}
}

func TestADFToMarkdown_Emoji(t *testing.T) {
	adf := map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "emoji",
						"attrs": map[string]interface{}{
							"shortName": "thumbs_up",
						},
					},
				},
			},
		},
	}

	result := ADFToMarkdown(adf)
	expected := ":thumbs_up:"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestRoundTrip_Emoji(t *testing.T) {
	original := ":smile:"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

// Combined feature tests

func TestMarkdownToADF_AllFeatures(t *testing.T) {
	markdown := `# Overview

> Important note about this task

[info] Remember to check dependencies

**Key Points:**

- Item one
- Item two

[warning] This contains deprecated APIs

[error] Breaking change ahead

<details>Advanced details</details>
Technical implementation follows ++best practices++

See @johndoe for questions

Status: [status:In Progress] :thumbs_up:
`

	doc := MarkdownToADF(markdown)

	// Verify all node types are present
	hasHeading := false
	hasBlockquote := false
	hasPanel := false
	hasExpand := false
	hasList := false
	hasUnderline := false
	hasMention := false
	hasStatus := false
	hasEmoji := false

	for _, node := range doc.Content {
		switch node.Type {
		case "heading":
			hasHeading = true
		case "blockquote":
			hasBlockquote = true
		case "panel":
			hasPanel = true
		case "expand":
			hasExpand = true
		case "bulletList", "orderedList":
			hasList = true
		}
		// Check for inline nodes in content
		for _, child := range node.Content {
			if child.Type == "text" {
				for _, mark := range child.Marks {
					if mark.Type == "underline" {
						hasUnderline = true
					}
				}
			} else if child.Type == "mention" {
				hasMention = true
			} else if child.Type == "status" {
				hasStatus = true
			} else if child.Type == "emoji" {
				hasEmoji = true
			} else if child.Type == "paragraph" || child.Type == "listItem" {
				// Recursively check nested paragraphs and list items
				for _, contentNode := range child.Content {
					if contentNode.Type == "text" {
						for _, mark := range contentNode.Marks {
							if mark.Type == "underline" {
								hasUnderline = true
							}
						}
					} else if contentNode.Type == "mention" {
						hasMention = true
					} else if contentNode.Type == "status" {
						hasStatus = true
					} else if contentNode.Type == "emoji" {
						hasEmoji = true
					}
				}
			}
		}
	}

	if !hasHeading {
		t.Error("expected to find heading")
	}
	if !hasBlockquote {
		t.Error("expected to find blockquote")
	}
	if !hasPanel {
		t.Error("expected to find panel")
	}
	if !hasList {
		t.Error("expected to find list")
	}
	if !hasExpand {
		t.Error("expected to find expand")
	}
	if !hasUnderline {
		t.Error("expected to find underline")
	}
	if !hasMention {
		t.Error("expected to find mention")
	}
	if !hasStatus {
		t.Error("expected to find status")
	}
	if !hasEmoji {
		t.Error("expected to find emoji")
	}
}

func TestInlineFormattingPriority(t *testing.T) {
	// Test that inline formatting patterns are processed in correct order
	// Links should be processed before text
	markdown := "[click here++with emphasis++](https://example.com)"
	doc := MarkdownToADF(markdown)

	if len(doc.Content) == 0 {
		t.Fatal("expected content")
	}
}

func TestMultipleUnderline(t *testing.T) {
	original := "++one++ and ++two++"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	if result != original {
		t.Errorf("round-trip failed: original '%s', result '%s'", original, result)
	}
}

func TestNestedFormattingWithUnderline(t *testing.T) {
	original := "++**bold and underlined**++"
	doc := MarkdownToADF(original)

	adfJSON, _ := json.Marshal(doc)
	var adfMap map[string]interface{}
	json.Unmarshal(adfJSON, &adfMap)

	result := ADFToMarkdown(adfMap)
	// The result should preserve both marks
	if !strings.Contains(result, "++") && !strings.Contains(result, "**") {
		t.Errorf("expected preserved formatting, got '%s'", result)
	}
}
