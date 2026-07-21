package api

import (
	models "github.com/inteliagenciadigital/quepasa/models"
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

// ContactsResponse is the API transport shape for contact list endpoints.
type ContactsResponse struct {
	models.QpResponse
	Total    int                     `json:"total"`
	Contacts []whatsapp.WhatsappChat `json:"contacts,omitempty"`
}
