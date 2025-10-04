package opsgenie

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codeownersnet/atlas/internal/mcp"
	"github.com/codeownersnet/atlas/pkg/atlassian/opsgenie"
)

// OpsgenieCreateAlertTool creates the opsgenie_create_alert tool
func OpsgenieCreateAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_create_alert",
		"Create a new Opsgenie alert. Alerts are notifications for specific events or issues. Requires message, and can include description, priority, responders, and tags.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"message":     mcp.NewStringProperty("Brief message describing the alert (required)"),
				"description": mcp.NewStringProperty("Detailed description of the alert"),
				"priority": mcp.NewStringProperty("Priority level (P1, P2, P3, P4, P5 - default P3)").
					WithDefault("P3"),
				"responders": mcp.NewStringProperty("JSON string of responders array. Each responder should have 'type' (user/team/escalation/schedule) and 'id'. Example: '[{\"type\":\"user\",\"id\":\"user-id\"},{\"type\":\"team\",\"id\":\"team-id\"}]'"),
				"tags":       mcp.NewStringProperty("Comma-separated tags to categorize the alert"),
			},
			"message",
		),
		opsgenieCreateAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieCreateAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return nil, fmt.Errorf("message is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Build alert request
	req := &opsgenie.AlertRequest{
		Message: message,
	}

	// Add optional description
	if desc, ok := args["description"].(string); ok && desc != "" {
		req.Description = desc
	}

	// Add priority (default to P3)
	priority := "P3"
	if p, ok := args["priority"].(string); ok && p != "" {
		priority = p
	}
	req.Priority = opsgenie.Priority(priority)

	// Add responders (accept either JSON string or array)
	if respondersStr, ok := args["responders"].(string); ok && respondersStr != "" {
		var respondersList []map[string]interface{}
		if err := json.Unmarshal([]byte(respondersStr), &respondersList); err == nil {
			responders := make([]opsgenie.Responder, 0, len(respondersList))
			for _, respMap := range respondersList {
				responder := opsgenie.Responder{}
				if rType, ok := respMap["type"].(string); ok {
					responder.Type = opsgenie.ResponderType(rType)
				}
				if id, ok := respMap["id"].(string); ok {
					responder.ID = id
				}
				if name, ok := respMap["name"].(string); ok {
					responder.Name = name
				}
				if responder.Type != "" && responder.ID != "" {
					responders = append(responders, responder)
				}
			}
			if len(responders) > 0 {
				req.Responders = responders
			}
		}
	}

	// Add tags (accept comma-separated string)
	if tagsStr, ok := args["tags"].(string); ok && tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		trimmedTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				trimmedTags = append(trimmedTags, trimmed)
			}
		}
		if len(trimmedTags) > 0 {
			req.Tags = trimmedTags
		}
	}

	// Create alert
	alert, err := client.CreateAlert(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return mcp.NewJSONResult(alert)
}

// OpsgenieCloseAlertTool creates the opsgenie_close_alert tool
func OpsgenieCloseAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_close_alert",
		"Close an Opsgenie alert by ID. Optionally add a note explaining the closure reason.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Alert ID to close (required)"),
				"note": mcp.NewStringProperty("Optional note explaining the closure reason"),
			},
			"id",
		),
		opsgenieCloseAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieCloseAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.CloseAlert(ctx, id, note)
	if err != nil {
		return nil, fmt.Errorf("failed to close alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Alert %s closed successfully", id),
	})
}

// OpsgenieAcknowledgeAlertTool creates the opsgenie_acknowledge_alert tool
func OpsgenieAcknowledgeAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_acknowledge_alert",
		"Acknowledge an Opsgenie alert by ID. Optionally add a note explaining the acknowledgment.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Alert ID to acknowledge (required)"),
				"note": mcp.NewStringProperty("Optional note explaining the acknowledgment"),
			},
			"id",
		),
		opsgenieAcknowledgeAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieAcknowledgeAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.AcknowledgeAlert(ctx, id, note)
	if err != nil {
		return nil, fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Alert %s acknowledged successfully", id),
	})
}

