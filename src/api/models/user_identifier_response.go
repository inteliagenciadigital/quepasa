package api

import (
	models "github.com/inteliagenciadigital/quepasa/models"
)

type UserIdentifierResponse struct {
	models.QpResponse
	UserIdentifierRequest
}
