package opsgenie

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeownersnet/atlas/internal/mcp"
)

// toJSON converts a value to a JSON string
func toJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"failed to marshal JSON: %v\"}", err)
	}
	return string(data)
}

// OpsgenieGetAlertTool creates the opsgenie_get_alert tool
func OpsgenieGetAlertTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_alert",
		"Get detailed information about an Opsgenie alert by ID or alias.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("Alert ID or alias to retrieve"),
			},
			"id",
		),
		opsgenieGetAlertHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetAlertHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	alert, err := client.GetAlert(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return mcp.NewJSONResult(alert)
}

// OpsgenieListAlertsTool creates the opsgenie_list_alerts tool
func OpsgenieListAlertsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_list_alerts",
		"List and search Opsgenie alerts with optional query filtering and pagination. Query syntax supports field:value pairs (e.g., 'status:open priority:P1').",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"query": mcp.NewStringProperty("Search query to filter alerts (e.g., 'status:open', 'priority:P1'). Leave empty to list all."),
				"limit": mcp.NewIntegerProperty("Maximum number of alerts to return (default 20, max 100)").
					WithDefault(20),
				"offset": mcp.NewIntegerProperty("Number of alerts to skip for pagination (default 0)").
					WithDefault(0),
			},
		),
		opsgenieListAlertsHandler,
		"opsgenie", "read",
	)
}

func opsgenieListAlertsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	query := ""
	if q, ok := args["query"].(string); ok {
		query = q
	}

	limit := getIntArg(args, "limit", 20)
	offset := getIntArg(args, "offset", 0)

	result, err := client.ListAlerts(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// OpsgenieCountAlertsTool creates the opsgenie_count_alerts tool
func OpsgenieCountAlertsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_count_alerts",
		"Get the count of Opsgenie alerts matching a query. Useful for understanding alert volume without fetching all alert details.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"query": mcp.NewStringProperty("Search query to filter alerts (e.g., 'status:open', 'priority:P1'). Leave empty to count all."),
			},
		),
		opsgenieCountAlertsHandler,
		"opsgenie", "read",
	)
}

func opsgenieCountAlertsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	query := ""
	if q, ok := args["query"].(string); ok {
		query = q
	}

	count, err := client.CountAlerts(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count alerts: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"count": count,
		"query": query,
	})
}

// OpsgenieGetRequestStatusTool creates the opsgenie_get_request_status tool
func OpsgenieGetRequestStatusTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_request_status",
		"Get the status of an asynchronous Opsgenie request by request ID. Used to track the completion status of long-running operations.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"request_id": mcp.NewStringProperty("The request ID returned from an asynchronous operation"),
			},
			"request_id",
		),
		opsgenieGetRequestStatusHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetRequestStatusHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	requestID, ok := args["request_id"].(string)
	if !ok || requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	status, err := client.GetRequestStatus(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get request status: %w", err)
	}

	return mcp.NewJSONResult(status)
}

// OpsgenieGetIncidentTool creates the opsgenie_get_incident tool
func OpsgenieGetIncidentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_incident",
		"Get detailed information about an Opsgenie incident by ID. Incidents are major issues affecting multiple services or users.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("Incident ID to retrieve"),
			},
			"id",
		),
		opsgenieGetIncidentHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetIncidentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	incident, err := client.GetIncident(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return mcp.NewJSONResult(incident)
}

// OpsgenieListIncidentsTool creates the opsgenie_list_incidents tool
func OpsgenieListIncidentsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_list_incidents",
		"List and search Opsgenie incidents with optional query filtering and pagination. Query syntax supports field:value pairs (e.g., 'status:open priority:P1').",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"query": mcp.NewStringProperty("Search query to filter incidents (e.g., 'status:open', 'priority:P1'). Leave empty to list all."),
				"limit": mcp.NewIntegerProperty("Maximum number of incidents to return (default 20, max 100)").
					WithDefault(20),
				"offset": mcp.NewIntegerProperty("Number of incidents to skip for pagination (default 0)").
					WithDefault(0),
			},
		),
		opsgenieListIncidentsHandler,
		"opsgenie", "read",
	)
}

func opsgenieListIncidentsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	query := ""
	if q, ok := args["query"].(string); ok {
		query = q
	}

	limit := getIntArg(args, "limit", 20)
	offset := getIntArg(args, "offset", 0)

	result, err := client.ListIncidents(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// OpsgenieGetScheduleTool creates the opsgenie_get_schedule tool
func OpsgenieGetScheduleTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_schedule",
		"Get detailed information about an Opsgenie schedule by ID. Returns schedule configuration including rotations and participants.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("Schedule ID to retrieve"),
			},
			"id",
		),
		opsgenieGetScheduleHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetScheduleHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	schedule, err := client.GetSchedule(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	return mcp.NewJSONResult(schedule)
}

