package api

import models "github.com/inteliagenciadigital/quepasa/models"

// IsOnWhatsAppResponse is the API transport shape for registration-check endpoints.
type IsOnWhatsAppResponse struct {
	models.QpResponse
	Total      int      `json:"total"`
	Registered []string `json:"registered,omitempty"`
}
