package api

import (
	models "github.com/inteliagenciadigital/quepasa/models"
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

// PictureResponse is the API transport shape for profile-picture endpoints.
type PictureResponse struct {
	models.QpResponse
	Info *whatsapp.WhatsappProfilePicture `json:"info,omitempty"`
}
