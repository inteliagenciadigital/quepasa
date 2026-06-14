package environment

// WhatsApp environment variable names
const (
	ENV_DISPATCH_UNHANDLED                = "DISPATCHUNHANDLED"                // enable or disable dispatch unhandled messages
	ENV_WHATSMEOWLOGLEVEL                 = "WHATSMEOW_LOGLEVEL"               // Whatsmeow log level
	ENV_WHATSMEOWDBLOGLEVEL               = "WHATSMEOW_DBLOGLEVEL"             // Whatsmeow database log level
	ENV_WHATSMEOW_USE_RETRY_MESSAGE_STORE = "WHATSMEOW_USE_RETRY_MESSAGE_STORE" // persist outgoing messages for retry receipts
	ENV_VOIP_AUTO_ANSWER                  = "VOIP_AUTO_ANSWER"                  // enable or disable auto answer for incoming calls
	ENV_VOIP_MEDIA_DEBUG                  = "VOIP_MEDIA_DEBUG"                  // enable or disable dumping call media / signaling nodes
)

// WhatsmeowSettings holds all WhatsApp configuration loaded from environment
type WhatsmeowSettings struct {
	DispatchUnhandled    bool   `json:"dispatch_unhandled"`
	LogLevel             string `json:"whatsmeow_log_level"`
	DBLogLevel           string `json:"whatsmeow_db_log_level"`
	UseRetryMessageStore bool   `json:"whatsmeow_use_retry_message_store"`
	VoIPAutoAnswer       bool   `json:"voip_auto_answer"`
	VoIPMediaDebug       bool   `json:"voip_media_debug"`
}

// NewWhatsmeowSettings creates a new Whatsmeow settings by loading all values from environment
func NewWhatsmeowSettings() WhatsmeowSettings {
	return WhatsmeowSettings{
		DispatchUnhandled:    getEnvOrDefaultBool(ENV_DISPATCH_UNHANDLED, false),
		LogLevel:             getEnvOrDefaultString(ENV_WHATSMEOWLOGLEVEL, ""),
		DBLogLevel:           getEnvOrDefaultString(ENV_WHATSMEOWDBLOGLEVEL, ""),
		UseRetryMessageStore: getEnvOrDefaultBool(ENV_WHATSMEOW_USE_RETRY_MESSAGE_STORE, false),
		VoIPAutoAnswer:       getEnvOrDefaultBool(ENV_VOIP_AUTO_ANSWER, false),
		VoIPMediaDebug:       getEnvOrDefaultBool(ENV_VOIP_MEDIA_DEBUG, false),
	}
}
