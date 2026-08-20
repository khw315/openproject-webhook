package main

import "encoding/json"

// =============================================================================
// Generic Webhook Payload
// =============================================================================

// WebhookPayload is the top-level structure sent by OpenProject webhooks.
// The "action" field indicates the event type (e.g., "work_package:created").
// The actual resource data is present in a field named after the resource type.
type WebhookPayload struct {
	Action       string          `json:"action"`
	WorkPackage  json.RawMessage `json:"work_package,omitempty"`
	TimeEntry    json.RawMessage `json:"time_entry,omitempty"`
	Project      json.RawMessage `json:"project,omitempty"`
	Membership   json.RawMessage `json:"membership,omitempty"`
	Attachment   json.RawMessage `json:"attachment,omitempty"`
	WikiPage     json.RawMessage `json:"wiki_page,omitempty"`
	News         json.RawMessage `json:"news,omitempty"`
}

// =============================================================================
// HAL Link
// =============================================================================

// HALLink represents a single HAL+JSON link with href and title.
type HALLink struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

// =============================================================================
// Work Package
// =============================================================================

// WorkPackageLinks holds the _links section of a work package resource.
type WorkPackageLinks struct {
	Self       HALLink `json:"self"`
	Project    HALLink `json:"project"`
	Type       HALLink `json:"type"`
	Status     HALLink `json:"status"`
	Priority   HALLink `json:"priority"`
	Author     HALLink `json:"author"`
	Assignee   HALLink `json:"assignee"`
	Responsible HALLink `json:"responsible"`
	Version    HALLink `json:"version"`
}

// TextNode represents a rich-text field in OpenProject (description, etc.).
type TextNode struct {
	Format string `json:"format"`
	Raw    string `json:"raw"`
	HTML   string `json:"html"`
}

// WorkPackage represents the work package resource from the webhook payload.
type WorkPackage struct {
	Type           string           `json:"_type"`
	ID             int              `json:"id"`
	LockVersion    int              `json:"lockVersion"`
	Subject        string           `json:"subject"`
	Description    TextNode         `json:"description"`
	StartDate      *string          `json:"startDate"`
	DueDate        *string          `json:"dueDate"`
	DerivedStartDate *string        `json:"derivedStartDate"`
	DerivedDueDate *string          `json:"derivedDueDate"`
	EstimatedTime  *string          `json:"estimatedTime"`
	PercentageDone int              `json:"percentageDone"`
	CreatedAt      string           `json:"createdAt"`
	UpdatedAt      string           `json:"updatedAt"`
	Links          WorkPackageLinks `json:"_links"`
}

// =============================================================================
// Time Entry
// =============================================================================

// TimeEntryLinks holds the _links section of a time entry resource.
type TimeEntryLinks struct {
	Self        HALLink `json:"self"`
	Project     HALLink `json:"project"`
	WorkPackage HALLink `json:"workPackage"`
	User        HALLink `json:"user"`
	Activity    HALLink `json:"activity"`
}

// TimeEntry represents the time entry resource from the webhook payload.
type TimeEntry struct {
	Type      string         `json:"_type"`
	ID        int            `json:"id"`
	Hours     string         `json:"hours"`
	Comment   TextNode       `json:"comment"`
	SpentOn   string         `json:"spentOn"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
	Links     TimeEntryLinks `json:"_links"`
}

// =============================================================================
// Project
// =============================================================================

// ProjectLinks holds the _links section of a project resource.
type ProjectLinks struct {
	Self   HALLink `json:"self"`
	Parent HALLink `json:"parent"`
	Status HALLink `json:"status"`
}

// ProjectResource represents the project resource from the webhook payload.
type ProjectResource struct {
	Type        string       `json:"_type"`
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Identifier  string       `json:"identifier"`
	Description TextNode     `json:"description"`
	Public      bool         `json:"public"`
	Active      bool         `json:"active"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
	Links       ProjectLinks `json:"_links"`
}

// =============================================================================
// Membership
// =============================================================================

// MembershipLinks holds the _links section of a membership resource.
type MembershipLinks struct {
	Self      HALLink   `json:"self"`
	Project   HALLink   `json:"project"`
	Principal HALLink   `json:"principal"`
	Roles     []HALLink `json:"roles"`
}

// MembershipResource represents the membership resource from the webhook payload.
type MembershipResource struct {
	Type      string          `json:"_type"`
	ID        int             `json:"id"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	Links     MembershipLinks `json:"_links"`
}
