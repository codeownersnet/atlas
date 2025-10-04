package opsgenie

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codeownersnet/atlas/internal/auth"
	"github.com/codeownersnet/atlas/internal/client"
)

const (
	// API path
	apiVersion = "/v2"
)

// Client is an Opsgenie API client
type Client struct {
	httpClient *client.Client
	baseURL    string
}

// Config holds the configuration for creating an Opsgenie client
type Config struct {
	BaseURL       string
	Auth          auth.Provider
	CustomHeaders map[string]string
	SSLVerify     bool
	HTTPProxy     string
	HTTPSProxy    string
	SOCKSProxy    string
	NoProxy       string
}

// NewClient creates a new Opsgenie client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if cfg.Auth == nil {
		return nil, fmt.Errorf("auth provider is required")
	}

	// Create HTTP client
	httpClient, err := client.NewClient(&client.Config{
		BaseURL:       cfg.BaseURL,
		Auth:          cfg.Auth,
		CustomHeaders: cfg.CustomHeaders,
		SSLVerify:     cfg.SSLVerify,
		HTTPProxy:     cfg.HTTPProxy,
		HTTPSProxy:    cfg.HTTPSProxy,
		SOCKSProxy:    cfg.SOCKSProxy,
		NoProxy:       cfg.NoProxy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
	}, nil
}

// buildURL constructs a full URL from a path
func (c *Client) buildURL(endpoint string) string {
	// If endpoint already has API version, use as-is
	if strings.HasPrefix(endpoint, apiVersion) {
		return c.baseURL + endpoint
	}
	// Otherwise prepend API version
	return c.baseURL + apiVersion + endpoint
}

// buildURLWithParams constructs a full URL with query parameters
func buildURLWithParams(base string, params map[string]string) string {
	if len(params) == 0 {
		return base
	}

	values := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			values = append(values, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
		}
	}

	if len(values) == 0 {
		return base
	}

	return base + "?" + strings.Join(values, "&")
}

// doRequest performs an HTTP request and decodes the response
func (c *Client) doRequest(ctx context.Context, method, path string, body []byte, result interface{}) error {
	var resp *http.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = c.httpClient.Get(ctx, path)
	case http.MethodPost:
		resp, err = c.httpClient.Post(ctx, path, body)
	case http.MethodPut:
		resp, err = c.httpClient.Put(ctx, path, body)
	case http.MethodDelete:
		resp, err = c.httpClient.Delete(ctx, path)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return c.parseError(resp.StatusCode, respBody)
	}

	// Decode response if result is provided
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// parseError parses an error response from Opsgenie
func (c *Client) parseError(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// If we can't parse the error, return the raw body
		return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
	}

	if errResp.Message != "" {
		return fmt.Errorf("HTTP %d: %s", statusCode, errResp.Message)
	}

	return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
}

// getAPIPath returns the API path
func (c *Client) getAPIPath() string {
	return apiVersion
}

// GetAlert retrieves an alert by ID or alias
func (c *Client) GetAlert(ctx context.Context, id string) (*Alert, error) {
	path := fmt.Sprintf("%s/alerts/%s", apiVersion, id)

	var response struct {
		Data      *Alert  `json:"data"`
		Took      float64 `json:"took,omitempty"`
		RequestID string  `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get alert %s: %w", id, err)
	}

	return response.Data, nil
}

// ListAlerts retrieves a list of alerts based on query parameters
func (c *Client) ListAlerts(ctx context.Context, query string, limit, offset int) (*ListAlertsResponse, error) {
	path := fmt.Sprintf("%s/alerts", apiVersion)

	// Build query parameters
	params := make(map[string]string)
	if query != "" {
		params["query"] = query
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["offset"] = fmt.Sprintf("%d", offset)
	}

	path = buildURLWithParams(path, params)

	var response ListAlertsResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}

	return &response, nil
}

// CountAlerts returns the count of alerts matching the query
func (c *Client) CountAlerts(ctx context.Context, query string) (int, error) {
	path := fmt.Sprintf("%s/alerts/count", apiVersion)

	// Build query parameters
	params := make(map[string]string)
	if query != "" {
		params["query"] = query
	}

	path = buildURLWithParams(path, params)

	var response struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
		Took      float64 `json:"took,omitempty"`
		RequestID string  `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	return response.Data.Count, nil
}

// CreateAlert creates a new alert
func (c *Client) CreateAlert(ctx context.Context, req *AlertRequest) (*CreateAlertResponse, error) {
	path := fmt.Sprintf("%s/alerts", apiVersion)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alert request: %w", err)
	}

	var response CreateAlertResponse
	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return &response, nil
}

// CloseAlert closes an alert by ID or alias
func (c *Client) CloseAlert(ctx context.Context, id, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/close", apiVersion, id)

	request := make(map[string]interface{})
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal close alert request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to close alert %s: %w", id, err)
	}

	return nil
}

