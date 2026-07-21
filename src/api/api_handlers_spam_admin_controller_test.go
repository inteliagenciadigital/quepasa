package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inteliagenciadigital/quepasa/models"
)

func TestSpamAdminStatusReportsMissingMasterKey(t *testing.T) {
	restore := SetupTestMasterKey(t, "")
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/api/spam/status", nil)
	rec := httptest.NewRecorder()

	SpamAdminStatusController(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Configured bool `json:"configured"`
		Unlocked   bool `json:"unlocked"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Configured {
		t.Fatal("expected configured=false without master key")
	}
	if response.Unlocked {
		t.Fatal("expected unlocked=false without master key")
	}
}

func TestSpamAdminSectionsRejectsMissingMasterHeader(t *testing.T) {
	restore := SetupTestMasterKey(t, "spam-master-key")
	defer restore()

	req := httptest.NewRequest(http.MethodGet, "/api/spam/sections", nil)
	rec := httptest.NewRecorder()

	SpamAdminSectionsListController(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSpamAdminSectionsSearchReturnsServerRows(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := SetupTestMasterKey(t, "spam-master-key")
	defer restore()

	if _, err := testDB.Exec(
		models.ApplyTablePrefix(`INSERT INTO quepasa_users (username, password) VALUES (?, ?)`),
		"owner@example.com",
		"hash",
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	server := &models.QpServer{Token: "spam-token-1", Verified: true}
	server.SetWId("5511999999999@s.whatsapp.net")
	server.SetUser("owner@example.com")
	server.SetContextId("context-1")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}

	body, _ := json.Marshal(spamSearchRequest{Search: "5511999999999", Limit: 10})
	req := httptest.NewRequest(http.MethodPost, "/api/spam/sections/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-MASTERKEY", "spam-master-key")
	rec := httptest.NewRecorder()

	SpamAdminSectionsSearchController(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Items []spamSectionView `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Items) != 1 {
		t.Fatalf("expected one search result, got %d", len(response.Items))
	}
	if response.Items[0].Token != "spam-token-1" {
		t.Fatalf("expected token spam-token-1, got %q", response.Items[0].Token)
	}
	if response.Items[0].InSpam {
		t.Fatal("expected search result to start outside spam queue")
	}
}

func TestSpamAdminSectionUpsertAcceptsPriority(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := SetupTestMasterKey(t, "spam-master-key")
	defer restore()

	if _, err := testDB.Exec(
		models.ApplyTablePrefix(`INSERT INTO quepasa_users (username, password) VALUES (?, ?)`),
		"owner@example.com",
		"hash",
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	server := &models.QpServer{Token: "spam-token-1", Verified: true}
	server.SetWId("5511999999999@s.whatsapp.net")
	server.SetUser("owner@example.com")
	server.SetContextId("context-1")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}

	body, _ := json.Marshal(spamSectionRequest{Token: "spam-token-1", Priority: intPtr(4), Label: "grupo verde"})
	req := httptest.NewRequest(http.MethodPost, "/api/spam/sections", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-MASTERKEY", "spam-master-key")
	rec := httptest.NewRecorder()

	SpamAdminSectionUpsertController(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Item spamSectionView `json:"item"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Item.Priority != 4 {
		t.Fatalf("expected priority 4, got %d", response.Item.Priority)
	}
	if !response.Item.Enabled {
		t.Fatal("expected enabled by default")
	}
	if response.Item.Label != "grupo verde" {
		t.Fatalf("expected label to be preserved, got %q", response.Item.Label)
	}
}

func TestSpamAdminSectionPatchPreservesPriorityWhenOmitted(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	restore := SetupTestMasterKey(t, "spam-master-key")
	defer restore()

	if _, err := testDB.Exec(
		models.ApplyTablePrefix(`INSERT INTO quepasa_users (username, password) VALUES (?, ?)`),
		"owner@example.com",
		"hash",
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	server := &models.QpServer{Token: "spam-token-1", Verified: true}
	server.SetWId("5511999999999@s.whatsapp.net")
	server.SetUser("owner@example.com")
	server.SetContextId("context-1")
	if err := models.WhatsappService.DB.Servers.Add(server); err != nil {
		t.Fatalf("add server: %v", err)
	}
	if err := models.WhatsappService.DB.SpamSections.Upsert(&models.QpSpamSection{
		Token:    "spam-token-1",
		Priority: 7,
		Enabled:  false,
		Label:    "fila azul",
	}); err != nil {
		t.Fatalf("seed spam section: %v", err)
	}

	body, _ := json.Marshal(spamSectionRequest{Token: "spam-token-1", Enabled: boolPtr(true), Label: "fila azul"})
	req := httptest.NewRequest(http.MethodPatch, "/api/spam/sections", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-QUEPASA-MASTERKEY", "spam-master-key")
	rec := httptest.NewRecorder()

	SpamAdminSectionUpsertController(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response struct {
		Item spamSectionView `json:"item"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Item.Priority != 7 {
		t.Fatalf("expected priority 7 to be preserved, got %d", response.Item.Priority)
	}
	if !response.Item.Enabled {
		t.Fatal("expected enabled to be updated")
	}
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
