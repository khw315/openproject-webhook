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

// writeLinkField writes an emoji, label, and link title if the link title is not empty.
func writeLinkField(b *strings.Builder, emoji, label string, link HALLink) {
	if title := linkTitle(link, ""); title != "" {
		b.WriteString(fmt.Sprintf("%s <b>%s:</b> %s\n", emoji, label, escapeHTML(title)))
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

	emoji := eventEmoji(event)
	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s <b>Work Package %s</b>\n", emoji, escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("📌 <b>Subject:</b> %s\n", escapeHTML(wp.Subject)))

	writeWorkPackageLinks(&b, &wp.Links)
	writeWorkPackageMeta(&b, &wp)

	if desc := truncate(wp.Description.Raw, 200); desc != "" {
		b.WriteString(fmt.Sprintf("\n📝 %s\n", escapeHTML(desc)))
	}

	if baseURL != "" && wp.ID > 0 {
		link := fmt.Sprintf("%s/work_packages/%d", baseURL, wp.ID)
		b.WriteString(fmt.Sprintf("\n🔗 <a href=\"%s\">Open in OpenProject</a>", escapeHTML(link)))
	}

	return b.String()
}

func writeWorkPackageLinks(b *strings.Builder, links *WorkPackageLinks) {
	writeLinkField(b, "🏷️", "Type", links.Type)
	writeLinkField(b, "📊", "Status", links.Status)
	writeLinkField(b, "⚡", "Priority", links.Priority)
	writeLinkField(b, "👤", "Author", links.Author)
	writeLinkField(b, "👷", "Assignee", links.Assignee)
	writeLinkField(b, "🎯", "Responsible", links.Responsible)
	writeLinkField(b, "📁", "Project", links.Project)
	writeLinkField(b, "📦", "Version", links.Version)
}

func writeWorkPackageMeta(b *strings.Builder, wp *WorkPackage) {
	if wp.PercentageDone > 0 {
		b.WriteString(fmt.Sprintf("📈 <b>Progress:</b> %d%%\n", wp.PercentageDone))
	}
	if wp.StartDate != nil && *wp.StartDate != "" {
		b.WriteString(fmt.Sprintf("📅 <b>Start:</b> %s\n", escapeHTML(*wp.StartDate)))
	}
	if wp.DueDate != nil && *wp.DueDate != "" {
		b.WriteString(fmt.Sprintf("📅 <b>Due:</b> %s\n", escapeHTML(*wp.DueDate)))
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

	emoji := eventEmoji(event)
	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s <b>Time Entry %s</b>\n", emoji, escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)

	writeLinkField(&b, "👤", "User", te.Links.User)
	b.WriteString(fmt.Sprintf("⏱️ <b>Hours:</b> %s\n", escapeHTML(te.Hours)))
	if te.SpentOn != "" {
		b.WriteString(fmt.Sprintf("📅 <b>Date:</b> %s\n", escapeHTML(te.SpentOn)))
	}
	writeLinkField(&b, "🏷️", "Activity", te.Links.Activity)
	writeLinkField(&b, "📌", "Work Package", te.Links.WorkPackage)
	writeLinkField(&b, "📁", "Project", te.Links.Project)

	if comment := truncate(te.Comment.Raw, 200); comment != "" {
		b.WriteString(fmt.Sprintf("\n💬 %s\n", escapeHTML(comment)))
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

	emoji := eventEmoji(event)
	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s <b>Project %s</b>\n", emoji, escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("📁 <b>Name:</b> %s\n", escapeHTML(proj.Name)))

	if proj.Identifier != "" {
		b.WriteString(fmt.Sprintf("🆔 <b>Identifier:</b> %s\n", escapeHTML(proj.Identifier)))
	}
	writeLinkField(&b, "📂", "Parent", proj.Links.Parent)
	writeLinkField(&b, "📊", "Status", proj.Links.Status)

	statusText := "Active"
	if !proj.Active {
		statusText = "Inactive"
	}
	b.WriteString(fmt.Sprintf("✅ <b>Active:</b> %s\n", statusText))

	if desc := truncate(proj.Description.Raw, 200); desc != "" {
		b.WriteString(fmt.Sprintf("\n📝 %s\n", escapeHTML(desc)))
	}

	if baseURL != "" && proj.Identifier != "" {
		link := fmt.Sprintf("%s/projects/%s", baseURL, proj.Identifier)
		b.WriteString(fmt.Sprintf("\n🔗 <a href=\"%s\">Open in OpenProject</a>", escapeHTML(link)))
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

	emoji := eventEmoji(event)
	actionLabel := eventLabel(event)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s <b>Membership %s</b>\n", emoji, escapeHTML(actionLabel)))
	b.WriteString(messageSeparator)

	writeLinkField(&b, "👤", "User", m.Links.Principal)
	writeLinkField(&b, "📁", "Project", m.Links.Project)

	if len(m.Links.Roles) > 0 {
		var roleNames []string
		for _, r := range m.Links.Roles {
			if r.Title != "" {
				roleNames = append(roleNames, r.Title)
			}
		}
		if len(roleNames) > 0 {
			b.WriteString(fmt.Sprintf("🔑 <b>Roles:</b> %s\n", escapeHTML(strings.Join(roleNames, ", "))))
		}
	}

	return b.String()
}

// =============================================================================
// Generic / fallback formatters
// =============================================================================

func formatGeneric(action string, _ *WebhookPayload) string {
	var b strings.Builder
	b.WriteString("🔔 <b>OpenProject Event</b>\n")
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("⚙️ <b>Action:</b> %s\n", escapeHTML(action)))
	return b.String()
}

func formatFallback(resource, event string, err error) string {
	var b strings.Builder
	emoji := eventEmoji(event)
	b.WriteString(fmt.Sprintf("%s <b>%s %s</b>\n", emoji, escapeHTML(resource), escapeHTML(eventLabel(event))))
	b.WriteString(messageSeparator)
	b.WriteString(fmt.Sprintf("⚠️ Could not parse details: %s\n", escapeHTML(err.Error())))
	return b.String()
}

// =============================================================================
// Helpers
// =============================================================================

func eventEmoji(event string) string {
	switch event {
	case "created":
		return "✨"
	case "updated":
		return "✏️"
	case "deleted":
		return "🗑️"
	default:
		return "🔔"
	}
}

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