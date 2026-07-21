package voip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	models "github.com/inteliagenciadigital/quepasa/models"
	runtime "github.com/inteliagenciadigital/quepasa/runtime"
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

// ModeSessionResolver resolves the authenticated WhatsApp session that owns a
// VoIP mode request.
type ModeSessionResolver func(r *http.Request) (*models.QpWhatsappServer, error)

// ErrorResponder writes an error response using the host API response contract.
type ErrorResponder func(w http.ResponseWriter, err error, code int)

// ResolveErrorResponder writes a session resolve error using the host API
// response contract.
type ResolveErrorResponder func(w http.ResponseWriter, err error)

// SuccessResponder writes a successful response using the host API response
// contract.
type SuccessResponder func(w http.ResponseWriter, response interface{})

// ModeController owns the authenticated VoIP mode endpoints.
type ModeController struct {
	ResolveServer       ModeSessionResolver
	RespondError        ErrorResponder
	RespondResolveError ResolveErrorResponder
	RespondSuccess      SuccessResponder
}

// RegisterAuthenticatedModeRoutes registers authenticated VoIP endpoints.
func RegisterAuthenticatedModeRoutes(r chi.Router, controller ModeController) {
	controller = controller.withDefaults()
	r.Get("/voip/mode", controller.GetMode)
	r.Post("/voip/mode", controller.SetMode)
}

// GetMode returns the per-instance VoIP mode.
//
//	GET /api/voip/mode?token=<token>
func (controller ModeController) GetMode(w http.ResponseWriter, r *http.Request) {
	server, ok := controller.resolveServer(w, r)
	if !ok {
		return
	}

	mode, err := runtime.GetSessionVoIPMode(server)
	if err != nil {
		controller.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	controller.RespondSuccess(w, modeResponse(mode))
}

// SetMode updates the per-instance VoIP mode and persists it to the database.
//
//	POST /api/voip/mode?token=<token>&mode=<disabled|exclusive|additional>
func (controller ModeController) SetMode(w http.ResponseWriter, r *http.Request) {
	server, ok := controller.resolveServer(w, r)
	if !ok {
		return
	}

	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	var mode whatsapp.VoIPMode
	switch raw {
	case string(whatsapp.VoIPModeDisabled), string(whatsapp.VoIPModeExclusive), string(whatsapp.VoIPModeAdditional):
		mode = whatsapp.VoIPMode(raw)
	default:
		controller.RespondError(w, fmt.Errorf("invalid mode %q; expected disabled, exclusive or additional", raw), http.StatusBadRequest)
		return
	}

	if err := runtime.SetSessionVoIPMode(server, mode); err != nil {
		controller.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	controller.RespondSuccess(w, modeResponse(mode))
}

func (controller ModeController) resolveServer(w http.ResponseWriter, r *http.Request) (*models.QpWhatsappServer, bool) {
	if controller.ResolveServer == nil {
		controller.RespondError(w, fmt.Errorf("voip mode server resolver is not configured"), http.StatusInternalServerError)
		return nil, false
	}

	server, err := controller.ResolveServer(r)
	if err != nil {
		controller.RespondResolveError(w, err)
		return nil, false
	}

	return server, true
}

func (controller ModeController) withDefaults() ModeController {
	if controller.RespondError == nil {
		controller.RespondError = defaultErrorResponder
	}
	if controller.RespondResolveError == nil {
		controller.RespondResolveError = func(w http.ResponseWriter, err error) {
			defaultErrorResponder(w, err, http.StatusNotFound)
		}
	}
	if controller.RespondSuccess == nil {
		controller.RespondSuccess = defaultSuccessResponder
	}
	return controller
}

func modeResponse(mode whatsapp.VoIPMode) map[string]interface{} {
	return map[string]interface{}{
		"mode": mode.String(),
		"options": []string{
			string(whatsapp.VoIPModeDisabled),
			string(whatsapp.VoIPModeExclusive),
			string(whatsapp.VoIPModeAdditional),
		},
	}
}

func defaultErrorResponder(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"result": err.Error()})
}

func defaultSuccessResponder(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
