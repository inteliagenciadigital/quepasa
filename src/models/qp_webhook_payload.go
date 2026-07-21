package models

import (
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

// Payload to include extra content
type QpWebhookPayload struct {
	*whatsapp.WhatsappMessage
	Extra interface{} `db:"extra" json:"extra,omitempty"` // extra info to append on payload
}
