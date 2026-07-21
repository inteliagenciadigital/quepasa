package whatsapp

import log "github.com/inteliagenciadigital/quepasa/qplog"

type IWhatsappConnectionOptions interface {
	GetLogger() log.Logger
}
