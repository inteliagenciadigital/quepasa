package models

import (
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

// Payload to include extra content for RabbitMQ
type QpRabbitMQPayload struct {
	*whatsapp.WhatsappMessage
	Extra interface{} `json:"extra,omitempty"` // extra info to append on payload
}