// AcknowledgeAlert acknowledges an alert by ID or alias
func (c *Client) AcknowledgeAlert(ctx context.Context, id, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/acknowledge", apiVersion, id)

	request := make(map[string]interface{})
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal acknowledge alert request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to acknowledge alert %s: %w", id, err)
	}

	return nil
}

// SnoozeAlert snoozes an alert by ID or alias until the specified end time
func (c *Client) SnoozeAlert(ctx context.Context, id, endTime, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/snooze", apiVersion, id)

	request := map[string]interface{}{
		"endTime": endTime,
	}
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal snooze alert request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to snooze alert %s: %w", id, err)
	}

	return nil
}

// GetSchedule retrieves a schedule by ID
func (c *Client) GetSchedule(ctx context.Context, id string) (*Schedule, error) {
	path := fmt.Sprintf("%s/schedules/%s", apiVersion, id)

	var response struct {
		Data      *Schedule `json:"data"`
		Took      float64   `json:"took,omitempty"`
		RequestID string    `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get schedule %s: %w", id, err)
	}

	return response.Data, nil
}

// ListSchedules retrieves a list of all schedules
func (c *Client) ListSchedules(ctx context.Context) ([]Schedule, error) {
	path := fmt.Sprintf("%s/schedules", apiVersion)

	var response struct {
		Data      []Schedule `json:"data"`
		Took      float64    `json:"took,omitempty"`
		RequestID string     `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}

	return response.Data, nil
}

// GetScheduleTimeline retrieves the timeline for a schedule within a time range
func (c *Client) GetScheduleTimeline(ctx context.Context, id string, from, to time.Time) (*ScheduleTimeline, error) {
	path := fmt.Sprintf("%s/schedules/%s/timeline", apiVersion, id)

	// Build query parameters
	params := make(map[string]string)
	if !from.IsZero() {
		params["date"] = from.Format(time.RFC3339)
	}
	if !to.IsZero() {
		params["interval"] = fmt.Sprintf("%d", int(to.Sub(from).Hours()/24))
		params["intervalUnit"] = "days"
	}

	path = buildURLWithParams(path, params)

	var response struct {
		Data      *ScheduleTimeline `json:"data"`
		Took      float64           `json:"took,omitempty"`
		RequestID string            `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get schedule timeline for %s: %w", id, err)
	}

	return response.Data, nil
}

// GetOnCalls retrieves the current on-call participants
// If schedule is provided, returns on-calls for that schedule only.
// If schedule is empty, returns on-calls for all schedules.
func (c *Client) GetOnCalls(ctx context.Context, schedule string) ([]OnCall, error) {
	if schedule != "" {
		// Get on-calls for specific schedule
		return c.getScheduleOnCalls(ctx, schedule)
	}

	// Get on-calls for all schedules
	schedules, err := c.ListSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}

	var allOnCalls []OnCall
	for _, sched := range schedules {
		onCalls, err := c.getScheduleOnCalls(ctx, sched.ID)
		if err != nil {
			// Log error but continue with other schedules
			continue
		}
		allOnCalls = append(allOnCalls, onCalls...)
	}

	return allOnCalls, nil
}

// getScheduleOnCalls retrieves on-calls for a specific schedule
func (c *Client) getScheduleOnCalls(ctx context.Context, scheduleID string) ([]OnCall, error) {
	path := fmt.Sprintf("%s/schedules/%s/on-calls", apiVersion, scheduleID)

	var response struct {
		Data      *ScheduleOnCallResponse `json:"data"`
		Took      float64                 `json:"took,omitempty"`
		RequestID string                  `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get on-calls for schedule %s: %w", scheduleID, err)
	}

	if response.Data == nil {
		return []OnCall{}, nil
	}

	// Convert to OnCall format with schedule info
	result := []OnCall{
		{
			ScheduleID:       scheduleID,
			ScheduleName:     response.Data.Parent.Name,
			OnCallRecipients: response.Data.OnCallParticipants,
		},
	}

	return result, nil
}

// GetTeam retrieves a team by ID or name
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	path := fmt.Sprintf("%s/teams/%s", apiVersion, id)

	var response struct {
		Data      *Team   `json:"data"`
		Took      float64 `json:"took,omitempty"`
		RequestID string  `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get team %s: %w", id, err)
	}

	return response.Data, nil
}

// ListTeams retrieves a list of all teams
func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	path := fmt.Sprintf("%s/teams", apiVersion)

	var response struct {
		Data      []Team  `json:"data"`
		Took      float64 `json:"took,omitempty"`
		RequestID string  `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return response.Data, nil
}

