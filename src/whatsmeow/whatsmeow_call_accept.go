package whatsmeow

import (
	"context"
	"crypto/rand"
	"fmt"

	logrus "github.com/sirupsen/logrus"
	golangCurve "golang.org/x/crypto/curve25519"
	whatsmeow "go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	types "go.mau.fi/whatsmeow/types"
)

// GenerateX25519KeyPair produces a fresh Curve25519 private/public keypair
func GenerateX25519KeyPair() (privateKey, publicKey []byte, err error) {
	privateKey = make([]byte, 32)
	_, err = rand.Read(privateKey)
	if err != nil {
		return nil, nil, err
	}
	publicKey, err = golangCurve.X25519(privateKey, golangCurve.Basepoint)
	return privateKey, publicKey, err
}

// AcceptCallNatively performs curve25519 negotiation, constructs the accept stanza and transmits it
func AcceptCallNatively(ctx context.Context, client *whatsmeow.Client, from types.JID, callID string, remotePublicKey []byte, logentry *logrus.Entry) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("nil whatsmeow client")
	}

	ownID := client.Store.ID.ToNonAD()
	peer := from.ToNonAD()

	// 1. Generate local key pair for ECDH key exchange
	privKey, pubKey, err := GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate X25519 keypair: %v", err)
	}

	// 2. Perform ECDH if remote key is provided
	var sharedSecret []byte
	if len(remotePublicKey) == 32 {
		sharedSecret, err = golangCurve.X25519(privKey, remotePublicKey)
		if err != nil {
			logentry.Warnf("failed to derive shared secret: %v", err)
		} else {
			logentry.Infof("[VoIP Accept] derived ECDH shared secret (first 8 bytes: %x)", sharedSecret[:8])
			// Store in call session if tracked
			if session, exists := GlobalVoIPSessionManager.GetSession(callID); exists {
				session.Keys["shared_secret"] = sharedSecret
			}
		}
	} else {
		logentry.Warn("[VoIP Accept] remote public key was missing or invalid length, skipping ECDH")
	}

	// 3. Assemble accept child nodes
	audio16k := waBinary.Node{Tag: "audio", Attrs: waBinary.Attrs{"enc": "opus", "rate": "16000"}}
	audio8k := waBinary.Node{Tag: "audio", Attrs: waBinary.Attrs{"enc": "opus", "rate": "8000"}}
	netNode := waBinary.Node{Tag: "net", Attrs: waBinary.Attrs{"medium": "3"}}

	keyNode := waBinary.Node{Tag: "key", Content: pubKey}
	encopt := waBinary.Node{
		Tag: "encopt",
		Attrs: waBinary.Attrs{
			"keygen": "2",
		},
		Content: []waBinary.Node{keyNode},
	}

	acceptContent := []waBinary.Node{audio16k, audio8k, netNode, encopt}

	// 4. Assemble main call stanza
	node := waBinary.Node{
		Tag: "call",
		Attrs: waBinary.Attrs{
			"id":   client.GenerateMessageID(),
			"to":   peer,
			"from": ownID,
		},
		Content: []waBinary.Node{
			{
				Tag: "accept",
				Attrs: waBinary.Attrs{
					"call-id":      callID,
					"call-creator": peer,
				},
				Content: acceptContent,
			},
		},
	}

	logentry.Infof("[VoIP Accept] Sending accept node to: %s for call: %s", peer, callID)
	if logentry.Logger.GetLevel() >= logrus.DebugLevel {
		logentry.Debugf("Accept node payload:\n%s", node.XMLString())
	}

	// 5. Send node natively
	err = client.DangerousInternals().SendNode(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("failed to send accept node: %v", err)
	}

	logentry.Infof("[VoIP Accept] Successfully transmitted accept node for call: %s", callID)

	return pubKey, nil
}
