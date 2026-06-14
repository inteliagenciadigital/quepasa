package whatsmeow

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	logrus "github.com/sirupsen/logrus"
	waBinary "go.mau.fi/whatsmeow/binary"
	types "go.mau.fi/whatsmeow/types"
	events "go.mau.fi/whatsmeow/types/events"
)

// DumpCallNodeToFixture saves the XML representation of a signaling node to the fixtures directory
func DumpCallNodeToFixture(logentry *logrus.Entry, callID string, eventType string, node *waBinary.Node) {
	if node == nil {
		return
	}

	// Define stable fixtures path relative to the workspace root E:\quepasa\quepasa
	// Or we can write to /mnt/e/quepasa/quepasa/extra/voip-fixtures/ locally on Windows/WSL
	fixturesDir := filepath.Join("..", "extra", "voip-fixtures")
	
	// Fallback to absolute or sibling layout if we cannot find it
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		// Try root dir from src path
		fixturesDir = "extra/voip-fixtures"
		if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
			_ = os.MkdirAll(fixturesDir, 0755)
		}
	}

	filename := fmt.Sprintf("call_%s_%s_%d.xml", callID, eventType, time.Now().Unix())
	filePath := filepath.Join(fixturesDir, filename)

	xmlStr := node.XMLString()
	err := os.WriteFile(filePath, []byte(xmlStr), 0644)
	if err != nil {
		logentry.Errorf("failed to write signaling fixture: %v", err)
	} else {
		logentry.Infof("saved call signaling node dump: %s", filePath)
	}
}

// LogAndTrackInboundCall handles structured logging and session initiation for inbound call offer
func LogAndTrackInboundCall(logentry *logrus.Entry, evt *events.CallOffer, debugEnabled bool) {
	logentry.Infof("[VoIP Inbound Offer] CallID: %s | From: %s | Creator: %s | Platform: %s | Version: %s",
		evt.CallID, evt.From, evt.CallCreator, evt.RemotePlatform, evt.RemoteVersion)

	// Register in session manager
	session := &VoIPCallSession{
		CallID:         evt.CallID,
		Peer:           evt.From,
		State:          "Offered",
		RemotePlatform: evt.RemotePlatform,
		RemoteVersion:  evt.RemoteVersion,
		OfferTime:      time.Now(),
		Keys:           make(map[string][]byte),
	}
	GlobalVoIPSessionManager.AddSession(session)

	if debugEnabled {
		DumpCallNodeToFixture(logentry, evt.CallID, "offer", evt.Data)
	}
}

// LogAndTrackCallEvent logs and updates the session for general call negotiation events
func LogAndTrackCallEvent(logentry *logrus.Entry, eventType string, basicMeta types.BasicCallMeta, dataNode *waBinary.Node, debugEnabled bool) {
	logentry.Infof("[VoIP Event: %s] CallID: %s | Peer: %s", eventType, basicMeta.CallID, basicMeta.From)

	// Update state in manager if session exists
	if _, exists := GlobalVoIPSessionManager.GetSession(basicMeta.CallID); exists {
		GlobalVoIPSessionManager.UpdateState(basicMeta.CallID, eventType)
	}

	if debugEnabled && dataNode != nil {
		DumpCallNodeToFixture(logentry, basicMeta.CallID, eventType, dataNode)
	}
}
