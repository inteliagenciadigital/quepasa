package api

import (
	models "github.com/inteliagenciadigital/quepasa/models"
	runtime "github.com/inteliagenciadigital/quepasa/runtime"
)

func findConversationLabelStore() (models.QpDataConversationLabelsInterface, error) {
	return runtime.GetConversationLabelStore()
}
