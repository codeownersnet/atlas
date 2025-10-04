package opsgenie

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codeownersnet/atlas/internal/auth"
)

func TestGetOnCalls_WithSchedule(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/schedules/test-schedule-123/on-calls" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"_parent": map[string]interface{}{
					"id":      "test-schedule-123",
					"name":    "Test Schedule",
					"enabled": true,
				},
				"onCallParticipants": []string{"user1@example.com", "user2@example.com"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	authProvider, err := auth.NewAPIKeyAuth("test-api-key")
	if err != nil {
		t.Fatalf("failed to create auth provider: %v", err)
	}

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      authProvider,
		SSLVerify: false,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test GetOnCalls with schedule
	onCalls, err := client.GetOnCalls(context.Background(), "test-schedule-123")
	if err != nil {
		t.Fatalf("GetOnCalls failed: %v", err)
	}

	if len(onCalls) != 1 {
		t.Errorf("expected 1 on-call entry, got %d", len(onCalls))
	}

	// Verify schedule info
	if onCalls[0].ScheduleID != "test-schedule-123" {
		t.Errorf("expected schedule ID 'test-schedule-123', got %s", onCalls[0].ScheduleID)
	}

	if onCalls[0].ScheduleName != "Test Schedule" {
		t.Errorf("expected schedule name 'Test Schedule', got %s", onCalls[0].ScheduleName)
	}

	if len(onCalls[0].OnCallRecipients) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(onCalls[0].OnCallRecipients))
	}

	expectedRecipients := []string{"user1@example.com", "user2@example.com"}
	for i, recipient := range onCalls[0].OnCallRecipients {
		if recipient != expectedRecipients[i] {
			t.Errorf("expected recipient %s, got %s", expectedRecipients[i], recipient)
		}
	}
}

func TestGetOnCalls_AllSchedules(t *testing.T) {
	// Track which endpoints were called
	listSchedulesCalled := false
	schedule1Called := false
	schedule2Called := false

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v2/schedules":
			listSchedulesCalled = true
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "schedule-1", "name": "Schedule 1", "enabled": true},
					{"id": "schedule-2", "name": "Schedule 2", "enabled": true},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/v2/schedules/schedule-1/on-calls":
			schedule1Called = true
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"_parent": map[string]interface{}{
						"id":      "schedule-1",
						"name":    "Schedule 1",
						"enabled": true,
					},
					"onCallParticipants": []string{"user1@example.com"},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/v2/schedules/schedule-2/on-calls":
			schedule2Called = true
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"_parent": map[string]interface{}{
						"id":      "schedule-2",
						"name":    "Schedule 2",
						"enabled": true,
					},
					"onCallParticipants": []string{"user2@example.com"},
				},
			}
			json.NewEncoder(w).Encode(response)

		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client
	authProvider, err := auth.NewAPIKeyAuth("test-api-key")
	if err != nil {
		t.Fatalf("failed to create auth provider: %v", err)
	}

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      authProvider,
		SSLVerify: false,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test GetOnCalls without schedule (all schedules)
	onCalls, err := client.GetOnCalls(context.Background(), "")
	if err != nil {
		t.Fatalf("GetOnCalls failed: %v", err)
	}

	// Verify all endpoints were called
	if !listSchedulesCalled {
		t.Error("ListSchedules was not called")
	}
	if !schedule1Called {
		t.Error("schedule-1 on-calls was not called")
	}
	if !schedule2Called {
		t.Error("schedule-2 on-calls was not called")
	}

	// Verify we got on-calls from both schedules
	if len(onCalls) != 2 {
		t.Errorf("expected 2 on-call entries (one per schedule), got %d", len(onCalls))
	}

	// Verify schedule info for each entry
	scheduleFound := map[string]bool{
		"schedule-1": false,
		"schedule-2": false,
	}

	for _, oc := range onCalls {
		if oc.ScheduleID == "schedule-1" {
			scheduleFound["schedule-1"] = true
			if oc.ScheduleName != "Schedule 1" {
				t.Errorf("expected schedule name 'Schedule 1', got %s", oc.ScheduleName)
			}
		} else if oc.ScheduleID == "schedule-2" {
			scheduleFound["schedule-2"] = true
			if oc.ScheduleName != "Schedule 2" {
				t.Errorf("expected schedule name 'Schedule 2', got %s", oc.ScheduleName)
			}
		} else {
			t.Errorf("unexpected schedule ID: %s", oc.ScheduleID)
		}
	}

	if !scheduleFound["schedule-1"] || !scheduleFound["schedule-2"] {
		t.Error("not all schedules found in on-call results")
	}

	// Verify recipients
	totalRecipients := 0
	for _, oc := range onCalls {
		totalRecipients += len(oc.OnCallRecipients)
	}
	if totalRecipients != 2 {
		t.Errorf("expected 2 total recipients, got %d", totalRecipients)
	}
}

func TestGetOnCalls_EmptyResponse(t *testing.T) {
	// Create mock server that returns empty on-calls
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"_parent": map[string]interface{}{
					"id":      "test-schedule-123",
					"name":    "Test Schedule",
					"enabled": true,
				},
				"onCallParticipants": []string{},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	authProvider, err := auth.NewAPIKeyAuth("test-api-key")
	if err != nil {
		t.Fatalf("failed to create auth provider: %v", err)
	}

	client, err := NewClient(&Config{
		BaseURL:   server.URL,
		Auth:      authProvider,
		SSLVerify: false,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test GetOnCalls with empty response
	onCalls, err := client.GetOnCalls(context.Background(), "test-schedule-123")
	if err != nil {
		t.Fatalf("GetOnCalls failed: %v", err)
	}

	if len(onCalls) != 1 {
		t.Errorf("expected 1 on-call entry, got %d", len(onCalls))
	}

	// Verify schedule info is still populated even with empty recipients
	if onCalls[0].ScheduleID != "test-schedule-123" {
		t.Errorf("expected schedule ID 'test-schedule-123', got %s", onCalls[0].ScheduleID)
	}

	if onCalls[0].ScheduleName != "Test Schedule" {
		t.Errorf("expected schedule name 'Test Schedule', got %s", onCalls[0].ScheduleName)
	}

	if len(onCalls[0].OnCallRecipients) != 0 {
		t.Errorf("expected 0 recipients, got %d", len(onCalls[0].OnCallRecipients))
	}
}
