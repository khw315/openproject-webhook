package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

const messageSeparator = "━━━━━━━━━━━━━━━━━━━━━\n"

// escapeHTML escapes special HTML characters in a string for Telegram HTML parse mode.
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// linkTitle returns the title of a HAL link, or fallback if empty.
func linkTitle(link HALLink, fallback string) string {
	if link.Title != "" {
		return link.Title
	}
	return fallback
}

// writeLinkField writes a label and link title if the link title is not empty.
func writeLinkField(b *strings.Builder, label string, link HALLink) {
	if title := linkTitle(link, ""); title != "" {
		b.WriteString(fmt.Sprintf("<b>%s:</b> %s\n", label, escapeHTML(title)))
	}
}

// truncate limits a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// =============================================================================
// Main formatter dispatcher
// =============================================================================

// FormatMessage takes the webhook action and raw payload, and returns a formatted
// Telegram message string in HTML format.
func FormatMessage(action string, payload *WebhookPayload, openProjectURL string) string {
	parts := strings.SplitN(action, ":", 2)
	if len(parts) != 2 {
		return formatGeneric(action, payload)
	}

	resource := parts[0]
	event := parts[1]

	switch resource {
	case "work_package":
		return formatWorkPackage(event, payload, openProjectURL)
	case "work_package_comment", "comment", "activity", "journal":
		return formatWorkPackageComment(event, payload, openProjectURL)
	case "time_entry":
		return formatTimeEntry(event, payload)
	case "project":
		return formatProject(event, payload, openProjectURL)
	case "membership":
		return formatMembership(event, payload)
	default:
		return formatGeneric(action, payload)
	}
}

// =============================================================================
// Work Package formatter
// =============================================================================

func formatWorkPackage(event string, payload *WebhookPayload, baseURL string) string {
	var wp WorkPackage
	if err := unmarshalSafe(payload.WorkPackage, &wp); err != nil {
		return formatFallback("Work Package", event, err)
	}

	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>Work Package %s</b>\n", escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("<b>Subject:</b> %s\n", escapeHTML(wp.Subject)))

	writeWorkPackageLinks(&b, &wp.Links)
	writeWorkPackageMeta(&b, &wp)

	if desc := truncate(wp.Description.Raw, 200); desc != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", escapeHTML(desc)))
	}

	if baseURL != "" && wp.ID > 0 {
		link := fmt.Sprintf("%s/work_packages/%d", baseURL, wp.ID)
		b.WriteString(fmt.Sprintf("\n<a href=\"%s\">Open in OpenProject</a>", escapeHTML(link)))
	}

	return b.String()
}

func writeWorkPackageLinks(b *strings.Builder, links *WorkPackageLinks) {
	writeLinkField(b, "Type", links.Type)
	writeLinkField(b, "Status", links.Status)
	writeLinkField(b, "Priority", links.Priority)
	writeLinkField(b, "Author", links.Author)
	writeLinkField(b, "Assignee", links.Assignee)
	writeLinkField(b, "Responsible", links.Responsible)
	writeLinkField(b, "Project", links.Project)
	writeLinkField(b, "Version", links.Version)
}

func writeWorkPackageMeta(b *strings.Builder, wp *WorkPackage) {
	if wp.PercentageDone > 0 {
		b.WriteString(fmt.Sprintf("<b>Progress:</b> %d%%\n", wp.PercentageDone))
	}
	if wp.StartDate != nil && *wp.StartDate != "" {
		b.WriteString(fmt.Sprintf("<b>Start:</b> %s\n", escapeHTML(*wp.StartDate)))
	}
	if wp.DueDate != nil && *wp.DueDate != "" {
		b.WriteString(fmt.Sprintf("<b>Due:</b> %s\n", escapeHTML(*wp.DueDate)))
	}
}

// =============================================================================
// Time Entry formatter
// =============================================================================

func formatTimeEntry(event string, payload *WebhookPayload) string {
	var te TimeEntry
	if err := unmarshalSafe(payload.TimeEntry, &te); err != nil {
		return formatFallback("Time Entry", event, err)
	}

	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>Time Entry %s</b>\n", escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)

	writeLinkField(&b, "User", te.Links.User)
	b.WriteString(fmt.Sprintf("<b>Hours:</b> %s\n", escapeHTML(te.Hours)))
	if te.SpentOn != "" {
		b.WriteString(fmt.Sprintf("<b>Date:</b> %s\n", escapeHTML(te.SpentOn)))
	}
	writeLinkField(&b, "Activity", te.Links.Activity)
	writeLinkField(&b, "Work Package", te.Links.WorkPackage)
	writeLinkField(&b, "Project", te.Links.Project)

	if comment := truncate(te.Comment.Raw, 200); comment != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", escapeHTML(comment)))
	}

	return b.String()
}

// =============================================================================
// Project formatter
// =============================================================================

func formatProject(event string, payload *WebhookPayload, baseURL string) string {
	var proj ProjectResource
	if err := unmarshalSafe(payload.Project, &proj); err != nil {
		return formatFallback("Project", event, err)
	}

	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>Project %s</b>\n", escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("<b>Name:</b> %s\n", escapeHTML(proj.Name)))

	if proj.Identifier != "" {
		b.WriteString(fmt.Sprintf("<b>Identifier:</b> %s\n", escapeHTML(proj.Identifier)))
	}
	writeLinkField(&b, "Parent", proj.Links.Parent)
	writeLinkField(&b, "Status", proj.Links.Status)

	statusText := "Active"
	if !proj.Active {
		statusText = "Inactive"
	}
	b.WriteString(fmt.Sprintf("<b>Active:</b> %s\n", statusText))

	if desc := truncate(proj.Description.Raw, 200); desc != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", escapeHTML(desc)))
	}

	if baseURL != "" && proj.Identifier != "" {
		link := fmt.Sprintf("%s/projects/%s", baseURL, proj.Identifier)
		b.WriteString(fmt.Sprintf("\n<a href=\"%s\">Open in OpenProject</a>", escapeHTML(link)))
	}

	return b.String()
}

