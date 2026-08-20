package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatMessageWorkPackageFull(t *testing.T) {
	startDate := "2026-08-20"
	dueDate := "2026-08-25"
	wp := WorkPackage{
		ID:             101,
		Subject:        "Fix security bug",
		PercentageDone: 50,
		StartDate:      &startDate,
		DueDate:        &dueDate,
		Description:    TextNode{Raw: "A detailed description that explains the security vulnerability."},
		Links: WorkPackageLinks{
			Type:        HALLink{Title: "Bug"},
			Status:      HALLink{Title: "In Progress"},
			Priority:    HALLink{Title: "High"},
			Author:      HALLink{Title: "Alice"},
			Assignee:    HALLink{Title: "Bob"},
			Responsible: HALLink{Title: "Charlie"},
			Project:     HALLink{Title: "Core App"},
			Version:     HALLink{Title: "v1.0.0"},
		},
	}
	rawWP, _ := json.Marshal(wp)

	payload := WebhookPayload{
		Action:      "work_package:created",
		WorkPackage: rawWP,
	}

	msg := FormatMessage(payload.Action, &payload, "https://openproject.example.com")
	if !strings.Contains(msg, "Work Package Created") {
		t.Errorf("expected header in message, got %s", msg)
	}
	if !strings.Contains(msg, "50%") {
		t.Errorf("expected progress, got %s", msg)
	}
	if !strings.Contains(msg, "Alice") || !strings.Contains(msg, "Bob") || !strings.Contains(msg, "Charlie") {
		t.Errorf("expected user names, got %s", msg)
	}
	if !strings.Contains(msg, "v1.0.0") {
		t.Errorf("expected version, got %s", msg)
	}
}

func TestFormatMessageWorkPackageFallback(t *testing.T) {
	payload := WebhookPayload{
		Action:      "work_package:updated",
		WorkPackage: json.RawMessage("invalid json"),
	}

	msg := FormatMessage(payload.Action, &payload, "")
	if !strings.Contains(msg, "Could not parse details") {
		t.Errorf("expected fallback error, got %s", msg)
	}
}

func TestFormatMessageTimeEntry(t *testing.T) {
	te := TimeEntry{
		ID:      5,
		Hours:   "4.5",
		SpentOn: "2026-08-20",
		Comment: TextNode{Raw: "Refactored payment module"},
		Links: TimeEntryLinks{
			User:        HALLink{Title: "Alice"},
			Activity:    HALLink{Title: "Development"},
			WorkPackage: HALLink{Title: "Feature #12"},
			Project:     HALLink{Title: "Website"},
		},
	}
	rawTE, _ := json.Marshal(te)

	payload := WebhookPayload{
		Action:    "time_entry:created",
		TimeEntry: rawTE,
	}

	msg := FormatMessage(payload.Action, &payload, "")
	if !strings.Contains(msg, "Time Entry Created") {
		t.Errorf("expected time entry header, got %s", msg)
	}
	if !strings.Contains(msg, "4.5") || !strings.Contains(msg, "Alice") {
		t.Errorf("expected hours and user, got %s", msg)
	}

	// Test fallback
	badPayload := WebhookPayload{Action: "time_entry:created", TimeEntry: json.RawMessage("bad")}
	msgBad := FormatMessage(badPayload.Action, &badPayload, "")
	if !strings.Contains(msgBad, "Could not parse details") {
		t.Errorf("expected fallback, got %s", msgBad)
	}
}

func TestFormatMessageProject(t *testing.T) {
	proj := ProjectResource{
		ID:          1,
		Name:        "Secret Project",
		Identifier:  "secret-proj",
		Active:      false,
		Description: TextNode{Raw: "Project description goes here"},
		Links: ProjectLinks{
			Parent: HALLink{Title: "Parent Org"},
			Status: HALLink{Title: "On Track"},
		},
	}
	rawProj, _ := json.Marshal(proj)

	payload := WebhookPayload{
		Action:  "project:updated",
		Project: rawProj,
	}

	msg := FormatMessage(payload.Action, &payload, "https://op.example.com")
	if !strings.Contains(msg, "Project Updated") {
		t.Errorf("expected project header, got %s", msg)
	}
	if !strings.Contains(msg, "Inactive") {
		t.Errorf("expected inactive status, got %s", msg)
	}
	if !strings.Contains(msg, "https://op.example.com/projects/secret-proj") {
		t.Errorf("expected link, got %s", msg)
	}

	// Test fallback
	badPayload := WebhookPayload{Action: "project:created", Project: json.RawMessage("bad")}
	msgBad := FormatMessage(badPayload.Action, &badPayload, "")
	if !strings.Contains(msgBad, "Could not parse details") {
		t.Errorf("expected fallback, got %s", msgBad)
	}
}

