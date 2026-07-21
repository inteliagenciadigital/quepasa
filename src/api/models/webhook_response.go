package api

import models "github.com/inteliagenciadigital/quepasa/models"

// WebhookResponse is the API transport shape for webhook configuration endpoints.
type WebhookResponse struct {
	models.QpResponse
	Affected uint                `json:"affected,omitempty"`
	Webhooks []*models.QpWebhook `json:"webhooks,omitempty"`
}
