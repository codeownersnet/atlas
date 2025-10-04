package opsgenie

import "time"

// DeploymentType represents the Opsgenie deployment type
type DeploymentType string

const (
	DeploymentCloud DeploymentType = "cloud"
)

// Priority represents alert/incident priority
type Priority string

const (
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
	PriorityP3 Priority = "P3"
	PriorityP4 Priority = "P4"
	PriorityP5 Priority = "P5"
)

// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	AlertStatusOpen   AlertStatus = "open"
	AlertStatusClosed AlertStatus = "closed"
)

// IncidentStatus represents the current status of an incident
type IncidentStatus string

const (
	IncidentStatusOpen     IncidentStatus = "open"
	IncidentStatusResolved IncidentStatus = "resolved"
	IncidentStatusClosed   IncidentStatus = "closed"
)

// ResponderType represents the type of responder
type ResponderType string

const (
	ResponderTypeUser       ResponderType = "user"
	ResponderTypeTeam       ResponderType = "team"
	ResponderTypeEscalation ResponderType = "escalation"
	ResponderTypeSchedule   ResponderType = "schedule"
)

// Pagination represents pagination information
type Pagination struct {
	First string `json:"first,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
	Prev  string `json:"prev,omitempty"`
}

// Responder represents a responder (user, team, escalation, or schedule)
type Responder struct {
	Type ResponderType `json:"type"`
	ID   string        `json:"id,omitempty"`
	Name string        `json:"name,omitempty"`
}

// Alert represents an Opsgenie alert
type Alert struct {
	ID             string            `json:"id"`
	TinyID         string            `json:"tinyId,omitempty"`
	Alias          string            `json:"alias,omitempty"`
	Message        string            `json:"message"`
	Status         AlertStatus       `json:"status,omitempty"`
	Acknowledged   bool              `json:"acknowledged,omitempty"`
	IsSeen         bool              `json:"isSeen,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Snoozed        bool              `json:"snoozed,omitempty"`
	SnoozedUntil   *time.Time        `json:"snoozedUntil,omitempty"`
	Count          int               `json:"count,omitempty"`
	LastOccurredAt *time.Time        `json:"lastOccurredAt,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      *time.Time        `json:"updatedAt,omitempty"`
	Source         string            `json:"source,omitempty"`
	Owner          string            `json:"owner,omitempty"`
	Priority       Priority          `json:"priority,omitempty"`
	Responders     []Responder       `json:"responders,omitempty"`
	Integration    *Integration      `json:"integration,omitempty"`
	Report         *Report           `json:"report,omitempty"`
	Actions        []string          `json:"actions,omitempty"`
	Entity         string            `json:"entity,omitempty"`
	Description    string            `json:"description,omitempty"`
	Details        map[string]string `json:"details,omitempty"`
}

// Integration represents integration information
type Integration struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// Report represents alert report information
type Report struct {
	AckTime        int64  `json:"ackTime,omitempty"`
	CloseTime      int64  `json:"closeTime,omitempty"`
	AcknowledgedBy string `json:"acknowledgedBy,omitempty"`
	ClosedBy       string `json:"closedBy,omitempty"`
}

// AlertRequest represents a request to create or update an alert
type AlertRequest struct {
	Message     string            `json:"message"`
	Alias       string            `json:"alias,omitempty"`
	Description string            `json:"description,omitempty"`
	Responders  []Responder       `json:"responders,omitempty"`
	VisibleTo   []Responder       `json:"visibleTo,omitempty"`
	Actions     []string          `json:"actions,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	Entity      string            `json:"entity,omitempty"`
	Source      string            `json:"source,omitempty"`
	Priority    Priority          `json:"priority,omitempty"`
	User        string            `json:"user,omitempty"`
	Note        string            `json:"note,omitempty"`
}

// CreateAlertResponse represents the response when creating an alert
type CreateAlertResponse struct {
	Result    string  `json:"result"`
	Took      float64 `json:"took"`
	RequestID string  `json:"requestId"`
}

