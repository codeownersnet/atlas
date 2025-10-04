package jira

import (
	"encoding/json"
	"fmt"
	"time"
)

// DeploymentType represents the Jira deployment type
type DeploymentType string

const (
	DeploymentCloud  DeploymentType = "cloud"
	DeploymentServer DeploymentType = "server"
)

// AtlassianTime is a custom time type that handles multiple timestamp formats from Atlassian APIs
type AtlassianTime struct {
	time.Time
}

// Atlassian timestamp formats
var atlassianTimeFormats = []string{
	time.RFC3339,                   // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05.000-0700", // Atlassian format with milliseconds and +HHMM timezone
	"2006-01-02T15:04:05-0700",     // Atlassian format without milliseconds
	time.RFC3339Nano,               // "2006-01-02T15:04:05.999999999Z07:00"
}

// UnmarshalJSON implements json.Unmarshaler interface
func (at *AtlassianTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		at.Time = time.Time{}
		return nil
	}

	// Try each format until one works
	var lastErr error
	for _, format := range atlassianTimeFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			at.Time = t
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("unable to parse time %q: %w", s, lastErr)
}

// MarshalJSON implements json.Marshaler interface
func (at AtlassianTime) MarshalJSON() ([]byte, error) {
	if at.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(at.Time.Format(time.RFC3339Nano))
}

// String returns the string representation of the time
func (at AtlassianTime) String() string {
	if at.Time.IsZero() {
		return ""
	}
	return at.Time.Format(time.RFC3339Nano)
}

// IsZero returns true if the time is the zero value
func (at AtlassianTime) IsZero() bool {
	return at.Time.IsZero()
}

// Description represents a Jira description field that can be either plain text or ADF format.
//
// Jira returns descriptions in different formats depending on the instance configuration:
//   - Plain text: Older instances or when text-only format is configured
//   - Atlassian Document Format (ADF): Modern Cloud instances and newer Server/DC versions
//
// This type automatically detects the format during JSON unmarshaling and provides
// convenient methods to access the content:
//
//   - String() returns the plain text representation (extracted from ADF if needed)
//   - IsADF() returns true if the description is in ADF format
//   - Raw() returns the raw JSON for advanced use cases
//
// Example usage:
//
//	issue, err := client.GetIssue(ctx, "PROJ-123", nil)
//	if err != nil {
//	    return err
//	}
//	if issue.Fields.Description != nil {
//	    fmt.Println("Description:", issue.Fields.Description.String())
//	    if issue.Fields.Description.IsADF() {
//	        fmt.Println("(ADF format)")
//	    }
//	}
//
// When creating or updating issues, pass description as a plain string in the fields map:
//
//	fields := map[string]interface{}{
//	    "description": "This is a plain text description",
//	    // ... other fields
//	}
type Description struct {
	raw   json.RawMessage
	isADF bool
	text  string
}

// UnmarshalJSON implements json.Unmarshaler interface to handle both plain text and ADF format
func (d *Description) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as an object first (ADF format)
	var adfObj map[string]interface{}
	if err := json.Unmarshal(data, &adfObj); err == nil {
		// Check if it looks like an ADF object (has "type" and "version" fields)
		if _, hasType := adfObj["type"]; hasType {
			if _, hasVersion := adfObj["version"]; hasVersion {
				d.raw = data
				d.isADF = true
				d.text = extractTextFromADF(adfObj)
				return nil
			}
		}
	}

	// Fall back to plain text
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("description must be string or ADF object: %w", err)
	}
	d.raw = data
	d.isADF = false
	d.text = s
	return nil
}

// MarshalJSON implements json.Marshaler interface
func (d Description) MarshalJSON() ([]byte, error) {
	if d.raw != nil {
		return d.raw, nil
	}
	return json.Marshal(d.text)
}

// String returns the plain text representation of the description
func (d *Description) String() string {
	if d == nil {
		return ""
	}
	return d.text
}

// IsADF returns true if the description is in ADF format
func (d *Description) IsADF() bool {
	if d == nil {
		return false
	}
	return d.isADF
}

// Raw returns the raw JSON representation
func (d *Description) Raw() json.RawMessage {
	if d == nil {
		return nil
	}
	return d.raw
}

// extractTextFromADF recursively extracts text content from an ADF object
func extractTextFromADF(obj map[string]interface{}) string {
	var text string

	// Check if this node has text content
	if textVal, ok := obj["text"].(string); ok {
		text += textVal
	}

	// Recursively process content array
	if content, ok := obj["content"].([]interface{}); ok {
		for _, item := range content {
			if itemMap, ok := item.(map[string]interface{}); ok {
				extracted := extractTextFromADF(itemMap)
				if extracted != "" {
					if text != "" {
						// Add space or newline between content blocks
						nodeType, _ := itemMap["type"].(string)
						if nodeType == "paragraph" || nodeType == "heading" {
							text += "\n"
						} else if text != "" && extracted != "" {
							text += " "
						}
					}
					text += extracted
				}
			}
		}
	}

	return text
}