// OpsgenieSnoozeAlertTool creates the opsgenie_snooze_alert tool
func OpsgenieSnoozeAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_snooze_alert",
		"Snooze an Opsgenie alert by ID until a specified end time. The alert will be temporarily suppressed and automatically reactivated at the end time.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":       mcp.NewStringProperty("Alert ID to snooze (required)"),
				"end_time": mcp.NewStringProperty("End time for snooze in ISO 8601 format (e.g., 2024-01-01T12:00:00Z) (required)"),
				"note":     mcp.NewStringProperty("Optional note explaining the snooze reason"),
			},
			"id", "end_time",
		),
		opsgenieSnoozeAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieSnoozeAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	endTime, ok := args["end_time"].(string)
	if !ok || endTime == "" {
		return nil, fmt.Errorf("end_time is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.SnoozeAlert(ctx, id, endTime, note)
	if err != nil {
		return nil, fmt.Errorf("failed to snooze alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Alert %s snoozed successfully until %s", id, endTime),
	})
}

// OpsgenieEscalateAlertTool creates the opsgenie_escalate_alert tool
func OpsgenieEscalateAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_escalate_alert",
		"Escalate an Opsgenie alert to a specified escalation policy, team, or user. Use this to route alerts to appropriate responders.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":             mcp.NewStringProperty("Alert ID to escalate (required)"),
				"responder_type": mcp.NewStringProperty("Responder type: user, team, escalation, or schedule (required)"),
				"responder_id":   mcp.NewStringProperty("Responder ID (required)"),
				"responder_name": mcp.NewStringProperty("Responder name (optional)"),
				"note":           mcp.NewStringProperty("Optional note explaining the escalation reason"),
			},
			"id", "responder_type", "responder_id",
		),
		opsgenieEscalateAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieEscalateAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	responderType, ok := args["responder_type"].(string)
	if !ok || responderType == "" {
		return nil, fmt.Errorf("responder_type is required")
	}

	responderID, ok := args["responder_id"].(string)
	if !ok || responderID == "" {
		return nil, fmt.Errorf("responder_id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Build responder object
	responder := &opsgenie.Responder{
		Type: opsgenie.ResponderType(responderType),
		ID:   responderID,
	}

	if name, ok := args["responder_name"].(string); ok && name != "" {
		responder.Name = name
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.EscalateAlert(ctx, id, responder, note)
	if err != nil {
		return nil, fmt.Errorf("failed to escalate alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Alert %s escalated successfully", id),
	})
}

// OpsgenieAssignAlertTool creates the opsgenie_assign_alert tool
func OpsgenieAssignAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_assign_alert",
		"Assign an Opsgenie alert to a specific user or team. Use this to designate ownership of an alert.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":             mcp.NewStringProperty("Alert ID to assign (required)"),
				"responder_type": mcp.NewStringProperty("Responder type: user or team (required)"),
				"responder_id":   mcp.NewStringProperty("Responder ID (required)"),
				"responder_name": mcp.NewStringProperty("Responder name (optional)"),
				"note":           mcp.NewStringProperty("Optional note explaining the assignment"),
			},
			"id", "responder_type", "responder_id",
		),
		opsgenieAssignAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieAssignAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	responderType, ok := args["responder_type"].(string)
	if !ok || responderType == "" {
		return nil, fmt.Errorf("responder_type is required")
	}

	responderID, ok := args["responder_id"].(string)
	if !ok || responderID == "" {
		return nil, fmt.Errorf("responder_id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Build responder object
	responder := &opsgenie.Responder{
		Type: opsgenie.ResponderType(responderType),
		ID:   responderID,
	}

	if name, ok := args["responder_name"].(string); ok && name != "" {
		responder.Name = name
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.AssignAlert(ctx, id, responder, note)
	if err != nil {
		return nil, fmt.Errorf("failed to assign alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Alert %s assigned successfully", id),
	})
}

// OpsgenieAddNoteToAlertTool creates the opsgenie_add_note_to_alert tool
func OpsgenieAddNoteToAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_add_note_to_alert",
		"Add a note to an existing Opsgenie alert by ID. Use this to document alert progress, investigation findings, or resolution details.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Alert ID to add note to (required)"),
				"note": mcp.NewStringProperty("Note text to add to the alert (required)"),
			},
			"id", "note",
		),
		opsgenieAddNoteToAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieAddNoteToAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	note, ok := args["note"].(string)
	if !ok || note == "" {
		return nil, fmt.Errorf("note is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	err := client.AddNoteToAlert(ctx, id, note)
	if err != nil {
		return nil, fmt.Errorf("failed to add note to alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Note added to alert %s successfully", id),
	})
}

