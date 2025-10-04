package confluence

import "fmt"

// DeploymentType represents the Confluence deployment type
type DeploymentType string

const (
	DeploymentCloud  DeploymentType = "cloud"
	DeploymentServer DeploymentType = "server"
)

// ContentType represents the type of Confluence content
type ContentType string

const (
	ContentTypePage       ContentType = "page"
	ContentTypeBlogPost   ContentType = "blogpost"
	ContentTypeComment    ContentType = "comment"
	ContentTypeAttachment ContentType = "attachment"
)

// ContentStatus represents the status of content
type ContentStatus string

const (
	ContentStatusCurrent ContentStatus = "current"
	ContentStatusTrashed ContentStatus = "trashed"
	ContentStatusDeleted ContentStatus = "deleted"
	ContentStatusDraft   ContentStatus = "draft"
)

// ContentFormat represents the format of content
type ContentFormat string

const (
	FormatStorage ContentFormat = "storage" // Confluence storage format (XHTML)
	FormatView    ContentFormat = "view"    // HTML view format
	FormatExport  ContentFormat = "export_view"
	FormatEditor  ContentFormat = "editor"
	FormatWiki    ContentFormat = "wiki" // Wiki markup (legacy)
)

// Content represents a piece of Confluence content (page, blogpost, comment, etc.)
type Content struct {
	ID          string                 `json:"id"`
	Type        ContentType            `json:"type"`
	Status      ContentStatus          `json:"status"`
	Title       string                 `json:"title,omitempty"`
	Space       *Space                 `json:"space,omitempty"`
	Version     *Version               `json:"version,omitempty"`
	Ancestors   []Content              `json:"ancestors,omitempty"`
	Body        *Body                  `json:"body,omitempty"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"` // Can contain various types
	Links       *Links                 `json:"_links,omitempty"`
	Expandable  *Expandable            `json:"_expandable,omitempty"`
	History     *History               `json:"history,omitempty"`
	Children    *Children              `json:"children,omitempty"`
	Descendants *Descendants           `json:"descendants,omitempty"`
	Container   *Container             `json:"container,omitempty"`
	Metadata    *Metadata              `json:"metadata,omitempty"`
}

// Space represents a Confluence space
type Space struct {
	ID          interface{} `json:"id,omitempty"` // Can be string or number depending on API version
	Key         string      `json:"key"`
	Name        string      `json:"name,omitempty"`
	Type        string      `json:"type,omitempty"`
	Status      string      `json:"status,omitempty"`
	Description *Body       `json:"description,omitempty"`
	Homepage    *Content    `json:"homepage,omitempty"`
	Icon        *Icon       `json:"icon,omitempty"`
	Links       *Links      `json:"_links,omitempty"`
	Expandable  *Expandable `json:"_expandable,omitempty"`
}

