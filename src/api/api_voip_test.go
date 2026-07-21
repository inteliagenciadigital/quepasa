package api

import (
	"testing"

	"github.com/inteliagenciadigital/quepasa/models"
	"github.com/inteliagenciadigital/quepasa/runtime"
)

func TestGetOwnedOrContextServerRecordAllowsOwner(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	owner := CreateTestUser(t, "owner@example.test", "Password123!")

	server := &models.QpServer{Token: "owner-token", Verified: true}
	server.SetUser(owner.Username)
	server.SetContextId("tenant-a")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}

	record, err := runtime.GetOwnedOrContextServerRecord(owner, "owner-token")
	if err != nil {
		t.Fatalf("expected owner access, got error: %v", err)
	}
	if record.Token != "owner-token" {
		t.Fatalf("expected owner-token, got %q", record.Token)
	}
}

func TestGetOwnedOrContextServerRecordAllowsEnabledContextAccess(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	owner := CreateTestUser(t, "owner@example.test", "Password123!")
	viewer := CreateTestUser(t, "viewer@example.test", "Password123!")

	server := &models.QpServer{Token: "shared-token", Verified: true}
	server.SetUser(owner.Username)
	server.SetContextId("tenant-shared")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}

	if err := models.WhatsappService.DB.UserContexts.Upsert(&models.QpUserContextAccess{
		Username:  viewer.Username,
		ContextID: "tenant-shared",
		Label:     "Shared tenant",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("upsert context access: %v", err)
	}

	record, err := runtime.GetOwnedOrContextServerRecord(viewer, "shared-token")
	if err != nil {
		t.Fatalf("expected context access, got error: %v", err)
	}
	if record.Token != "shared-token" {
		t.Fatalf("expected shared-token, got %q", record.Token)
	}
}

func TestGetOwnedOrContextServerRecordRejectsDisabledContextAccess(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	owner := CreateTestUser(t, "owner@example.test", "Password123!")
	viewer := CreateTestUser(t, "viewer@example.test", "Password123!")

	server := &models.QpServer{Token: "disabled-context-token", Verified: true}
	server.SetUser(owner.Username)
	server.SetContextId("tenant-disabled")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}

	if err := models.WhatsappService.DB.UserContexts.Upsert(&models.QpUserContextAccess{
		Username:  viewer.Username,
		ContextID: "tenant-disabled",
		Label:     "Disabled tenant",
		Enabled:   false,
	}); err != nil {
		t.Fatalf("upsert context access: %v", err)
	}

	if _, err := runtime.GetOwnedOrContextServerRecord(viewer, "disabled-context-token"); err == nil {
		t.Fatalf("expected disabled context access to be rejected")
	}
}
