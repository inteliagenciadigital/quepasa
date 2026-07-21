package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/inteliagenciadigital/quepasa/library"
	"github.com/inteliagenciadigital/quepasa/models"
	"github.com/inteliagenciadigital/quepasa/runtime"
)

type contextAccessRequest struct {
	Username  string `json:"username"`
	ContextID string `json:"contextid"`
	Label     string `json:"label"`
	Enabled   *bool  `json:"enabled"`
}

type contextSessionSearchRequest struct {
	Search    string `json:"search"`
	ContextID string `json:"contextid"`
	Limit     int    `json:"limit"`
}

func RegisterContextAccessControllers(r chi.Router) {
	r.Get("/context-access/status", ContextAccessStatusController)
	r.Get("/context-access", ContextAccessListController)
	r.Post("/context-access", ContextAccessUpsertController)
	r.Patch("/context-access", ContextAccessUpsertController)
	r.Delete("/context-access", ContextAccessDeleteController)
	r.Post("/context-access/sessions/search", ContextAccessSessionsSearchController)
}

func ContextAccessStatusController(w http.ResponseWriter, r *http.Request) {
	configured := isMasterKeyConfigured(models.ENV.MasterKey())
	RespondSuccess(w, map[string]interface{}{
		"configured": configured,
		"unlocked":   configured && IsMatchForMaster(r),
	})
}

func ContextAccessListController(w http.ResponseWriter, r *http.Request) {
	if !requireContextAccessMasterKey(w, r) {
		return
	}

	items, err := models.GetDatabase().UserContexts.ListAll()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	username := strings.TrimSpace(library.GetRequestParameter(r, "username"))
	contextID := strings.TrimSpace(library.GetRequestParameter(r, "contextid"))
	filtered := make([]*models.QpUserContextAccess, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if username != "" && !strings.EqualFold(item.Username, username) {
			continue
		}
		if contextID != "" && item.ContextID != contextID {
			continue
		}
		filtered = append(filtered, item)
	}

	RespondSuccess(w, map[string]interface{}{"items": filtered})
}

func ContextAccessUpsertController(w http.ResponseWriter, r *http.Request) {
	if !requireContextAccessMasterKey(w, r) {
		return
	}

	request := contextAccessRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			RespondBadRequest(w, fmt.Errorf("invalid json body: %w", err))
			return
		}
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "username")); value != "" {
		request.Username = value
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "contextid")); value != "" {
		request.ContextID = value
	}

	request.Username = strings.TrimSpace(request.Username)
	request.ContextID = strings.TrimSpace(request.ContextID)
	request.Label = strings.TrimSpace(request.Label)
	if request.Username == "" || request.ContextID == "" {
		RespondBadRequest(w, fmt.Errorf("username and contextid are required"))
		return
	}

	enabled := true
	if existing, _ := models.GetDatabase().UserContexts.Find(request.Username, request.ContextID); existing != nil {
		enabled = existing.Enabled
	}
	if request.Enabled != nil {
		enabled = *request.Enabled
	}

	access := &models.QpUserContextAccess{
		Username:  request.Username,
		ContextID: request.ContextID,
		Label:     request.Label,
		Enabled:   enabled,
	}
	if err := models.GetDatabase().UserContexts.Upsert(access); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	saved, _ := models.GetDatabase().UserContexts.Find(access.Username, access.ContextID)
	RespondSuccess(w, map[string]interface{}{"item": saved})
}

func ContextAccessDeleteController(w http.ResponseWriter, r *http.Request) {
	if !requireContextAccessMasterKey(w, r) {
		return
	}

	request := contextAccessRequest{}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "username")); value != "" {
		request.Username = value
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "contextid")); value != "" {
		request.ContextID = value
	}
	request.Username = strings.TrimSpace(request.Username)
	request.ContextID = strings.TrimSpace(request.ContextID)
	if request.Username == "" || request.ContextID == "" {
		RespondBadRequest(w, fmt.Errorf("username and contextid are required"))
		return
	}

	removed, err := models.GetDatabase().UserContexts.Delete(request.Username, request.ContextID)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]interface{}{"removed": removed})
}

func ContextAccessSessionsSearchController(w http.ResponseWriter, r *http.Request) {
	if !requireContextAccessMasterKey(w, r) {
		return
	}

	request := contextSessionSearchRequest{Limit: 100}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "search")); value != "" {
		request.Search = value
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "contextid")); value != "" {
		request.ContextID = value
	}
	if request.Limit <= 0 || request.Limit > 250 {
		request.Limit = 100
	}

	items := searchContextSessionSummaries(strings.TrimSpace(request.Search), strings.TrimSpace(request.ContextID), request.Limit)
	RespondSuccess(w, map[string]interface{}{"items": items})
}

func AuthenticatedContextAccessListController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	items, err := models.GetDatabase().UserContexts.ListForUser(user.Username, true)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]interface{}{"items": items})
}

func AuthenticatedContextAccessUpsertController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	request := contextAccessRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			RespondBadRequest(w, fmt.Errorf("invalid json body: %w", err))
			return
		}
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "contextid")); value != "" {
		request.ContextID = value
	}

	contextID := strings.TrimSpace(request.ContextID)
	if contextID == "" {
		RespondBadRequest(w, fmt.Errorf("contextid is required"))
		return
	}

	enabled := true
	if existing, _ := models.GetDatabase().UserContexts.Find(user.Username, contextID); existing != nil {
		enabled = existing.Enabled
	}
	if request.Enabled != nil {
		enabled = *request.Enabled
	}

	access := &models.QpUserContextAccess{
		Username:  user.Username,
		ContextID: contextID,
		Label:     strings.TrimSpace(request.Label),
		Enabled:   enabled,
	}
	if err := models.GetDatabase().UserContexts.Upsert(access); err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	saved, _ := models.GetDatabase().UserContexts.Find(access.Username, access.ContextID)
	RespondSuccess(w, map[string]interface{}{"item": saved})
}

