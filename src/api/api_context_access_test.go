package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inteliagenciadigital/quepasa/models"
)

func TestAuthenticatedContextSessionsListsSharedContextSessions(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")
	CreateTestUser(t, "viewer@example.com", "Password123!")

	shared := &models.QpServer{Token: "shared-session", Verified: true}
	shared.SetUser("owner@example.com")
	shared.SetContextId("context-shared")
	shared.SetWId("5511999999999:7@s.whatsapp.net")
	if err := models.WhatsappService.DB.Servers.Add(shared); err != nil {
		t.Fatalf("add shared server: %v", err)
	}

	private := &models.QpServer{Token: "private-session", Verified: true}
	private.SetUser("owner@example.com")
	private.SetContextId("context-private")
	private.SetWId("5511888888888:1@s.whatsapp.net")
	if err := models.WhatsappService.DB.Servers.Add(private); err != nil {
		t.Fatalf("add private server: %v", err)
	}

	if err := models.WhatsappService.DB.UserContexts.Upsert(&models.QpUserContextAccess{
		Username:  "viewer@example.com",
		ContextID: "context-shared",
		Label:     "Cliente compartilhado",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("upsert user context: %v", err)
	}

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodGet, "/api/auth/contexts/sessions?contextid=context-shared", nil, "viewer@example.com")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected context sessions to return 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Servers []struct {
			Token     string `json:"token"`
			User      string `json:"user"`
			ContextID string `json:"contextid"`
		} `json:"servers"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode context sessions response: %v", err)
	}

	if payload.Total != 1 || len(payload.Servers) != 1 {
		t.Fatalf("expected one shared context session, got total=%d len=%d", payload.Total, len(payload.Servers))
	}
	if payload.Servers[0].Token != "shared-session" {
		t.Fatalf("expected shared-session, got %q", payload.Servers[0].Token)
	}
	if payload.Servers[0].User != "owner@example.com" {
		t.Fatalf("expected owner user in shared session summary, got %q", payload.Servers[0].User)
	}
	if payload.Servers[0].ContextID != "context-shared" {
		t.Fatalf("expected context-shared, got %q", payload.Servers[0].ContextID)
	}
}

func TestAuthenticatedContextAccessUpsertUsesAuthenticatedUser(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "viewer@example.com", "Password123!")
	CreateTestUser(t, "other@example.com", "Password123!")

	router := newCanonicalTestRouter()
	body := []byte(`{"username":"other@example.com","contextid":"context-shared","label":"Cliente compartilhado","enabled":true}`)
	req := newCanonicalAuthRequest(t, http.MethodPost, "/api/auth/contexts", body, "viewer@example.com")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected authenticated context upsert to return 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	viewerAccess, err := models.WhatsappService.DB.UserContexts.Find("viewer@example.com", "context-shared")
	if err != nil {
		t.Fatalf("find viewer context: %v", err)
	}
	if viewerAccess == nil {
		t.Fatalf("expected context to be saved for authenticated user")
	}
	if viewerAccess.Username != "viewer@example.com" {
		t.Fatalf("expected authenticated username, got %q", viewerAccess.Username)
	}

	otherAccess, err := models.WhatsappService.DB.UserContexts.Find("other@example.com", "context-shared")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("find other context: %v", err)
	}
	if otherAccess != nil {
		t.Fatalf("did not expect request body username to receive access")
	}
}

func TestAuthenticatedContextAccessDeleteUsesAuthenticatedUser(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "viewer@example.com", "Password123!")
	CreateTestUser(t, "other@example.com", "Password123!")

	for _, username := range []string{"viewer@example.com", "other@example.com"} {
		if err := models.WhatsappService.DB.UserContexts.Upsert(&models.QpUserContextAccess{
			Username:  username,
			ContextID: "context-shared",
			Label:     "Cliente compartilhado",
			Enabled:   true,
		}); err != nil {
			t.Fatalf("upsert context for %s: %v", username, err)
		}
	}

	router := newCanonicalTestRouter()
	body := bytes.TrimSpace([]byte(`{"username":"other@example.com","contextid":"context-shared"}`))
	req := newCanonicalAuthRequest(t, http.MethodDelete, "/api/auth/contexts", body, "viewer@example.com")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected authenticated context delete to return 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	viewerAccess, err := models.WhatsappService.DB.UserContexts.Find("viewer@example.com", "context-shared")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("find viewer context: %v", err)
	}
	if viewerAccess != nil {
		t.Fatalf("expected authenticated user's context to be removed")
	}

	otherAccess, err := models.WhatsappService.DB.UserContexts.Find("other@example.com", "context-shared")
	if err != nil {
		t.Fatalf("find other context: %v", err)
	}
	if otherAccess == nil {
		t.Fatalf("did not expect request body username context to be removed")
	}
}

func TestAuthenticatedContextAccessDeleteRemovesOwnedContextSessions(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "owner@example.com", "Password123!")
	CreateTestUser(t, "other@example.com", "Password123!")

	for _, token := range []string{"owner-session-1", "owner-session-2"} {
		server := &models.QpServer{Token: token, Verified: true}
		server.SetUser("owner@example.com")
		server.SetContextId("context-shared")
		server.SetWId(token + "@s.whatsapp.net")
		if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
			t.Fatalf("add owner server %s: %v", token, err)
		}
	}

	other := &models.QpServer{Token: "other-session", Verified: true}
	other.SetUser("other@example.com")
	other.SetContextId("context-shared")
	other.SetWId("other-session@s.whatsapp.net")
	if err := models.WhatsappService.DB.Servers.Add(other); err != nil {
		t.Fatalf("add other server: %v", err)
	}

	if err := models.WhatsappService.DB.UserContexts.Upsert(&models.QpUserContextAccess{
		Username:  "owner@example.com",
		ContextID: "context-shared",
		Label:     "Cliente compartilhado",
		Enabled:   true,
	}); err != nil {
		t.Fatalf("upsert owner context: %v", err)
	}

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodDelete, "/api/auth/contexts?contextid=context-shared", nil, "owner@example.com")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected authenticated context delete to return 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Removed         bool `json:"removed"`
		RemovedAccess   bool `json:"removed_access"`
		DeletedSessions int  `json:"deleted_sessions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Removed || !payload.RemovedAccess || payload.DeletedSessions != 2 {
		t.Fatalf("unexpected delete response: %+v", payload)
	}

	for _, token := range []string{"owner-session-1", "owner-session-2"} {
		if server, _ := models.WhatsappService.DB.Servers.FindByToken(token); server != nil {
			t.Fatalf("expected owned session %s to be deleted", token)
		}
	}
	if server, err := models.WhatsappService.DB.Servers.FindByToken("other-session"); err != nil || server == nil {
		t.Fatalf("expected other user's session to remain, server=%v err=%v", server, err)
	}

	ownerAccess, err := models.WhatsappService.DB.UserContexts.Find("owner@example.com", "context-shared")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("find owner context: %v", err)
	}
	if ownerAccess != nil {
		t.Fatalf("expected owner access to be removed after removing context usage")
	}
}

func TestAuthenticatedContextSessionsRejectsUnauthorizedContext(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	CreateTestUser(t, "viewer@example.com", "Password123!")

	router := newCanonicalTestRouter()
	req := newCanonicalAuthRequest(t, http.MethodGet, "/api/auth/contexts/sessions?contextid=context-private", nil, "viewer@example.com")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected unauthorized context to return 403, got %d with body %s", rec.Code, rec.Body.String())
	}
}
