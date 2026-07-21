package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	models "github.com/inteliagenciadigital/quepasa/models"
	runtime "github.com/inteliagenciadigital/quepasa/runtime"
	"github.com/inteliagenciadigital/quepasa/voip"
)

func registerAuthenticatedVoIPRoutes(r chi.Router) {
	voip.RegisterAuthenticatedModeRoutes(r, voip.ModeController{
		ResolveServer:       resolveAuthenticatedVoIPModeServer,
		RespondError:        RespondErrorCode,
		RespondResolveError: respondAuthenticatedVoIPModeResolveError,
		RespondSuccess:      RespondSuccess,
	})
}

// resolveAuthenticatedVoIPModeServer resolves the authenticated instance from
// the "token" query-string parameter shared by the VoIP mode controller. Access
// is allowed to the owner or to users explicitly linked to the session context.
func resolveAuthenticatedVoIPModeServer(r *http.Request) (*models.QpWhatsappServer, error) {
	user, err := GetAuthenticatedUser(r)
	if err != nil {
		return nil, &authenticatedVoIPModeResolveError{err: err, code: http.StatusUnauthorized}
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		return nil, &authenticatedVoIPModeResolveError{err: fmt.Errorf("missing token parameter"), code: http.StatusBadRequest}
	}

	return runtime.GetOwnedOrContextLiveServer(user, token)
}

func respondAuthenticatedVoIPModeResolveError(w http.ResponseWriter, err error) {
	var resolveErr *authenticatedVoIPModeResolveError
	if errors.As(err, &resolveErr) {
		RespondErrorCode(w, resolveErr.err, resolveErr.code)
		return
	}

	respondServerLookupError(w, err)
}

type authenticatedVoIPModeResolveError struct {
	err  error
	code int
}

func (err *authenticatedVoIPModeResolveError) Error() string {
	return err.err.Error()
}

func (err *authenticatedVoIPModeResolveError) Unwrap() error {
	return err.err
}
