package api

import models "github.com/inteliagenciadigital/quepasa/models"

// InviteResponse is the API transport shape for invite-link endpoints.
type InviteResponse struct {
	models.QpResponse
	Url string `json:"url,omitempty"`
}