// OpsgenieAddTagsToAlertTool creates the opsgenie_add_tags_to_alert tool
func OpsgenieAddTagsToAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_add_tags_to_alert",
		"Add tags to an existing Opsgenie alert by ID. Tags help categorize and filter alerts. Provide tags as a comma-separated string.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Alert ID to add tags to (required)"),
				"tags": mcp.NewStringProperty("Comma-separated tags to add to the alert (required)"),
				"note": mcp.NewStringProperty("Optional note explaining the tag addition"),
			},
			"id", "tags",
		),
		opsgenieAddTagsToAlertHandler,
		"opsgenie", "write",
	)
}

func opsgenieAddTagsToAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	tagsStr, ok := args["tags"].(string)
	if !ok || tagsStr == "" {
		return nil, fmt.Errorf("tags is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Split tags by comma and trim whitespace
	tags := strings.Split(tagsStr, ",")
	trimmedTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			trimmedTags = append(trimmedTags, trimmed)
		}
	}

	if len(trimmedTags) == 0 {
		return nil, fmt.Errorf("no valid tags provided")
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.AddTagsToAlert(ctx, id, trimmedTags, note)
	if err != nil {
		return nil, fmt.Errorf("failed to add tags to alert: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Tags added to alert %s successfully", id),
	})
}

// OpsgenieCreateIncidentTool creates the opsgenie_create_incident tool
func OpsgenieCreateIncidentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_create_incident",
		"Create a new Opsgenie incident. Incidents are major issues affecting multiple services or users. Requires message, description, priority, and can include responders (array of objects with type and id fields) and tags.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"message":     mcp.NewStringProperty("Brief message describing the incident (required)"),
				"description": mcp.NewStringProperty("Detailed description of the incident"),
				"priority": mcp.NewStringProperty("Priority level (P1, P2, P3, P4, P5 - default P3)").
					WithDefault("P3"),
				"responders": mcp.NewStringProperty("JSON string of responders array. Each responder should have 'type' (user/team/escalation/schedule) and 'id'. Example: '[{\"type\":\"user\",\"id\":\"user-id\"},{\"type\":\"team\",\"id\":\"team-id\"}]'"),
				"tags":       mcp.NewStringProperty("Comma-separated tags to categorize the incident"),
			},
			"message",
		),
		opsgenieCreateIncidentHandler,
		"opsgenie", "write",
	)
}

func opsgenieCreateIncidentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return nil, fmt.Errorf("message is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Build incident request
	req := &opsgenie.IncidentRequest{
		Message: message,
	}

	// Add optional description
	if desc, ok := args["description"].(string); ok && desc != "" {
		req.Description = desc
	}

	// Add priority (default to P3)
	priority := "P3"
	if p, ok := args["priority"].(string); ok && p != "" {
		priority = p
	}
	req.Priority = opsgenie.Priority(priority)

	// Add responders (accept either JSON string or array)
	if respondersStr, ok := args["responders"].(string); ok && respondersStr != "" {
		var respondersList []map[string]interface{}
		if err := json.Unmarshal([]byte(respondersStr), &respondersList); err == nil {
			responders := make([]opsgenie.Responder, 0, len(respondersList))
			for _, respMap := range respondersList {
				responder := opsgenie.Responder{}
				if rType, ok := respMap["type"].(string); ok {
					responder.Type = opsgenie.ResponderType(rType)
				}
				if id, ok := respMap["id"].(string); ok {
					responder.ID = id
				}
				if name, ok := respMap["name"].(string); ok {
					responder.Name = name
				}
				if responder.Type != "" && responder.ID != "" {
					responders = append(responders, responder)
				}
			}
			if len(responders) > 0 {
				req.Responders = responders
			}
		}
	}

	// Add tags (accept comma-separated string)
	if tagsStr, ok := args["tags"].(string); ok && tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		trimmedTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				trimmedTags = append(trimmedTags, trimmed)
			}
		}
		if len(trimmedTags) > 0 {
			req.Tags = trimmedTags
		}
	}

	// Create incident
	incident, err := client.CreateIncident(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return mcp.NewJSONResult(incident)
}

