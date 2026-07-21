package api

import (
	models "github.com/inteliagenciadigital/quepasa/models"
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

// MessageResponse is the API transport shape for single-message reads.
type MessageResponse struct {
	models.QpResponse
	Message *whatsapp.WhatsappMessage `json:"message,omitempty"`
}
