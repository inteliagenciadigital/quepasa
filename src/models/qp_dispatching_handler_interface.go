package models

import whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"

type QpDispatchingHandlerInterface interface {

	// method for init process of dispatching messages
	HandleDispatching(*whatsapp.WhatsappMessage)
}