// OpsgenieCloseIncidentTool creates the opsgenie_close_incident tool
func OpsgenieCloseIncidentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_close_incident",
		"Close an Opsgenie incident by ID. Optionally add a note explaining the closure reason.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Incident ID to close (required)"),
				"note": mcp.NewStringProperty("Optional note explaining the closure reason"),
			},
			"id",
		),
		opsgenieCloseIncidentHandler,
		"opsgenie", "write",
	)
}

func opsgenieCloseIncidentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	note := ""
	if n, ok := args["note"].(string); ok {
		note = n
	}

	err := client.CloseIncident(ctx, id, note)
	if err != nil {
		return nil, fmt.Errorf("failed to close incident: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Incident %s closed successfully", id),
	})
}

// OpsgenieAddNoteToIncidentTool creates the opsgenie_add_note_to_incident tool
func OpsgenieAddNoteToIncidentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_add_note_to_incident",
		"Add a note to an existing Opsgenie incident by ID. Use this to document incident progress, investigation findings, or resolution details.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Incident ID to add note to (required)"),
				"note": mcp.NewStringProperty("Note text to add to the incident (required)"),
			},
			"id", "note",
		),
		opsgenieAddNoteToIncidentHandler,
		"opsgenie", "write",
	)
}

func opsgenieAddNoteToIncidentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	note, ok := args["note"].(string)
	if !ok || note == "" {
		return nil, fmt.Errorf("note is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	err := client.AddNoteToIncident(ctx, id, note)
	if err != nil {
		return nil, fmt.Errorf("failed to add note to incident: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Note added to incident %s successfully", id),
	})
}

// OpsgenieAddResponderToIncidentTool creates the opsgenie_add_responder_to_incident tool
func OpsgenieAddResponderToIncidentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_add_responder_to_incident",
		"Add a responder to an existing Opsgenie incident by ID. Responder can be a user, team, escalation policy, or schedule. Provide responder_type and responder_id.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":             mcp.NewStringProperty("Incident ID to add responder to (required)"),
				"responder_type": mcp.NewStringProperty("Responder type: user, team, escalation, or schedule (required)"),
				"responder_id":   mcp.NewStringProperty("Responder ID (required)"),
				"responder_name": mcp.NewStringProperty("Responder name (optional)"),
			},
			"id", "responder_type", "responder_id",
		),
		opsgenieAddResponderToIncidentHandler,
		"opsgenie", "write",
	)
}

func opsgenieAddResponderToIncidentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	responderType, ok := args["responder_type"].(string)
	if !ok || responderType == "" {
		return nil, fmt.Errorf("responder_type is required")
	}

	responderID, ok := args["responder_id"].(string)
	if !ok || responderID == "" {
		return nil, fmt.Errorf("responder_id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Build responder object
	responder := &opsgenie.Responder{
		Type: opsgenie.ResponderType(responderType),
		ID:   responderID,
	}

	if name, ok := args["responder_name"].(string); ok && name != "" {
		responder.Name = name
	}

	err := client.AddResponderToIncident(ctx, id, responder)
	if err != nil {
		return nil, fmt.Errorf("failed to add responder to incident: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Responder added to incident %s successfully", id),
	})
}
