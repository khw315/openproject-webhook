package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatMessageWorkPackage(t *testing.T) {
	wp := WorkPackage{
		ID:      101,
		Subject: "Fix security bug",
		Links: WorkPackageLinks{
			Type:     HALLink{Title: "Bug"},
			Status:   HALLink{Title: "In Progress"},
			Priority: HALLink{Title: "High"},
			Project:  HALLink{Title: "Core App"},
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
	if !strings.Contains(msg, "Fix security bug") {
		t.Errorf("expected subject in message, got %s", msg)
	}
	if !strings.Contains(msg, "https://openproject.example.com/work_packages/101") {
		t.Errorf("expected link in message, got %s", msg)
	}
}

func TestFormatMessageGeneric(t *testing.T) {
	payload := WebhookPayload{
		Action: "custom:event",
	}

	msg := FormatMessage(payload.Action, &payload, "")
	if !strings.Contains(msg, "OpenProject Event") {
		t.Errorf("expected generic header, got %s", msg)
	}
}
