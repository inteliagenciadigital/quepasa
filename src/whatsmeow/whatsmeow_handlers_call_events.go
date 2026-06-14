package whatsmeow

import (
	"context"

	events "go.mau.fi/whatsmeow/types/events"
)

// HandleCallOffer processes an incoming call invitation from a caller.
func (source *WhatsmeowHandlers) HandleCallOffer(evt *events.CallOffer) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	// 1. Log and track in the VoIP Session Manager
	LogAndTrackInboundCall(logentry, evt, debugEnabled)

	// 2. Extract caller's public key (if present under <encopt><key>...</key></encopt>)
	var remotePublicKey []byte
	if evt.Data != nil {
		encopt, ok := evt.Data.GetOptionalChildByTag("encopt")
		if ok {
			keyNode, ok := encopt.GetOptionalChildByTag("key")
			if ok {
				if bytes, ok := keyNode.Content.([]byte); ok {
					remotePublicKey = bytes
					logentry.Infof("[VoIP Inbound Offer] extracted remote X25519 public key (len: %d)", len(remotePublicKey))
				}
			}
		}
	}

	// 3. Trigger auto-answer if enabled and call handling is allowed
	if source.HandleCalls() && source.WhatsmeowOptions.VoIPAutoAnswer {
		logentry.Infof("[VoIP Inbound Offer] auto-answering call: %s", evt.CallID)
		go func() {
			_, err := AcceptCallNatively(context.Background(), source.Client, evt.From, evt.CallID, remotePublicKey, logentry)
			if err != nil {
				logentry.Errorf("[VoIP Inbound Offer] failed to auto-answer call: %v", err)
			}
		}()
	} else if !source.HandleCalls() {
		// Reject call natively
		err := source.Client.RejectCall(context.Background(), evt.From, evt.CallID)
		if err != nil {
			logentry.Errorf("error on rejecting call: %s", err.Error())
		} else {
			logentry.Infof("rejecting incoming call from: %s", evt.From)
		}
	}

	// 4. Fallback: propagate to internal message handlers for UI/event reporting
	go source.CallMessage(evt.BasicCallMeta)
}

// HandleCallOfferNotice processes notifications about call offers.
func (source *WhatsmeowHandlers) HandleCallOfferNotice(evt *events.CallOfferNotice) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	LogAndTrackCallEvent(logentry, "OfferNotice", evt.BasicCallMeta, evt.Data, debugEnabled)
	go source.CallMessage(evt.BasicCallMeta)
}

// HandleCallPreAccept processes the pre-accept negotiation signal.
func (source *WhatsmeowHandlers) HandleCallPreAccept(evt *events.CallPreAccept) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	LogAndTrackCallEvent(logentry, "PreAccept", evt.BasicCallMeta, evt.Data, debugEnabled)
}

// HandleCallAccept processes the accept negotiation signal (sent by peer or other client instance).
func (source *WhatsmeowHandlers) HandleCallAccept(evt *events.CallAccept) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	LogAndTrackCallEvent(logentry, "Accept", evt.BasicCallMeta, evt.Data, debugEnabled)
}

// HandleCallReject processes the reject signal sent when a call is declined.
func (source *WhatsmeowHandlers) HandleCallReject(evt *events.CallReject) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	LogAndTrackCallEvent(logentry, "Reject", evt.BasicCallMeta, evt.Data, debugEnabled)
}

// HandleCallTransport processes the transport negotiation node containing media candidates (TURN/relays).
func (source *WhatsmeowHandlers) HandleCallTransport(evt *events.CallTransport) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	LogAndTrackCallEvent(logentry, "Transport", evt.BasicCallMeta, evt.Data, debugEnabled)
}

// HandleCallTerminate processes the terminate signal when the call is ended.
func (source *WhatsmeowHandlers) HandleCallTerminate(evt *events.CallTerminate) {
	logentry := source.GetLogger()
	debugEnabled := source.WhatsmeowOptions.VoIPMediaDebug

	LogAndTrackCallEvent(logentry, "Terminate", evt.BasicCallMeta, evt.Data, debugEnabled)

	// Clean up call session
	GlobalVoIPSessionManager.RemoveSession(evt.CallID)
}