// ListAlertsResponse represents the response when listing alerts
type ListAlertsResponse struct {
	Data      []Alert     `json:"data"`
	Paging    *Pagination `json:"paging,omitempty"`
	Took      float64     `json:"took,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// Incident represents an Opsgenie incident
type Incident struct {
	ID               string                 `json:"id"`
	TinyID           string                 `json:"tinyId,omitempty"`
	Message          string                 `json:"message"`
	Status           IncidentStatus         `json:"status"`
	Tags             []string               `json:"tags,omitempty"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        *time.Time             `json:"updatedAt,omitempty"`
	Priority         Priority               `json:"priority"`
	OwnerTeam        string                 `json:"ownerTeam,omitempty"`
	Responders       []Responder            `json:"responders,omitempty"`
	ExtraProperties  map[string]interface{} `json:"extraProperties,omitempty"`
	Actions          []string               `json:"actions,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Details          map[string]string      `json:"details,omitempty"`
	ImpactedServices []string               `json:"impactedServices,omitempty"`
}

// IncidentRequest represents a request to create or update an incident
type IncidentRequest struct {
	Message            string            `json:"message"`
	Description        string            `json:"description,omitempty"`
	Responders         []Responder       `json:"responders,omitempty"`
	Tags               []string          `json:"tags,omitempty"`
	Details            map[string]string `json:"details,omitempty"`
	Priority           Priority          `json:"priority"`
	Note               string            `json:"note,omitempty"`
	ServiceID          string            `json:"serviceId,omitempty"`
	StatusPageEntity   *StatusPageEntity `json:"statusPageEntity,omitempty"`
	NotifyStakeholders bool              `json:"notifyStakeholders,omitempty"`
	ImpactedServices   []string          `json:"impactedServices,omitempty"`
}

// StatusPageEntity represents status page entity information
type StatusPageEntity struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// IncidentResponse represents a generic incident response
type IncidentResponse struct {
	Data      *Incident `json:"data,omitempty"`
	Result    string    `json:"result,omitempty"`
	Took      float64   `json:"took,omitempty"`
	RequestID string    `json:"requestId,omitempty"`
}

// Schedule represents an Opsgenie schedule
type Schedule struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Timezone    string     `json:"timezone,omitempty"`
	Enabled     bool       `json:"enabled,omitempty"`
	OwnerTeam   *Team      `json:"ownerTeam,omitempty"`
	Rotations   []Rotation `json:"rotations,omitempty"`
}

// Rotation represents a schedule rotation
type Rotation struct {
	ID              string           `json:"id,omitempty"`
	Name            string           `json:"name,omitempty"`
	StartDate       *time.Time       `json:"startDate,omitempty"`
	EndDate         *time.Time       `json:"endDate,omitempty"`
	Type            string           `json:"type,omitempty"`
	Length          int              `json:"length,omitempty"`
	Participants    []Responder      `json:"participants,omitempty"`
	TimeRestriction *TimeRestriction `json:"timeRestriction,omitempty"`
}

// TimeRestriction represents time-based restrictions
type TimeRestriction struct {
	Type         string        `json:"type,omitempty"`
	Restrictions []Restriction `json:"restrictions,omitempty"`
}

// Restriction represents a specific time restriction
type Restriction struct {
	StartDay  string `json:"startDay,omitempty"`
	StartHour int    `json:"startHour,omitempty"`
	StartMin  int    `json:"startMin,omitempty"`
	EndDay    string `json:"endDay,omitempty"`
	EndHour   int    `json:"endHour,omitempty"`
	EndMin    int    `json:"endMin,omitempty"`
}

// OnCall represents on-call information
type OnCall struct {
	ScheduleID       string   `json:"scheduleId,omitempty"`
	ScheduleName     string   `json:"scheduleName,omitempty"`
	OnCallRecipients []string `json:"onCallRecipients,omitempty"`
}

// ScheduleOnCallResponse represents the response from per-schedule on-call endpoint
type ScheduleOnCallResponse struct {
	Parent struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	} `json:"_parent"`
	OnCallParticipants []string `json:"onCallParticipants"`
}

// ScheduleTimeline represents schedule timeline information
type ScheduleTimeline struct {
	StartDate     time.Time      `json:"startDate"`
	EndDate       time.Time      `json:"endDate"`
	FinalTimeline *FinalTimeline `json:"finalTimeline,omitempty"`
	BaseTimeline  *BaseTimeline  `json:"baseTimeline,omitempty"`
	Overrides     []Override     `json:"overrides,omitempty"`
	Forwardings   []Forwarding   `json:"forwardings,omitempty"`
}

// FinalTimeline represents the final computed timeline
type FinalTimeline struct {
	Rotations []TimelineRotation `json:"rotations,omitempty"`
}

// BaseTimeline represents the base timeline before overrides
type BaseTimeline struct {
	Rotations []TimelineRotation `json:"rotations,omitempty"`
}

// TimelineRotation represents a rotation in the timeline
type TimelineRotation struct {
	ID      string           `json:"id,omitempty"`
	Name    string           `json:"name,omitempty"`
	Periods []TimelinePeriod `json:"periods,omitempty"`
}

// TimelinePeriod represents a period in a timeline rotation
type TimelinePeriod struct {
	StartDate time.Time  `json:"startDate"`
	EndDate   time.Time  `json:"endDate"`
	Recipient *Responder `json:"recipient,omitempty"`
}

// Override represents a schedule override
type Override struct {
	Alias     string    `json:"alias,omitempty"`
	User      *User     `json:"user,omitempty"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// Forwarding represents a forwarding rule
type Forwarding struct {
	FromUser  *User     `json:"fromUser,omitempty"`
	ToUser    *User     `json:"toUser,omitempty"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// Team represents an Opsgenie team
type Team struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Members     []TeamMember `json:"members,omitempty"`
}

// TeamMember represents a team member
type TeamMember struct {
	User *User  `json:"user,omitempty"`
	Role string `json:"role,omitempty"`
}

// User represents an Opsgenie user
type User struct {
	ID       string   `json:"id,omitempty"`
	Username string   `json:"username,omitempty"`
	FullName string   `json:"fullName,omitempty"`
	Email    string   `json:"email,omitempty"`
	Role     *Role    `json:"role,omitempty"`
	Blocked  bool     `json:"blocked,omitempty"`
	Verified bool     `json:"verified,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Timezone string   `json:"timezone,omitempty"`
	Locale   string   `json:"locale,omitempty"`
}

// Role represents a user role
type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// AsyncResponse represents an asynchronous operation response
type AsyncResponse struct {
	IsSuccess     bool   `json:"isSuccess"`
	Status        string `json:"status"`
	Action        string `json:"action,omitempty"`
	ProcessedAt   string `json:"processedAt,omitempty"`
	IntegrationID string `json:"integrationId,omitempty"`
	AlertID       string `json:"alertId,omitempty"`
	Alias         string `json:"alias,omitempty"`
}

// ErrorResponse represents an Opsgenie error response
type ErrorResponse struct {
	Message   string  `json:"message,omitempty"`
	Took      float64 `json:"took,omitempty"`
	RequestID string  `json:"requestId,omitempty"`
}
