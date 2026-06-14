package whatsmeow

import (
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Whatsmeow service options, setted on start, so if want to changed then, you have to restart the entire service
type WhatsmeowOptions struct {
	whatsapp.WhatsappOptionsExtended

	// default whatsmeow log level
	WMLogLevel string `json:"wmloglevel,omitempty"`

	// default database log level
	DBLogLevel string `json:"dbloglevel,omitempty"`

	// persist outgoing messages in store to support retry receipts after restarts
	UseRetryMessageStore bool `json:"use_retry_message_store,omitempty"`

	// VoIP experimental settings
	VoIPAutoAnswer bool `json:"voip_auto_answer,omitempty"`
	VoIPMediaDebug bool `json:"voip_media_debug,omitempty"`
}

func (source WhatsmeowOptions) IsDefault() bool {
	return len(source.WMLogLevel) == 0 && len(source.DBLogLevel) == 0 && !source.UseRetryMessageStore && !source.VoIPAutoAnswer && !source.VoIPMediaDebug && source.WhatsappOptionsExtended.IsDefault()
}