// Issue represents a Jira issue
type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
	Expand string      `json:"expand,omitempty"`
}

// IssueFields represents all possible fields in a Jira issue
type IssueFields struct {
	Summary     string        `json:"summary,omitempty"`
	Description *Description  `json:"description,omitempty"`
	IssueType   *IssueType    `json:"issuetype,omitempty"`
	Project     *Project      `json:"project,omitempty"`
	Reporter    *User         `json:"reporter,omitempty"`
	Assignee    *User         `json:"assignee,omitempty"`
	Priority    *Priority     `json:"priority,omitempty"`
	Status      *Status       `json:"status,omitempty"`
	Resolution  *Resolution   `json:"resolution,omitempty"`
	Labels      []string      `json:"labels,omitempty"`
	Components  []Component   `json:"components,omitempty"`
	FixVersions []Version     `json:"fixVersions,omitempty"`
	Versions    []Version     `json:"versions,omitempty"`
	Created     AtlassianTime `json:"created,omitempty"`
	Updated     AtlassianTime `json:"updated,omitempty"`
	DueDate     *string       `json:"duedate,omitempty"`
	Parent      *IssueParent  `json:"parent,omitempty"`
	Subtasks    []Issue       `json:"subtasks,omitempty"`
	IssueLinks  []IssueLink   `json:"issuelinks,omitempty"`
	Attachment  []Attachment  `json:"attachment,omitempty"`
	Comment     *Comments     `json:"comment,omitempty"`
	Worklog     *Worklogs     `json:"worklog,omitempty"`

	// Custom fields stored as raw JSON
	Unknowns map[string]interface{} `json:"-"`
}

// IssueType represents a Jira issue type
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Subtask     bool   `json:"subtask,omitempty"`
}

// Project represents a Jira project
type Project struct {
	ID              string           `json:"id"`
	Key             string           `json:"key"`
	Name            string           `json:"name"`
	Self            string           `json:"self,omitempty"`
	ProjectTypeKey  string           `json:"projectTypeKey,omitempty"`
	AvatarUrls      *AvatarUrls      `json:"avatarUrls,omitempty"`
	Lead            *User            `json:"lead,omitempty"`
	Description     string           `json:"description,omitempty"`
	URL             string           `json:"url,omitempty"`
	IssueTypes      []IssueType      `json:"issueTypes,omitempty"`
	ProjectCategory *ProjectCategory `json:"projectCategory,omitempty"`
	Versions        []Version        `json:"versions,omitempty"`
	Components      []Component      `json:"components,omitempty"`
}

// ProjectCategory represents a project category
type ProjectCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

// User represents a Jira user
type User struct {
	Self         string      `json:"self,omitempty"`
	AccountID    string      `json:"accountId,omitempty"` // Cloud
	Name         string      `json:"name,omitempty"`      // Server/DC
	Key          string      `json:"key,omitempty"`       // Server/DC
	EmailAddress string      `json:"emailAddress,omitempty"`
	DisplayName  string      `json:"displayName,omitempty"`
	Active       bool        `json:"active,omitempty"`
	TimeZone     string      `json:"timeZone,omitempty"`
	AvatarUrls   *AvatarUrls `json:"avatarUrls,omitempty"`
}

// AvatarUrls represents avatar URLs
type AvatarUrls struct {
	Size48 string `json:"48x48,omitempty"`
	Size24 string `json:"24x24,omitempty"`
	Size16 string `json:"16x16,omitempty"`
	Size32 string `json:"32x32,omitempty"`
}

// Priority represents issue priority
type Priority struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Self    string `json:"self,omitempty"`
	IconURL string `json:"iconUrl,omitempty"`
}

// Status represents issue status
type Status struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Self           string          `json:"self,omitempty"`
	Description    string          `json:"description,omitempty"`
	IconURL        string          `json:"iconUrl,omitempty"`
	StatusCategory *StatusCategory `json:"statusCategory,omitempty"`
}

// StatusCategory represents a status category
type StatusCategory struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	Self      string `json:"self,omitempty"`
	ColorName string `json:"colorName,omitempty"`
}

// Resolution represents issue resolution
type Resolution struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

// Component represents a project component
type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
	Lead        *User  `json:"lead,omitempty"`
}

// Version represents a project version
type Version struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Self        string         `json:"self,omitempty"`
	Archived    bool           `json:"archived,omitempty"`
	Released    bool           `json:"released,omitempty"`
	ReleaseDate *AtlassianTime `json:"releaseDate,omitempty"`
	ProjectID   int            `json:"projectId,omitempty"`
}