func TestFormatMessageMembership(t *testing.T) {
	m := MembershipResource{
		ID: 10,
		Links: MembershipLinks{
			Principal: HALLink{Title: "John Doe"},
			Project:   HALLink{Title: "Backend API"},
			Roles: []HALLink{
				{Title: "Project Admin"},
				{Title: "Developer"},
			},
		},
	}
	rawM, _ := json.Marshal(m)

	payload := WebhookPayload{
		Action:     "membership:created",
		Membership: rawM,
	}

	msg := FormatMessage(payload.Action, &payload, "")
	if !strings.Contains(msg, "Membership Created") {
		t.Errorf("expected membership header, got %s", msg)
	}
	if !strings.Contains(msg, "Project Admin, Developer") {
		t.Errorf("expected roles, got %s", msg)
	}

	// Test fallback
	badPayload := WebhookPayload{Action: "membership:created", Membership: json.RawMessage("bad")}
	msgBad := FormatMessage(badPayload.Action, &badPayload, "")
	if !strings.Contains(msgBad, "Could not parse details") {
		t.Errorf("expected fallback, got %s", msgBad)
	}
}

func TestFormatMessageWorkPackageComment(t *testing.T) {
	wp := WorkPackage{
		ID:      101,
		Subject: "Fix security bug",
		Links: WorkPackageLinks{
			Project: HALLink{Title: "Core App"},
		},
	}
	rawWP, _ := json.Marshal(wp)

	act := ActivityResource{
		ID:      12,
		Comment: TextNode{Raw: "This has been resolved in master branch."},
		Links: ActivityLinks{
			User: HALLink{Title: "Alice"},
		},
	}
	rawAct, _ := json.Marshal(act)

	payload := WebhookPayload{
		Action:      "work_package_comment:comment",
		WorkPackage: rawWP,
		Comment:     rawAct,
	}

	msg := FormatMessage(payload.Action, &payload, "https://openproject.example.com")
	if !strings.Contains(msg, "Work Package Comment") {
		t.Errorf("expected Work Package Comment in message, got %s", msg)
	}
	if !strings.Contains(msg, "Fix security bug") {
		t.Errorf("expected Subject in message, got %s", msg)
	}
	if !strings.Contains(msg, "Core App") {
		t.Errorf("expected Project in message, got %s", msg)
	}
	if !strings.Contains(msg, "Alice") {
		t.Errorf("expected Author Alice in message, got %s", msg)
	}
	if !strings.Contains(msg, "This has been resolved in master branch.") {
		t.Errorf("expected comment text, got %s", msg)
	}
	if !strings.Contains(msg, "https://openproject.example.com/work_packages/101") {
		t.Errorf("expected OpenProject link, got %s", msg)
	}

	// Test with activity in payload.Activity and links inside activity
	actOnly := ActivityResource{
		Raw: "Another comment",
		Links: ActivityLinks{
			WorkPackage: HALLink{Href: "/api/v3/work_packages/202", Title: "Task 202"},
			Project:     HALLink{Title: "Sub project"},
			Author:      HALLink{Title: "Bob"},
		},
	}
	rawActOnly, _ := json.Marshal(actOnly)
	payloadAct := WebhookPayload{
		Action:   "work_package_comment:comment",
		Activity: rawActOnly,
	}
	msgAct := FormatMessage(payloadAct.Action, &payloadAct, "https://openproject.example.com")
	if !strings.Contains(msgAct, "Task 202") || !strings.Contains(msgAct, "Sub project") || !strings.Contains(msgAct, "Bob") {
		t.Errorf("expected fallback fields from activity resource, got %s", msgAct)
	}
	if !strings.Contains(msgAct, "/work_packages/202") {
		t.Errorf("expected link parsed from href, got %s", msgAct)
	}
}

func TestFormatMessageGenericAndFallback(t *testing.T) {
	payload := WebhookPayload{Action: "ping"}
	msg1 := FormatMessage("ping", &payload, "")
	if !strings.Contains(msg1, "OpenProject Event") {
		t.Errorf("expected generic format for single word action, got %s", msg1)
	}

	msg2 := FormatMessage("unknown_resource:action", &payload, "")
	if !strings.Contains(msg2, "OpenProject Event") {
		t.Errorf("expected generic format for unknown resource, got %s", msg2)
	}
}

func TestEventLabel(t *testing.T) {
	if eventLabel("created") != "Created" || eventLabel("updated") != "Updated" || eventLabel("deleted") != "Deleted" || eventLabel("custom") != "Custom" || eventLabel("") != "" {
		t.Error("unexpected label")
	}
}

func TestTruncateAndLinkTitle(t *testing.T) {
	longText := strings.Repeat("a", 300)
	truncated := truncate(longText, 10)
	if len(truncated) != 13 || !strings.HasSuffix(truncated, "...") {
		t.Errorf("expected truncated string with ..., got %s", truncated)
	}

	short := truncate("hello", 10)
	if short != "hello" {
		t.Errorf("expected short string unchanged, got %s", short)
	}

	link := HALLink{Title: "My Link"}
	if linkTitle(link, "fallback") != "My Link" {
		t.Errorf("expected My Link")
	}
	emptyLink := HALLink{}
	if linkTitle(emptyLink, "fallback") != "fallback" {
		t.Errorf("expected fallback")
	}
}

func TestUnmarshalSafe(t *testing.T) {
	var target struct{ Name string }
	if err := unmarshalSafe(nil, &target); err == nil {
		t.Error("expected error for empty data")
	}
}