func AuthenticatedContextAccessDeleteController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	request := contextAccessRequest{}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	if value := strings.TrimSpace(library.GetRequestParameter(r, "contextid")); value != "" {
		request.ContextID = value
	}

	contextID := strings.TrimSpace(request.ContextID)
	if contextID == "" {
		RespondBadRequest(w, fmt.Errorf("contextid is required"))
		return
	}

	deletedSessions, err := deleteOwnedContextSessions(user.Username, contextID)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	removed, err := models.GetDatabase().UserContexts.Delete(user.Username, contextID)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	RespondSuccess(w, map[string]interface{}{
		"removed":          removed || deletedSessions > 0,
		"removed_access":   removed,
		"deleted_sessions": deletedSessions,
	})
}

func AuthenticatedContextSessionsController(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	allowed, err := models.GetDatabase().UserContexts.ListForUser(user.Username, true)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	allowedMap := map[string]bool{}
	for _, item := range allowed {
		if item != nil && strings.TrimSpace(item.ContextID) != "" {
			allowedMap[item.ContextID] = true
		}
	}

	requested := parseContextIDs(r)
	if len(requested) == 0 {
		requested = make([]string, 0, len(allowedMap))
		for contextID := range allowedMap {
			requested = append(requested, contextID)
		}
	}

	contexts := map[string]bool{}
	for _, contextID := range requested {
		if !allowedMap[contextID] {
			RespondErrorCode(w, fmt.Errorf("context not authorized"), http.StatusForbidden)
			return
		}
		contexts[contextID] = true
	}

	items := listContextSessionSummaries(contexts)
	RespondSuccess(w, map[string]interface{}{
		"items":    items,
		"servers":  items,
		"total":    len(items),
		"username": user.Username,
	})
}

func requireContextAccessMasterKey(w http.ResponseWriter, r *http.Request) bool {
	if !isMasterKeyConfigured(models.ENV.MasterKey()) {
		RespondErrorCode(w, fmt.Errorf("master key is not configured"), http.StatusForbidden)
		return false
	}
	if !IsMatchForMaster(r) {
		RespondErrorCode(w, errSpamMasterKeyRequired, http.StatusUnauthorized)
		return false
	}
	return true
}

func parseContextIDs(r *http.Request) []string {
	values := append([]string{}, r.URL.Query()["contextid"]...)
	values = append(values, r.URL.Query()["contextids"]...)

	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			contextID := strings.TrimSpace(part)
			if contextID == "" || seen[contextID] {
				continue
			}
			seen[contextID] = true
			result = append(result, contextID)
		}
	}
	return result
}

func listContextSessionSummaries(contexts map[string]bool) []map[string]interface{} {
	if len(contexts) == 0 {
		return []map[string]interface{}{}
	}

	items := make([]map[string]interface{}, 0)
	for _, server := range models.GetDatabase().Servers.FindAll() {
		if server == nil || strings.TrimSpace(server.GetContextId()) == "" {
			continue
		}
		if !contexts[server.GetContextId()] {
			continue
		}
		items = append(items, BuildServerSummary(server, FindLiveServer(server.Token)))
	}

	sort.SliceStable(items, func(i, j int) bool {
		wi, _ := items[i]["wid"].(string)
		wj, _ := items[j]["wid"].(string)
		return wi < wj
	})
	return items
}

func deleteOwnedContextSessions(username string, contextID string) (int, error) {
	username = strings.TrimSpace(username)
	contextID = strings.TrimSpace(contextID)
	if username == "" || contextID == "" {
		return 0, nil
	}

	owned := []*models.QpServer{}
	for _, server := range models.GetDatabase().Servers.FindAll() {
		if server == nil {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(server.GetUser()), username) {
			continue
		}
		if strings.TrimSpace(server.GetContextId()) != contextID {
			continue
		}
		owned = append(owned, server)
	}

	deleted := 0
	for _, record := range owned {
		session := FindLiveServer(record.Token)
		var err error
		if session == nil {
			session, err = runtime.LoadSessionRecord(record)
			if err != nil {
				return deleted, err
			}
		}

		if err := runtime.DeleteSessionRecord(session, "authenticated-context-access"); err != nil {
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}

func searchContextSessionSummaries(query string, contextID string, limit int) []map[string]interface{} {
	query = strings.ToLower(strings.TrimSpace(query))
	contextID = strings.TrimSpace(contextID)
	items := make([]map[string]interface{}, 0)

	for _, server := range models.GetDatabase().Servers.FindAll() {
		if server == nil {
			continue
		}
		if contextID != "" && server.GetContextId() != contextID {
			continue
		}
		if query != "" && !contextSessionMatches(server, query) {
			continue
		}

		items = append(items, BuildServerSummary(server, FindLiveServer(server.Token)))
		if len(items) >= limit {
			break
		}
	}

	return items
}

func contextSessionMatches(server *models.QpServer, query string) bool {
	values := []string{
		server.Token,
		server.GetWId(),
		server.GetUser(),
		server.GetContextId(),
		whatsappNumberForContextSearch(server.GetWId()),
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

func whatsappNumberForContextSearch(wid string) string {
	value := strings.TrimSpace(wid)
	value = strings.TrimSuffix(value, "@s.whatsapp.net")
	return strings.TrimSuffix(value, "@lid")
}
