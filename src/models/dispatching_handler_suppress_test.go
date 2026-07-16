package models

import (
	"testing"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

func TestShouldSuppressRevoke(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.RevokeMessageType,
	}
	if !shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected REVOKE to be suppressed from create webhook")
	}
}

func TestShouldSuppressEdit(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:     "3EB09807FC425F4804388E",
		Type:   whatsapp.TextMessageType,
		Edited: true,
		Text:   "edited text",
	}
	if !shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected EDIT to be suppressed from create webhook")
	}
}

func TestShouldSuppressReaction(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:         "3EB09807FC425F4804388E",
		Type:       whatsapp.TextMessageType,
		InReaction: true,
		Text:       "👍",
	}
	if !shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected REACTION to be suppressed from create webhook")
	}
}

func TestShouldNotSuppressNormalText(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.TextMessageType,
		Text: "hello world",
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal text message to NOT be suppressed")
	}
}

func TestShouldNotSuppressNormalImage(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.ImageMessageType,
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal image message to NOT be suppressed")
	}
}

func TestShouldNotSuppressNormalAudio(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.AudioMessageType,
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal audio message to NOT be suppressed")
	}
}

func TestShouldNotSuppressNormalDocument(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.DocumentMessageType,
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal document message to NOT be suppressed")
	}
}

func TestShouldNotSuppressNormalVideo(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.VideoMessageType,
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal video message to NOT be suppressed")
	}
}

func TestShouldNotSuppressNormalLocation(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.LocationMessageType,
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal location message to NOT be suppressed")
	}
}

func TestShouldNotSuppressNormalContact(t *testing.T) {
	msg := &whatsapp.WhatsappMessage{
		Id:   "3EB09807FC425F4804388E",
		Type: whatsapp.ContactMessageType,
	}
	if shouldSuppressFromCreateWebhook(msg) {
		t.Fatal("expected normal contact message to NOT be suppressed")
	}
}

func TestShouldSuppressNilMessage(t *testing.T) {
	if !shouldSuppressFromCreateWebhook(nil) {
		t.Fatal("expected nil message to be suppressed")
	}
}
