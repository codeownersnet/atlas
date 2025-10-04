package jira

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAtlassianTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, AtlassianTime)
	}{
		{
			name:    "Atlassian format with milliseconds and +HHMM timezone",
			input:   `"2025-09-24T13:53:18.594+0200"`,
			wantErr: false,
			check: func(t *testing.T, at AtlassianTime) {
				if at.Year() != 2025 || at.Month() != 9 || at.Day() != 24 {
					t.Errorf("wrong date: %v", at.Time)
				}
				if at.Hour() != 13 || at.Minute() != 53 || at.Second() != 18 {
					t.Errorf("wrong time: %v", at.Time)
				}
			},
		},
		{
			name:    "Atlassian format without milliseconds",
			input:   `"2025-09-24T13:53:18+0200"`,
			wantErr: false,
			check: func(t *testing.T, at AtlassianTime) {
				if at.Year() != 2025 || at.Month() != 9 || at.Day() != 24 {
					t.Errorf("wrong date: %v", at.Time)
				}
			},
		},
		{
			name:    "RFC3339 format with Z timezone",
			input:   `"2025-09-24T13:53:18Z"`,
			wantErr: false,
			check: func(t *testing.T, at AtlassianTime) {
				if at.Year() != 2025 || at.Month() != 9 || at.Day() != 24 {
					t.Errorf("wrong date: %v", at.Time)
				}
			},
		},
		{
			name:    "RFC3339Nano format",
			input:   `"2025-09-24T13:53:18.594123Z"`,
			wantErr: false,
			check: func(t *testing.T, at AtlassianTime) {
				if at.Year() != 2025 || at.Month() != 9 || at.Day() != 24 {
					t.Errorf("wrong date: %v", at.Time)
				}
			},
		},
		{
			name:    "Empty string",
			input:   `""`,
			wantErr: false,
			check: func(t *testing.T, at AtlassianTime) {
				if !at.IsZero() {
					t.Errorf("expected zero time, got: %v", at.Time)
				}
			},
		},
		{
			name:    "Invalid format",
			input:   `"not a time"`,
			wantErr: true,
			check:   nil,
		},
		{
			name:    "Negative timezone",
			input:   `"2025-09-24T13:53:18.594-0700"`,
			wantErr: false,
			check: func(t *testing.T, at AtlassianTime) {
				if at.Year() != 2025 {
					t.Errorf("wrong year: %v", at.Time)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var at AtlassianTime
			err := json.Unmarshal([]byte(tt.input), &at)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, at)
			}
		})
	}
}

func TestAtlassianTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		time    AtlassianTime
		want    string
		wantErr bool
	}{
		{
			name: "Normal time without nanoseconds",
			time: AtlassianTime{Time: time.Date(2025, 9, 24, 13, 53, 18, 0, time.UTC)},
			want: `"2025-09-24T13:53:18Z"`,
		},
		{
			name: "Time with nanoseconds",
			time: AtlassianTime{Time: time.Date(2025, 9, 24, 13, 53, 18, 594000000, time.UTC)},
			want: `"2025-09-24T13:53:18.594Z"`,
		},
		{
			name: "Zero time",
			time: AtlassianTime{},
			want: `null`,
		},
		{
			name: "Time with timezone",
			time: AtlassianTime{Time: time.Date(2025, 9, 24, 13, 53, 18, 0, time.FixedZone("CEST", 2*3600))},
			want: `"2025-09-24T13:53:18+02:00"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.time)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestAtlassianTime_String(t *testing.T) {
	tests := []struct {
		name string
		time AtlassianTime
		want string
	}{
		{
			name: "Normal time without nanoseconds",
			time: AtlassianTime{Time: time.Date(2025, 9, 24, 13, 53, 18, 0, time.UTC)},
			want: "2025-09-24T13:53:18Z",
		},
		{
			name: "Time with nanoseconds",
			time: AtlassianTime{Time: time.Date(2025, 9, 24, 13, 53, 18, 594000000, time.UTC)},
			want: "2025-09-24T13:53:18.594Z",
		},
		{
			name: "Zero time",
			time: AtlassianTime{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.time.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtlassianTime_RoundTrip(t *testing.T) {
	// Test that we can unmarshal and marshal back correctly
	original := `"2025-09-24T13:53:18.594+0200"`

	var at AtlassianTime
	err := json.Unmarshal([]byte(original), &at)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	marshaled, err := json.Marshal(at)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal the marshaled value to verify it's valid
	var at2 AtlassianTime
	err = json.Unmarshal(marshaled, &at2)
	if err != nil {
		t.Fatalf("UnmarshalJSON of marshaled value failed: %v", err)
	}

	// Times should be equal (using RFC3339Nano preserves nanoseconds)
	if !at.Time.Equal(at2.Time) {
		t.Errorf("Round trip failed: original %v != final %v", at.Time, at2.Time)
	}
}

func TestIssueFields_TimestampParsing(t *testing.T) {
	// Test that IssueFields correctly unmarshals timestamps
	jsonData := `{
		"summary": "Test Issue",
		"created": "2025-09-24T13:53:18.594+0200",
		"updated": "2025-09-24T14:53:18.594+0200"
	}`

	var fields IssueFields
	err := json.Unmarshal([]byte(jsonData), &fields)
	if err != nil {
		t.Fatalf("Failed to unmarshal IssueFields: %v", err)
	}

	if fields.Summary != "Test Issue" {
		t.Errorf("Summary = %v, want Test Issue", fields.Summary)
	}

	if fields.Created.IsZero() {
		t.Error("Created time should not be zero")
	}

	if fields.Updated.IsZero() {
		t.Error("Updated time should not be zero")
	}

	// Verify times are different
	if fields.Created.Equal(fields.Updated.Time) {
		t.Error("Created and Updated should be different")
	}
}
