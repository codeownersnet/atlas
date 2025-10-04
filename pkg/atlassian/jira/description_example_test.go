package jira

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Example showing how to use the Description type with plain text
func ExampleDescription_plainText() {
	jsonData := `{
		"summary": "Bug in authentication",
		"description": "User cannot login after password reset"
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	if err != nil {
		panic(err)
	}

	fmt.Println("Summary:", fields.Summary)
	if fields.Description != nil {
		fmt.Println("Description:", fields.Description.String())
		fmt.Println("Is ADF:", fields.Description.IsADF())
	}
	// Output:
	// Summary: Bug in authentication
	// Description: User cannot login after password reset
	// Is ADF: false
}

// Example showing how to use the Description type with ADF format
func ExampleDescription_adfFormat() {
	jsonData := `{
		"summary": "Feature request",
		"description": {
			"version": 1,
			"type": "doc",
			"content": [
				{
					"type": "paragraph",
					"content": [
						{
							"type": "text",
							"text": "Add support for "
						},
						{
							"type": "text",
							"text": "OAuth 2.0",
							"marks": [
								{
									"type": "strong"
								}
							]
						}
					]
				},
				{
					"type": "paragraph",
					"content": [
						{
							"type": "text",
							"text": "This will improve security."
						}
					]
				}
			]
		}
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	if err != nil {
		panic(err)
	}

	fmt.Println("Summary:", fields.Summary)
	if fields.Description != nil {
		fmt.Println("Description:", fields.Description.String())
		fmt.Println("Is ADF:", fields.Description.IsADF())
	}
	// Output:
	// Summary: Feature request
	// Description: Add support for  OAuth 2.0
	// This will improve security.
	// Is ADF: true
}

// Test that demonstrates real-world Jira API response handling
func TestRealWorldIssueResponse(t *testing.T) {
	// This simulates a real Jira Cloud API response with ADF description
	issueJSON := `{
		"id": "10001",
		"key": "PROJ-123",
		"fields": {
			"summary": "Integration test issue",
			"description": {
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "This is a detailed description with "
							},
							{
								"type": "text",
								"text": "bold text",
								"marks": [
									{
										"type": "strong"
									}
								]
							},
							{
								"type": "text",
								"text": " and normal text."
							}
						]
					},
					{
						"type": "heading",
						"attrs": {
							"level": 2
						},
						"content": [
							{
								"type": "text",
								"text": "Steps to reproduce"
							}
						]
					},
					{
						"type": "orderedList",
						"content": [
							{
								"type": "listItem",
								"content": [
									{
										"type": "paragraph",
										"content": [
											{
												"type": "text",
												"text": "Step 1"
											}
										]
									}
								]
							},
							{
								"type": "listItem",
								"content": [
									{
										"type": "paragraph",
										"content": [
											{
												"type": "text",
												"text": "Step 2"
											}
										]
									}
								]
							}
						]
					}
				]
			}
		}
	}`

	var issue Issue
	err := json.Unmarshal([]byte(issueJSON), &issue)
	if err != nil {
		t.Fatalf("Failed to unmarshal issue: %v", err)
	}

	if issue.Key != "PROJ-123" {
		t.Errorf("Key = %q, want %q", issue.Key, "PROJ-123")
	}

	if issue.Fields.Summary != "Integration test issue" {
		t.Errorf("Summary = %q, want %q", issue.Fields.Summary, "Integration test issue")
	}

	if issue.Fields.Description == nil {
		t.Fatal("Description is nil")
	}

	if !issue.Fields.Description.IsADF() {
		t.Error("Expected ADF format, got plain text")
	}

	text := issue.Fields.Description.String()
	if text == "" {
		t.Error("Extracted text is empty")
	}

	// Verify text extraction captured the main content
	expectedSubstrings := []string{
		"This is a detailed description with",
		"bold text",
		"Steps to reproduce",
		"Step 1",
		"Step 2",
	}

	for _, substr := range expectedSubstrings {
		if !containsString(text, substr) {
			t.Errorf("Extracted text doesn't contain %q. Got: %q", substr, text)
		}
	}
}

// Helper function for substring checking
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test that demonstrates handling of issues without descriptions
func TestIssueWithoutDescription(t *testing.T) {
	issueJSON := `{
		"id": "10002",
		"key": "PROJ-124",
		"fields": {
			"summary": "Issue without description"
		}
	}`

	var issue Issue
	err := json.Unmarshal([]byte(issueJSON), &issue)
	if err != nil {
		t.Fatalf("Failed to unmarshal issue: %v", err)
	}

	if issue.Fields.Description != nil {
		t.Error("Expected nil description, got non-nil")
	}
}
