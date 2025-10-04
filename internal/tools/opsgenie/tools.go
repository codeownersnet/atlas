package opsgenie

import (
	"fmt"

	"github.com/codeownersnet/atlas/internal/mcp"
)

// RegisterOpsgenieTools registers all Opsgenie tools with the MCP server
func RegisterOpsgenieTools(server *mcp.Server) error {
	tools := []struct {
		name string
		tool *mcp.ToolDefinition
	}{
		// Read operations (13 tools)
		{"opsgenie_get_alert", OpsgenieGetAlertTool()},
		{"opsgenie_list_alerts", OpsgenieListAlertsTool()},
		{"opsgenie_count_alerts", OpsgenieCountAlertsTool()},
		{"opsgenie_get_request_status", OpsgenieGetRequestStatusTool()},
		{"opsgenie_get_incident", OpsgenieGetIncidentTool()},
		{"opsgenie_list_incidents", OpsgenieListIncidentsTool()},
		{"opsgenie_get_schedule", OpsgenieGetScheduleTool()},
		{"opsgenie_list_schedules", OpsgenieListSchedulesTool()},
		{"opsgenie_get_schedule_timeline", OpsgenieGetScheduleTimelineTool()},
		{"opsgenie_get_on_calls", OpsgenieGetOnCallsTool()},
		{"opsgenie_get_team", OpsgenieGetTeamTool()},
		{"opsgenie_list_teams", OpsgenieListTeamsTool()},
		{"opsgenie_get_user", OpsgenieGetUserTool()},

		// Write operations (12 tools)
		{"opsgenie_create_alert", OpsgenieCreateAlertTool()},
		{"opsgenie_close_alert", OpsgenieCloseAlertTool()},
		{"opsgenie_acknowledge_alert", OpsgenieAcknowledgeAlertTool()},
		{"opsgenie_snooze_alert", OpsgenieSnoozeAlertTool()},
		{"opsgenie_escalate_alert", OpsgenieEscalateAlertTool()},
		{"opsgenie_assign_alert", OpsgenieAssignAlertTool()},
		{"opsgenie_add_note_to_alert", OpsgenieAddNoteToAlertTool()},
		{"opsgenie_add_tags_to_alert", OpsgenieAddTagsToAlertTool()},
		{"opsgenie_create_incident", OpsgenieCreateIncidentTool()},
		{"opsgenie_close_incident", OpsgenieCloseIncidentTool()},
		{"opsgenie_add_note_to_incident", OpsgenieAddNoteToIncidentTool()},
		{"opsgenie_add_responder_to_incident", OpsgenieAddResponderToIncidentTool()},
	}

	for _, t := range tools {
		if err := server.RegisterTool(t.tool); err != nil {
			return fmt.Errorf("failed to register %s: %w", t.name, err)
		}
	}

	return nil
}