// GetUser retrieves a user by identifier (ID, username, or email)
func (c *Client) GetUser(ctx context.Context, identifier string) (*User, error) {
	path := fmt.Sprintf("%s/users/%s", apiVersion, identifier)

	var response struct {
		Data      *User   `json:"data"`
		Took      float64 `json:"took,omitempty"`
		RequestID string  `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", identifier, err)
	}

	return response.Data, nil
}

// GetIncident retrieves an incident by ID
func (c *Client) GetIncident(ctx context.Context, id string) (*Incident, error) {
	path := fmt.Sprintf("%s/incidents/%s", apiVersion, id)

	var response struct {
		Data      *Incident `json:"data"`
		Took      float64   `json:"took,omitempty"`
		RequestID string    `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get incident %s: %w", id, err)
	}

	return response.Data, nil
}

// ListIncidents retrieves a list of incidents based on query parameters
func (c *Client) ListIncidents(ctx context.Context, query string, limit, offset int) (*IncidentResponse, error) {
	path := fmt.Sprintf("%s/incidents", apiVersion)

	// Build query parameters
	params := make(map[string]string)
	if query != "" {
		params["query"] = query
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["offset"] = fmt.Sprintf("%d", offset)
	}

	path = buildURLWithParams(path, params)

	var response IncidentResponse
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	return &response, nil
}

// CreateIncident creates a new incident
func (c *Client) CreateIncident(ctx context.Context, req *IncidentRequest) (*Incident, error) {
	path := fmt.Sprintf("%s/incidents", apiVersion)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal incident request: %w", err)
	}

	var response struct {
		Data      *Incident `json:"data"`
		Result    string    `json:"result"`
		Took      float64   `json:"took"`
		RequestID string    `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return response.Data, nil
}

// CloseIncident closes an incident by ID
func (c *Client) CloseIncident(ctx context.Context, id, note string) error {
	path := fmt.Sprintf("%s/incidents/%s/close", apiVersion, id)

	request := make(map[string]interface{})
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal close incident request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to close incident %s: %w", id, err)
	}

	return nil
}

// AddNoteToIncident adds a note to an incident by ID
func (c *Client) AddNoteToIncident(ctx context.Context, id, note string) error {
	path := fmt.Sprintf("%s/incidents/%s/notes", apiVersion, id)

	request := map[string]interface{}{
		"note": note,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal add note request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to add note to incident %s: %w", id, err)
	}

	return nil
}

// AddResponderToIncident adds a responder to an incident by ID
func (c *Client) AddResponderToIncident(ctx context.Context, id string, responder *Responder) error {
	path := fmt.Sprintf("%s/incidents/%s/responders", apiVersion, id)

	request := map[string]interface{}{
		"responder": responder,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal add responder request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to add responder to incident %s: %w", id, err)
	}

	return nil
}

// EscalateAlert escalates an alert to a specified responder (escalation policy)
func (c *Client) EscalateAlert(ctx context.Context, id string, escalation *Responder, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/escalate", apiVersion, id)

	request := map[string]interface{}{
		"escalation": escalation,
	}
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal escalate alert request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to escalate alert %s: %w", id, err)
	}

	return nil
}

// AssignAlert assigns an alert to a specified owner
func (c *Client) AssignAlert(ctx context.Context, id string, owner *Responder, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/assign", apiVersion, id)

	request := map[string]interface{}{
		"owner": owner,
	}
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal assign alert request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to assign alert %s: %w", id, err)
	}

	return nil
}

// AddNoteToAlert adds a note to an alert by ID or alias
func (c *Client) AddNoteToAlert(ctx context.Context, id, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/notes", apiVersion, id)

	request := map[string]interface{}{
		"note": note,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal add note request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to add note to alert %s: %w", id, err)
	}

	return nil
}

// AddTagsToAlert adds tags to an alert by ID or alias
func (c *Client) AddTagsToAlert(ctx context.Context, id string, tags []string, note string) error {
	path := fmt.Sprintf("%s/alerts/%s/tags", apiVersion, id)

	request := map[string]interface{}{
		"tags": tags,
	}
	if note != "" {
		request["note"] = note
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal add tags request: %w", err)
	}

	var response struct {
		Result    string  `json:"result"`
		Took      float64 `json:"took"`
		RequestID string  `json:"requestId"`
	}

	if err := c.doRequest(ctx, http.MethodPost, path, reqBody, &response); err != nil {
		return fmt.Errorf("failed to add tags to alert %s: %w", id, err)
	}

	return nil
}

// GetRequestStatus retrieves the status of an asynchronous request
func (c *Client) GetRequestStatus(ctx context.Context, requestID string) (*AsyncResponse, error) {
	path := fmt.Sprintf("%s/alerts/requests/%s", apiVersion, requestID)

	var response struct {
		Data      *AsyncResponse `json:"data"`
		Took      float64        `json:"took,omitempty"`
		RequestID string         `json:"requestId,omitempty"`
	}

	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get request status %s: %w", requestID, err)
	}

	return response.Data, nil
}