// IssueParent represents the parent of a subtask
type IssueParent struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self,omitempty"`
	Fields IssueFields `json:"fields,omitempty"`
}

// IssueLink represents a link between issues
type IssueLink struct {
	ID           string        `json:"id"`
	Type         IssueLinkType `json:"type"`
	InwardIssue  *LinkedIssue  `json:"inwardIssue,omitempty"`
	OutwardIssue *LinkedIssue  `json:"outwardIssue,omitempty"`
	Self         string        `json:"self,omitempty"`
}

// IssueLinkType represents the type of link between issues
type IssueLinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
	Self    string `json:"self,omitempty"`
}

// LinkedIssue represents a linked issue
type LinkedIssue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self,omitempty"`
	Fields IssueFields `json:"fields,omitempty"`
}

// Attachment represents an issue attachment
type Attachment struct {
	ID        string        `json:"id"`
	Self      string        `json:"self,omitempty"`
	Filename  string        `json:"filename"`
	Author    *User         `json:"author,omitempty"`
	Created   AtlassianTime `json:"created,omitempty"`
	Size      int64         `json:"size,omitempty"`
	MimeType  string        `json:"mimeType,omitempty"`
	Content   string        `json:"content,omitempty"`
	Thumbnail string        `json:"thumbnail,omitempty"`
}

// Comment represents an issue comment
type Comment struct {
	ID           string        `json:"id"`
	Self         string        `json:"self,omitempty"`
	Author       *User         `json:"author,omitempty"`
	Body         string        `json:"body"`
	UpdateAuthor *User         `json:"updateAuthor,omitempty"`
	Created      AtlassianTime `json:"created,omitempty"`
	Updated      AtlassianTime `json:"updated,omitempty"`
	Visibility   *Visibility   `json:"visibility,omitempty"`
}

// Comments represents a list of comments
type Comments struct {
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	Comments   []Comment `json:"comments"`
}

// Visibility represents comment/worklog visibility
type Visibility struct {
	Type  string `json:"type"`  // "group" or "role"
	Value string `json:"value"` // group name or role name
}

// Worklog represents a worklog entry
type Worklog struct {
	ID               string        `json:"id"`
	Self             string        `json:"self,omitempty"`
	Author           *User         `json:"author,omitempty"`
	UpdateAuthor     *User         `json:"updateAuthor,omitempty"`
	Comment          string        `json:"comment,omitempty"`
	Created          AtlassianTime `json:"created,omitempty"`
	Updated          AtlassianTime `json:"updated,omitempty"`
	Started          AtlassianTime `json:"started"`
	TimeSpent        string        `json:"timeSpent"`
	TimeSpentSeconds int           `json:"timeSpentSeconds"`
	Visibility       *Visibility   `json:"visibility,omitempty"`
}

// Worklogs represents a list of worklogs
type Worklogs struct {
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	Worklogs   []Worklog `json:"worklogs"`
}

// Field represents a Jira field (standard or custom)
type Field struct {
	ID          string       `json:"id"`
	Key         string       `json:"key,omitempty"`
	Name        string       `json:"name"`
	Custom      bool         `json:"custom"`
	Orderable   bool         `json:"orderable,omitempty"`
	Navigable   bool         `json:"navigable,omitempty"`
	Searchable  bool         `json:"searchable,omitempty"`
	ClauseNames []string     `json:"clauseNames,omitempty"`
	Schema      *FieldSchema `json:"schema,omitempty"`
}

// FieldSchema represents the schema of a field
type FieldSchema struct {
	Type     string `json:"type"`
	Items    string `json:"items,omitempty"`
	System   string `json:"system,omitempty"`
	Custom   string `json:"custom,omitempty"`
	CustomID int    `json:"customId,omitempty"`
}

// Transition represents a status transition
type Transition struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	To     Status               `json:"to"`
	Fields map[string]FieldMeta `json:"fields,omitempty"`
}

// FieldMeta represents metadata about a field in a transition
type FieldMeta struct {
	Required bool   `json:"required"`
	Schema   Schema `json:"schema,omitempty"`
	Name     string `json:"name,omitempty"`
}

// Schema represents a field schema
type Schema struct {
	Type   string `json:"type"`
	System string `json:"system,omitempty"`
}

// SearchResult represents the result of a JQL search
type SearchResult struct {
	Expand     string  `json:"expand,omitempty"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// Board represents an agile board
type Board struct {
	ID       int       `json:"id"`
	Self     string    `json:"self,omitempty"`
	Name     string    `json:"name"`
	Type     string    `json:"type"` // scrum, kanban, simple
	Location *Location `json:"location,omitempty"`
}

// Location represents a board location
type Location struct {
	ProjectID   int    `json:"projectId,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
	ProjectKey  string `json:"projectKey,omitempty"`
	ProjectType string `json:"projectTypeKey,omitempty"`
	AvatarURI   string `json:"avatarURI,omitempty"`
	Name        string `json:"name,omitempty"`
}

