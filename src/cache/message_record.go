package cache

import (
	"time"

	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
)

type MessageRecord struct {
	Message   *whatsapp.WhatsappMessage `json:"message,omitempty"`
	ExpiresAt time.Time                 `json:"expires_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}