// GetID returns the space ID as a string
func (s *Space) GetID() string {
	if s.ID == nil {
		return ""
	}
	switch v := s.ID.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Version represents content version information
type Version struct {
	By                  *User  `json:"by,omitempty"`
	When                string `json:"when,omitempty"`
	FriendlyWhen        string `json:"friendlyWhen,omitempty"`
	Message             string `json:"message,omitempty"`
	Number              int    `json:"number"`
	MinorEdit           bool   `json:"minorEdit,omitempty"`
	SyncRev             string `json:"syncRev,omitempty"`
	SyncRevSource       string `json:"syncRevSource,omitempty"`
	ConfRev             string `json:"confRev,omitempty"`
	ContentTypeModified bool   `json:"contentTypeModified,omitempty"`
}

// Body represents content body in various formats
type Body struct {
	Storage             *BodyContent `json:"storage,omitempty"`
	View                *BodyContent `json:"view,omitempty"`
	ExportView          *BodyContent `json:"export_view,omitempty"`
	StyledView          *BodyContent `json:"styled_view,omitempty"`
	Editor              *BodyContent `json:"editor,omitempty"`
	Editor2             *BodyContent `json:"editor2,omitempty"`
	AnonymousExportView *BodyContent `json:"anonymous_export_view,omitempty"`
	Wiki                *BodyContent `json:"wiki,omitempty"`
}

// BodyContent represents the actual content in a specific format
type BodyContent struct {
	Value          string        `json:"value"`
	Representation ContentFormat `json:"representation"`
	Embeddeds      []Embedded    `json:"embeddedContent,omitempty"`
	WebResource    *WebResource  `json:"webresource,omitempty"`
}

// Embedded represents embedded content
type Embedded struct {
	EntityID   string `json:"entityId,omitempty"`
	EntityType string `json:"entityType,omitempty"`
}

// WebResource represents web resources
type WebResource struct {
	Keys       []string          `json:"keys,omitempty"`
	Contexts   []string          `json:"contexts,omitempty"`
	Uris       map[string]string `json:"uris,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
	SuperBatch *SuperBatch       `json:"superbatch,omitempty"`
}

// SuperBatch represents super batch information
type SuperBatch struct {
	Uris map[string]string `json:"uris,omitempty"`
}

// User represents a Confluence user
type User struct {
	Type           string          `json:"type,omitempty"`
	AccountID      string          `json:"accountId,omitempty"` // Cloud
	Username       string          `json:"username,omitempty"`  // Server/DC
	UserKey        string          `json:"userKey,omitempty"`   // Server/DC
	AccountType    string          `json:"accountType,omitempty"`
	Email          string          `json:"email,omitempty"`
	PublicName     string          `json:"publicName,omitempty"`
	DisplayName    string          `json:"displayName,omitempty"`
	ProfilePicture *ProfilePicture `json:"profilePicture,omitempty"`
	Links          *Links          `json:"_links,omitempty"`
	Expandable     *Expandable     `json:"_expandable,omitempty"`
}

// ProfilePicture represents a user's profile picture
type ProfilePicture struct {
	Path      string `json:"path,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// History represents content history
type History struct {
	Latest          bool          `json:"latest,omitempty"`
	CreatedBy       *User         `json:"createdBy,omitempty"`
	CreatedDate     string        `json:"createdDate,omitempty"`
	LastUpdated     *Version      `json:"lastUpdated,omitempty"`
	PreviousVersion *Version      `json:"previousVersion,omitempty"`
	Contributors    *Contributors `json:"contributors,omitempty"`
	NextVersion     *Version      `json:"nextVersion,omitempty"`
	Links           *Links        `json:"_links,omitempty"`
	Expandable      *Expandable   `json:"_expandable,omitempty"`
}

// Contributors represents content contributors
type Contributors struct {
	Publishers *UserArray `json:"publishers,omitempty"`
}

// UserArray represents an array of users
type UserArray struct {
	Users      []User      `json:"users,omitempty"`
	UserKeys   []string    `json:"userKeys,omitempty"`
	Size       int         `json:"size,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expandable *Expandable `json:"_expandable,omitempty"`
}

// Label represents a label
type Label struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Prefix string `json:"prefix,omitempty"`
	Label  string `json:"label,omitempty"`
}

// Comment represents a comment
type Comment struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Status     string      `json:"status,omitempty"`
	Title      string      `json:"title,omitempty"`
	Body       *Body       `json:"body,omitempty"`
	Version    *Version    `json:"version,omitempty"`
	Container  *Container  `json:"container,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expandable *Expandable `json:"_expandable,omitempty"`
}

// Children represents child content
type Children struct {
	Page       *ContentArray `json:"page,omitempty"`
	Comment    *ContentArray `json:"comment,omitempty"`
	Attachment *ContentArray `json:"attachment,omitempty"`
	Links      *Links        `json:"_links,omitempty"`
	Expandable *Expandable   `json:"_expandable,omitempty"`
}

// Descendants represents descendant content
type Descendants struct {
	Page       *ContentArray `json:"page,omitempty"`
	Comment    *ContentArray `json:"comment,omitempty"`
	Attachment *ContentArray `json:"attachment,omitempty"`
	Links      *Links        `json:"_links,omitempty"`
	Expandable *Expandable   `json:"_expandable,omitempty"`
}

// ContentArray represents an array of content
type ContentArray struct {
	Results    []Content   `json:"results,omitempty"`
	Start      int         `json:"start,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Size       int         `json:"size,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expandable *Expandable `json:"_expandable,omitempty"`
}

// Container represents a content container
type Container struct {
	ID         string      `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Title      string      `json:"title,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expandable *Expandable `json:"_expandable,omitempty"`
}

// Metadata represents content metadata
type Metadata struct {
	Labels     *LabelArray            `json:"labels,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Frontend   *Frontend              `json:"frontend,omitempty"`
	Links      *Links                 `json:"_links,omitempty"`
	Expandable *Expandable            `json:"_expandable,omitempty"`
}

// LabelArray represents an array of labels
type LabelArray struct {
	Results    []Label     `json:"results,omitempty"`
	Start      int         `json:"start,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Size       int         `json:"size,omitempty"`
	Links      *Links      `json:"_links,omitempty"`
	Expandable *Expandable `json:"_expandable,omitempty"`
}

// Frontend represents frontend metadata
type Frontend struct {
	EditURL string `json:"editUrl,omitempty"`
	WebUI   string `json:"webui,omitempty"`
}

// Icon represents an icon
type Icon struct {
	Path      string `json:"path,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// Links represents HAL links
type Links struct {
	Self       string `json:"self,omitempty"`
	Base       string `json:"base,omitempty"`
	Context    string `json:"context,omitempty"`
	WebUI      string `json:"webui,omitempty"`
	Edit       string `json:"edit,omitempty"`
	TinyUI     string `json:"tinyui,omitempty"`
	Collection string `json:"collection,omitempty"`
	Download   string `json:"download,omitempty"`
}

// Expandable represents expandable fields
type Expandable struct {
	Container    string `json:"container,omitempty"`
	Metadata     string `json:"metadata,omitempty"`
	Operations   string `json:"operations,omitempty"`
	Children     string `json:"children,omitempty"`
	Restrictions string `json:"restrictions,omitempty"`
	History      string `json:"history,omitempty"`
	Ancestors    string `json:"ancestors,omitempty"`
	Body         string `json:"body,omitempty"`
	Version      string `json:"version,omitempty"`
	Descendants  string `json:"descendants,omitempty"`
	Space        string `json:"space,omitempty"`
}

// SearchResult represents search results
type SearchResult struct {
	Results        []Content `json:"results"`
	Start          int       `json:"start"`
	Limit          int       `json:"limit"`
	Size           int       `json:"size"`
	TotalSize      int       `json:"totalSize,omitempty"`
	CqlQuery       string    `json:"cqlQuery,omitempty"`
	SearchDuration int       `json:"searchDuration,omitempty"`
	Links          *Links    `json:"_links,omitempty"`
}

// CreateContentRequest represents a request to create content
type CreateContentRequest struct {
	Type      ContentType   `json:"type"`
	Title     string        `json:"title"`
	Space     *SpaceRef     `json:"space"`
	Body      *Body         `json:"body,omitempty"`
	Ancestors []ContentRef  `json:"ancestors,omitempty"`
	Status    ContentStatus `json:"status,omitempty"`
}

// UpdateContentRequest represents a request to update content
type UpdateContentRequest struct {
	Version *Version      `json:"version"`
	Title   string        `json:"title,omitempty"`
	Type    ContentType   `json:"type,omitempty"`
	Body    *Body         `json:"body,omitempty"`
	Status  ContentStatus `json:"status,omitempty"`
}

// SpaceRef represents a space reference
type SpaceRef struct {
	Key string `json:"key"`
}

// ContentRef represents a content reference
type ContentRef struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// CreateLabelRequest represents a request to create a label
type CreateLabelRequest struct {
	Prefix string `json:"prefix,omitempty"`
	Name   string `json:"name"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Type      string       `json:"type"`
	Container *ContentRef  `json:"container"`
	Body      *Body        `json:"body"`
	Ancestors []ContentRef `json:"ancestors,omitempty"`
}

// ErrorResponse represents a Confluence error response
type ErrorResponse struct {
	StatusCode int        `json:"statusCode,omitempty"`
	Data       *ErrorData `json:"data,omitempty"`
	Message    string     `json:"message,omitempty"`
	Reason     string     `json:"reason,omitempty"`
}

// ErrorData represents error data
type ErrorData struct {
	Authorized bool              `json:"authorized,omitempty"`
	Valid      bool              `json:"valid,omitempty"`
	Errors     []ValidationError `json:"errors,omitempty"`
	Successful bool              `json:"successful,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Message *ErrorMessage `json:"message,omitempty"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Key  string        `json:"key,omitempty"`
	Args []interface{} `json:"args,omitempty"`
}
