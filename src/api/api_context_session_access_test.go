package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inteliagenciadigital/quepasa/models"
	"github.com/inteliagenciadigital/quepasa/runtime"
)

func TestCanonicalSessionGetAllowsEnabledContextAccess(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	seedSharedContextSession(t)

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions/get", []byte(`{"token":"shared-context-session"}`), "viewer@example.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected session get to return 200 for context access, got %d with body %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Server struct {
			Token     string `json:"token"`
			ContextID string `json:"contextid"`
		} `json:"server"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode session get response: %v", err)
	}
	if payload.Server.Token != "shared-context-session" {
		t.Fatalf("expected shared-context-session, got %q", payload.Server.Token)
	}
	if payload.Server.ContextID != "tenant-shared" {
		t.Fatalf("expected tenant-shared, got %q", payload.Server.ContextID)
	}
}

func TestCanonicalSessionPatchAllowsSafeContextAccess(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	seedSharedContextSession(t)

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodPatch, "/api/sessions?token=shared-context-session", []byte(`{"calls":0}`), "viewer@example.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected session patch to return 200 for context access, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestCanonicalSessionPatchRejectsRestrictedContextAccessFields(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	seedSharedContextSession(t)

	router := newCanonicalTestRouter()
	body := []byte(`{"username":"viewer@example.test"}`)
	req := newCanonicalAuthRequest(t, http.MethodPatch, "/api/sessions?token=shared-context-session", body, "viewer@example.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected restricted context patch to return 403, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestCanonicalSessionDeleteStillRequiresOwnership(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	seedSharedContextSession(t)

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodDelete, "/api/sessions?token=shared-context-session", nil, "viewer@example.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected shared context delete to remain forbidden, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func seedSharedContextSession(t *testing.T) {
	t.Helper()

	CreateTestUser(t, "owner@example.test", "Password123!")
	CreateTestUser(t, "viewer@example.test", "Password123!")

	server := &models.QpServer{Token: "shared-context-session", Verified: true}
	server.SetUser("owner@example.test")
	server.SetContextId("tenant-shared")
	server.SetWId("551155501234:1@s.whatsapp.net")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add shared context server: %v", err)
	}
	if _, err := runtime.LoadSessionRecord(server); err != nil {
		t.Fatalf("load live shared context server: %v", err)
	}
	if err := models.WhatsappService.DB.UserContexts.Upsert(&models.QpUserContextAccess{
		Username:  "viewer@example.test",
		ContextID: "tenant-shared",
		Label:     "Shared tenant",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("upsert user context: %v", err)
	}
}

func TestCanonicalSessionGetRejectsContextlessNonOwner(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.test", "Password123!")
	CreateTestUser(t, "viewer@example.test", "Password123!")

	server := &models.QpServer{Token: "contextless-session", Verified: true}
	server.SetUser("owner@example.test")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add contextless server: %v", err)
	}

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodPost, "/api/sessions/get", []byte(`{"token":"contextless-session"}`), "viewer@example.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected contextless non-owner session get to return 403, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestCanonicalSessionsGetQueryTokenAllowsEnabledContextAccess(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	seedSharedContextSession(t)

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodGet, "/api/sessions?token=shared-context-session", nil, "viewer@example.test")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected session query token get to return 200 for context access, got %d with body %s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("shared-context-session")) {
		t.Fatalf("expected response to include shared context token, got %s", rec.Body.String())
	}
}
