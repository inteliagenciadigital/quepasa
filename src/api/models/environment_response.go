package api

import (
	environment "github.com/inteliagenciadigital/quepasa/environment"
	models "github.com/inteliagenciadigital/quepasa/models"
)

// EnvironmentResponse represents environment settings response
type EnvironmentResponse struct {
	models.QpResponse
	Settings            *environment.EnvironmentSettings        `json:"settings,omitempty"`
	Preview             *environment.EnvironmentSettingsPreview `json:"preview,omitempty"`
	MasterKeyConfigured *bool                                   `json:"masterKeyConfigured,omitempty"`
}
