package whatsmeow

import (
	"bytes"
	"testing"
	"time"

	golangCurve "golang.org/x/crypto/curve25519"
	types "go.mau.fi/whatsmeow/types"
)

func TestVoIPCallSessionManager_Lifecycle(t *testing.T) {
	manager := NewVoIPCallSessionManager()

	callID := "test-call-12345"
	peerJID, _ := types.ParseJID("5511999999999@s.whatsapp.net")

	session := &VoIPCallSession{
		CallID:    callID,
		Peer:      peerJID,
		State:     "Offered",
		OfferTime: time.Now(),
		Keys:      make(map[string][]byte),
	}

	// 1. Add session
	manager.AddSession(session)

	// 2. Retrieve session
	retrieved, exists := manager.GetSession(callID)
	if !exists {
		t.Fatalf("expected session to exist")
	}
	if retrieved.CallID != callID {
		t.Errorf("expected CallID %s, got %s", callID, retrieved.CallID)
	}
	if retrieved.State != "Offered" {
		t.Errorf("expected State Offered, got %s", retrieved.State)
	}

	// 3. Update state
	manager.UpdateState(callID, "Accepted")
	retrieved, _ = manager.GetSession(callID)
	if retrieved.State != "Accepted" {
		t.Errorf("expected updated state Accepted, got %s", retrieved.State)
	}
	if retrieved.AnswerTime.IsZero() {
		t.Errorf("expected AnswerTime to be set")
	}

	// 4. List sessions
	list := manager.ListSessions()
	if len(list) != 1 {
		t.Errorf("expected 1 session in list, got %d", len(list))
	}

	// 5. Remove session
	manager.RemoveSession(callID)
	_, exists = manager.GetSession(callID)
	if exists {
		t.Errorf("expected session to be removed")
	}
}

func TestX25519KeyGeneration_AndDerivation(t *testing.T) {
	// Generate Alice's keys
	alicePriv, alicePub, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("failed to generate Alice keypair: %v", err)
	}
	if len(alicePriv) != 32 || len(alicePub) != 32 {
		t.Errorf("keys must be 32 bytes. priv: %d, pub: %d", len(alicePriv), len(alicePub))
	}

	// Generate Bob's keys
	bobPriv, bobPub, err := GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("failed to generate Bob keypair: %v", err)
	}

	// Compute shared secret from Alice's side (Alice priv + Bob pub)
	aliceSecret, err := golangCurve.X25519(alicePriv, bobPub)
	if err != nil {
		t.Fatalf("failed to calculate Alice shared secret: %v", err)
	}

	// Compute shared secret from Bob's side (Bob priv + Alice pub)
	bobSecret, err := golangCurve.X25519(bobPriv, alicePub)
	if err != nil {
		t.Fatalf("failed to calculate Bob shared secret: %v", err)
	}

	// Verify both derived secrets match
	if !bytes.Equal(aliceSecret, bobSecret) {
		t.Errorf("ECDH mismatch: Alice secret %x != Bob secret %x", aliceSecret, bobSecret)
	}
}