// Sprint represents a sprint
type Sprint struct {
	ID            int            `json:"id"`
	Self          string         `json:"self,omitempty"`
	State         string         `json:"state"` // future, active, closed
	Name          string         `json:"name"`
	StartDate     *AtlassianTime `json:"startDate,omitempty"`
	EndDate       *AtlassianTime `json:"endDate,omitempty"`
	CompleteDate  *AtlassianTime `json:"completeDate,omitempty"`
	OriginBoardID int            `json:"originBoardId,omitempty"`
	Goal          string         `json:"goal,omitempty"`
}

// RemoteLink represents a remote issue link
type RemoteLink struct {
	ID           string           `json:"id,omitempty"`
	Self         string           `json:"self,omitempty"`
	GlobalID     string           `json:"globalId,omitempty"`
	Application  *LinkApplication `json:"application,omitempty"`
	Relationship string           `json:"relationship,omitempty"`
	Object       *LinkObject      `json:"object,omitempty"`
}

// LinkApplication represents the application in a remote link
type LinkApplication struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// LinkObject represents the object in a remote link
type LinkObject struct {
	URL     string      `json:"url"`
	Title   string      `json:"title"`
	Summary string      `json:"summary,omitempty"`
	Icon    *LinkIcon   `json:"icon,omitempty"`
	Status  *LinkStatus `json:"status,omitempty"`
}

// LinkIcon represents an icon in a remote link
type LinkIcon struct {
	URL16x16 string `json:"url16x16,omitempty"`
	Title    string `json:"title,omitempty"`
}

// LinkStatus represents status in a remote link
type LinkStatus struct {
	Resolved bool      `json:"resolved,omitempty"`
	Icon     *LinkIcon `json:"icon,omitempty"`
}

// Changelog represents issue changelog
type Changelog struct {
	ID      string          `json:"id"`
	Author  *User           `json:"author,omitempty"`
	Created AtlassianTime   `json:"created"`
	Items   []ChangelogItem `json:"items"`
}

// ChangelogItem represents a single changelog item
type ChangelogItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype"`
	From       string `json:"from,omitempty"`
	FromString string `json:"fromString,omitempty"`
	To         string `json:"to,omitempty"`
	ToString   string `json:"toString,omitempty"`
}

// CreateIssueRequest represents a request to create an issue
type CreateIssueRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

// UpdateIssueRequest represents a request to update an issue
type UpdateIssueRequest struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
	Update map[string]interface{} `json:"update,omitempty"`
}

// TransitionRequest represents a request to transition an issue
type TransitionRequest struct {
	Transition Transition             `json:"transition"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// CreateCommentRequest represents a request to add a comment
type CreateCommentRequest struct {
	Body       string      `json:"body"`
	Visibility *Visibility `json:"visibility,omitempty"`
}

// CreateWorklogRequest represents a request to add a worklog
type CreateWorklogRequest struct {
	Comment          string      `json:"comment,omitempty"`
	Started          string      `json:"started"`
	TimeSpentSeconds int         `json:"timeSpentSeconds"`
	Visibility       *Visibility `json:"visibility,omitempty"`
}

// CreateIssueLinkRequest represents a request to create an issue link
type CreateIssueLinkRequest struct {
	Type         IssueLinkType `json:"type"`
	InwardIssue  LinkIssueRef  `json:"inwardIssue"`
	OutwardIssue LinkIssueRef  `json:"outwardIssue"`
	Comment      *Comment      `json:"comment,omitempty"`
}

// LinkIssueRef represents an issue reference in a link request
type LinkIssueRef struct {
	Key string `json:"key,omitempty"`
	ID  string `json:"id,omitempty"`
}

// CreateVersionRequest represents a request to create a version
type CreateVersionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProjectID   int    `json:"projectId,omitempty"`
	Project     string `json:"project,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	Released    bool   `json:"released,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
}

// CreateSprintRequest represents a request to create a sprint
type CreateSprintRequest struct {
	Name          string `json:"name"`
	StartDate     string `json:"startDate,omitempty"`
	EndDate       string `json:"endDate,omitempty"`
	OriginBoardID int    `json:"originBoardId"`
	Goal          string `json:"goal,omitempty"`
}

// UpdateSprintRequest represents a request to update a sprint
type UpdateSprintRequest struct {
	Name         string `json:"name,omitempty"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	State        string `json:"state,omitempty"`
	Goal         string `json:"goal,omitempty"`
	CompleteDate string `json:"completeDate,omitempty"`
}

// ErrorResponse represents a Jira error response
type ErrorResponse struct {
	ErrorMessages []string          `json:"errorMessages,omitempty"`
	Errors        map[string]string `json:"errors,omitempty"`
}