// =============================================================================
// Membership formatter
// =============================================================================

func formatMembership(event string, payload *WebhookPayload) string {
	var m MembershipResource
	if err := unmarshalSafe(payload.Membership, &m); err != nil {
		return formatFallback("Membership", event, err)
	}

	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>Membership %s</b>\n", escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)

	writeLinkField(&b, "User", m.Links.Principal)
	writeLinkField(&b, "Project", m.Links.Project)

	if len(m.Links.Roles) > 0 {
		var roleNames []string
		for _, r := range m.Links.Roles {
			if r.Title != "" {
				roleNames = append(roleNames, r.Title)
			}
		}
		if len(roleNames) > 0 {
			b.WriteString(fmt.Sprintf("<b>Roles:</b> %s\n", escapeHTML(strings.Join(roleNames, ", "))))
		}
	}

	return b.String()
}

// =============================================================================
// Work Package Comment / Activity formatter
// =============================================================================

func formatWorkPackageComment(event string, payload *WebhookPayload, baseURL string) string {
	var wp WorkPackage
	_ = unmarshalSafe(payload.WorkPackage, &wp)

	var act ActivityResource
	var commentText, authorName string

	extractActivity := func(data []byte) bool {
		if len(data) == 0 {
			return false
		}
		if err := jsonDecode(data, &act); err == nil {
			if act.Comment.Raw != "" {
				commentText = act.Comment.Raw
			} else if act.Raw != "" {
				commentText = act.Raw
			}
			if act.Links.User.Title != "" {
				authorName = act.Links.User.Title
			} else if act.Links.Author.Title != "" {
				authorName = act.Links.Author.Title
			}
			return commentText != "" || authorName != ""
		}
		var tn TextNode
		if err := jsonDecode(data, &tn); err == nil && tn.Raw != "" {
			commentText = tn.Raw
			return true
		}
		return false
	}

	if !extractActivity(payload.Comment) {
		if !extractActivity(payload.Activity) {
			extractActivity(payload.Journal)
		}
	}

	var b strings.Builder
	b.WriteString("<b>Work Package Comment</b>\n")
	b.WriteString(messageSeparator)

	if wp.Subject != "" {
		b.WriteString(fmt.Sprintf("<b>Subject:</b> %s\n", escapeHTML(wp.Subject)))
	} else if act.Links.WorkPackage.Title != "" {
		b.WriteString(fmt.Sprintf("<b>Work Package:</b> %s\n", escapeHTML(act.Links.WorkPackage.Title)))
	}

	if title := linkTitle(wp.Links.Project, linkTitle(act.Links.Project, "")); title != "" {
		b.WriteString(fmt.Sprintf("<b>Project:</b> %s\n", escapeHTML(title)))
	}

	if authorName != "" {
		b.WriteString(fmt.Sprintf("<b>Author:</b> %s\n", escapeHTML(authorName)))
	} else if title := linkTitle(wp.Links.Author, linkTitle(wp.Links.Assignee, "")); title != "" {
		b.WriteString(fmt.Sprintf("<b>Author:</b> %s\n", escapeHTML(title)))
	}

	if commentText != "" {
		if truncated := truncate(commentText, 300); truncated != "" {
			b.WriteString(fmt.Sprintf("\n<i>\"%s\"</i>\n", escapeHTML(truncated)))
		}
	}

	wpID := wp.ID
	if wpID == 0 && act.Links.WorkPackage.Href != "" {
		parts := strings.Split(strings.TrimRight(act.Links.WorkPackage.Href, "/"), "/")
		if len(parts) > 0 {
			var id int
			if _, err := fmt.Sscanf(parts[len(parts)-1], "%d", &id); err == nil {
				wpID = id
			}
		}
	}

	if baseURL != "" && wpID > 0 {
		link := fmt.Sprintf("%s/work_packages/%d", baseURL, wpID)
		b.WriteString(fmt.Sprintf("\n<a href=\"%s\">Open in OpenProject</a>", escapeHTML(link)))
	}

	return b.String()
}

// =============================================================================
// Generic / fallback formatters
// =============================================================================

func formatGeneric(action string, _ *WebhookPayload) string {
	var b strings.Builder
	b.WriteString("<b>OpenProject Event</b>\n")
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("<b>Action:</b> %s\n", escapeHTML(action)))
	return b.String()
}

func formatFallback(resource, event string, err error) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>%s %s</b>\n", escapeHTML(resource), escapeHTML(eventLabel(event))))
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("Could not parse details: %s\n", escapeHTML(err.Error())))
	return b.String()
}

// =============================================================================
// Helpers
// =============================================================================

func eventLabel(event string) string {
	switch event {
	case "created":
		return "Created"
	case "updated":
		return "Updated"
	case "deleted":
		return "Deleted"
	default:
		if len(event) > 0 {
			return strings.ToUpper(event[:1]) + event[1:]
		}
		return event
	}
}

func unmarshalSafe(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty resource data")
	}
	return jsonUnmarshal(data, v)
}

// jsonUnmarshal wraps json.Unmarshal with a cleaner error message.
func jsonUnmarshal(data []byte, v interface{}) error {
	if err := jsonDecode(data, v); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

func jsonDecode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}