// OpsgenieListSchedulesTool creates the opsgenie_list_schedules tool
func OpsgenieListSchedulesTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_list_schedules",
		"List all Opsgenie schedules. Returns basic schedule information including ID, name, and status.",
		mcp.NewInputSchema(
			map[string]mcp.Property{},
		),
		opsgenieListSchedulesHandler,
		"opsgenie", "read",
	)
}

func opsgenieListSchedulesHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	schedules, err := client.ListSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}

	return mcp.NewJSONResult(schedules)
}

// OpsgenieGetScheduleTimelineTool creates the opsgenie_get_schedule_timeline tool
func OpsgenieGetScheduleTimelineTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_schedule_timeline",
		"Get the timeline for an Opsgenie schedule within a specified time range. Shows who is on-call during each period. Dates must be in ISO 8601 format (e.g., '2024-01-15T00:00:00Z').",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id":   mcp.NewStringProperty("Schedule ID to retrieve timeline for"),
				"from": mcp.NewStringProperty("Start date in ISO 8601 format (e.g., '2024-01-15T00:00:00Z')"),
				"to":   mcp.NewStringProperty("End date in ISO 8601 format (e.g., '2024-01-22T00:00:00Z')"),
			},
			"id", "from", "to",
		),
		opsgenieGetScheduleTimelineHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetScheduleTimelineHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	fromStr, ok := args["from"].(string)
	if !ok || fromStr == "" {
		return nil, fmt.Errorf("from date is required")
	}

	toStr, ok := args["to"].(string)
	if !ok || toStr == "" {
		return nil, fmt.Errorf("to date is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	// Parse ISO 8601 dates
	from, err := parseISO8601(fromStr)
	if err != nil {
		return nil, fmt.Errorf("invalid from date format (use ISO 8601, e.g., '2024-01-15T00:00:00Z'): %w", err)
	}

	to, err := parseISO8601(toStr)
	if err != nil {
		return nil, fmt.Errorf("invalid to date format (use ISO 8601, e.g., '2024-01-22T00:00:00Z'): %w", err)
	}

	timeline, err := client.GetScheduleTimeline(ctx, id, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule timeline: %w", err)
	}

	return mcp.NewJSONResult(timeline)
}

// OpsgenieGetOnCallsTool creates the opsgenie_get_on_calls tool
func OpsgenieGetOnCallsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_on_calls",
		"Get current on-call users. If schedule ID is provided, returns on-calls for that schedule only. If no schedule is provided, returns on-calls for all schedules.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"schedule": mcp.NewStringProperty("Optional schedule ID to filter on-call users by specific schedule. Leave empty to get on-calls for all schedules."),
			},
		),
		opsgenieGetOnCallsHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetOnCallsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	schedule := ""
	if s, ok := args["schedule"].(string); ok {
		schedule = s
	}

	onCalls, err := client.GetOnCalls(ctx, schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to get on-calls: %w", err)
	}

	return mcp.NewJSONResult(onCalls)
}

// OpsgenieGetTeamTool creates the opsgenie_get_team tool
func OpsgenieGetTeamTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_team",
		"Get detailed information about an Opsgenie team by ID or name. Returns team details including members and their roles.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"id": mcp.NewStringProperty("Team ID or name to retrieve"),
			},
			"id",
		),
		opsgenieGetTeamHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetTeamHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	team, err := client.GetTeam(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return mcp.NewJSONResult(team)
}

// OpsgenieListTeamsTool creates the opsgenie_list_teams tool
func OpsgenieListTeamsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_list_teams",
		"List all Opsgenie teams. Returns basic team information including ID, name, and description.",
		mcp.NewInputSchema(
			map[string]mcp.Property{},
		),
		opsgenieListTeamsHandler,
		"opsgenie", "read",
	)
}

func opsgenieListTeamsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	teams, err := client.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return mcp.NewJSONResult(teams)
}

// OpsgenieGetUserTool creates the opsgenie_get_user tool
func OpsgenieGetUserTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"opsgenie_get_user",
		"Get detailed information about an Opsgenie user by identifier (ID, username, or email). Returns user profile including role, timezone, and status.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"identifier": mcp.NewStringProperty("User identifier (ID, username, or email address)"),
			},
			"identifier",
		),
		opsgenieGetUserHandler,
		"opsgenie", "read",
	)
}

func opsgenieGetUserHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	identifier, ok := args["identifier"].(string)
	if !ok || identifier == "" {
		return nil, fmt.Errorf("identifier is required")
	}

	client := GetOpsgenieClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Opsgenie client not available")
	}

	user, err := client.GetUser(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return mcp.NewJSONResult(user)
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
		}
	}
	return defaultVal
}

// Helper function to parse ISO 8601 date string
func parseISO8601(dateStr string) (time.Time, error) {
	// Try common ISO 8601 formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date '%s' using ISO 8601 format", dateStr)
}
