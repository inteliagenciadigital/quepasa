package models

import media "github.com/inteliagenciadigital/quepasa/media"

type QpToWhatsappAttachment = media.QpToWhatsappAttachment

func IsValidExtensionFor(request string, content string) bool {
	return media.IsValidExtensionFor(request, content)
}

func IsCompatibleWithPTT(mime string) bool {
	return media.IsCompatibleWithPTT(mime)
}
