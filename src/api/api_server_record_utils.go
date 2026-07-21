package api

import (
	models "github.com/inteliagenciadigital/quepasa/models"
	runtime "github.com/inteliagenciadigital/quepasa/runtime"
)

func listPersistedServerRecords() ([]*models.QpServer, error) {
	return runtime.ListPersistedSessionRecords()
}

func findPersistedServerRecord(token string) (*models.QpServer, error) {
	return runtime.FindPersistedSessionRecord(token)
}
