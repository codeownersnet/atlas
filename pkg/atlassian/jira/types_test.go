package jira

import (
	"encoding/json"
	"testing"
)

func TestDescription_UnmarshalJSON_PlainText(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantADF bool
		wantErr bool
	}{
		{
			name:    "plain text string",
			json:    `"This is a plain text description"`,
			want:    "This is a plain text description",
			wantADF: false,
			wantErr: false,
		},
		{
			name:    "empty string",
			json:    `""`,
			want:    "",
			wantADF: false,
			wantErr: false,
		},
		{
			name: "ADF simple paragraph",
			json: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "This is the description"
							}
						]
					}
				]
			}`,
			want:    "This is the description",
			wantADF: true,
			wantErr: false,
		},
		{
			name: "ADF multiple paragraphs",
			json: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "First paragraph"
							}
						]
					},
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "Second paragraph"
							}
						]
					}
				]
			}`,
			want:    "First paragraph\nSecond paragraph",
			wantADF: true,
			wantErr: false,
		},
		{
			name: "ADF with nested content",
			json: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "Some "
							},
							{
								"type": "text",
								"text": "nested "
							},
							{
								"type": "text",
								"text": "text"
							}
						]
					}
				]
			}`,
			want:    "Some  nested  text",
			wantADF: true,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid}`,
			want:    "",
			wantADF: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var desc Description
			err := json.Unmarshal([]byte(tt.json), &desc)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got := desc.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}

			if got := desc.IsADF(); got != tt.wantADF {
				t.Errorf("IsADF() = %v, want %v", got, tt.wantADF)
			}

			if desc.Raw() == nil {
				t.Error("Raw() returned nil, want non-nil")
			}
		})
	}
}

func TestDescription_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "plain text",
			input:   `"Plain text description"`,
			want:    `"Plain text description"`,
			wantErr: false,
		},
		{
			name: "ADF format",
			input: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "ADF text"
							}
						]
					}
				]
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var desc Description
			err := json.Unmarshal([]byte(tt.input), &desc)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			got, err := json.Marshal(desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// For ADF, just verify it's valid JSON and contains the version field
			if desc.IsADF() {
				var adfObj map[string]interface{}
				if err := json.Unmarshal(got, &adfObj); err != nil {
					t.Errorf("MarshalJSON() produced invalid JSON: %v", err)
				}
				if _, hasVersion := adfObj["version"]; !hasVersion {
					t.Error("MarshalJSON() ADF object missing version field")
				}
			} else {
				// For plain text, verify exact match
				if string(got) != tt.want {
					t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
				}
			}
		})
	}
}

func TestDescription_NilPointer(t *testing.T) {
	var desc *Description

	if got := desc.String(); got != "" {
		t.Errorf("nil Description.String() = %q, want empty string", got)
	}

	if got := desc.IsADF(); got != false {
		t.Errorf("nil Description.IsADF() = %v, want false", got)
	}

	if got := desc.Raw(); got != nil {
		t.Errorf("nil Description.Raw() = %v, want nil", got)
	}
}

func TestIssueFields_DescriptionUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantSummary string
		wantDesc    string
		wantADF     bool
		wantErr     bool
	}{
		{
			name: "issue with plain text description",
			json: `{
				"summary": "Test Issue",
				"description": "Plain text description"
			}`,
			wantSummary: "Test Issue",
			wantDesc:    "Plain text description",
			wantADF:     false,
			wantErr:     false,
		},
		{
			name: "issue with ADF description",
			json: `{
				"summary": "Test Issue",
				"description": {
					"version": 1,
					"type": "doc",
					"content": [
						{
							"type": "paragraph",
							"content": [
								{
									"type": "text",
									"text": "ADF description"
								}
							]
						}
					]
				}
			}`,
			wantSummary: "Test Issue",
			wantDesc:    "ADF description",
			wantADF:     true,
			wantErr:     false,
		},
		{
			name: "issue without description",
			json: `{
				"summary": "Test Issue"
			}`,
			wantSummary: "Test Issue",
			wantDesc:    "",
			wantADF:     false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fields IssueFields
			err := json.Unmarshal([]byte(tt.json), &fields)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if fields.Summary != tt.wantSummary {
				t.Errorf("Summary = %q, want %q", fields.Summary, tt.wantSummary)
			}

			if fields.Description == nil && tt.wantDesc != "" {
				t.Error("Description is nil, expected non-nil")
				return
			}

			if fields.Description != nil {
				if got := fields.Description.String(); got != tt.wantDesc {
					t.Errorf("Description.String() = %q, want %q", got, tt.wantDesc)
				}

				if got := fields.Description.IsADF(); got != tt.wantADF {
					t.Errorf("Description.IsADF() = %v, want %v", got, tt.wantADF)
				}
			}
		})
	}
}

func TestExtractTextFromADF(t *testing.T) {
	tests := []struct {
		name string
		adf  map[string]interface{}
		want string
	}{
		{
			name: "simple text node",
			adf: map[string]interface{}{
				"type": "text",
				"text": "Hello",
			},
			want: "Hello",
		},
		{
			name: "paragraph with text",
			adf: map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Hello world",
					},
				},
			},
			want: "Hello world",
		},
		{
			name: "multiple paragraphs",
			adf: map[string]interface{}{
				"type": "doc",
				"content": []interface{}{
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "First",
							},
						},
					},
					map[string]interface{}{
						"type": "paragraph",
						"content": []interface{}{
							map[string]interface{}{
								"type": "text",
								"text": "Second",
							},
						},
					},
				},
			},
			want: "First\nSecond",
		},
		{
			name: "empty content",
			adf: map[string]interface{}{
				"type": "doc",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTextFromADF(tt.adf)
			if got != tt.want {
				t.Errorf("extractTextFromADF() = %q, want %q", got, tt.want)
			}
		})
	}
}
