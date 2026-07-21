package voip

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	sipproxy "github.com/inteliagenciadigital/quepasa/sipproxy"
	calls "github.com/inteliagenciadigital/quepasa/voip/calls"
)

var outboundManagers sync.Map // map[string]*VoipManager, keyed by section id and session token

func registerOutboundManager(m *VoipManager) {
	if m == nil {
		return
	}
	for _, key := range outboundManagerKeys(m) {
		outboundManagers.Store(key, m)
	}
}

func unregisterOutboundManager(m *VoipManager) {
	if m == nil {
		return
	}
	for _, key := range outboundManagerKeys(m) {
		if current, ok := outboundManagers.Load(key); ok && current == m {
			outboundManagers.Delete(key)
		}
	}
}

func outboundManagerKeys(m *VoipManager) []string {
	keys := make([]string, 0, 2)
	if value := strings.TrimSpace(m.sectionID); value != "" {
		keys = append(keys, value)
	}
	if value := strings.TrimSpace(m.sessionToken); value != "" {
		keys = append(keys, value)
	}
	return keys
}

// HandleOutboundSIPInvite routes a SIP-originated call into the WhatsApp
// section identified by req.SessionID.
func HandleOutboundSIPInvite(ctx context.Context, req sipproxy.OutboundWhatsAppInviteRequest) (sipproxy.OutboundWhatsAppInviteResponse, error) {
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("missing WhatsApp session id")
	}
	value, ok := outboundManagers.Load(sessionID)
	if !ok {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("WhatsApp session %s is not available", sessionID)
	}
	mgr, ok := value.(*VoipManager)
	if !ok || mgr == nil {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("WhatsApp session %s has invalid manager", sessionID)
	}
	return mgr.placeSIPOriginatedCall(ctx, req)
}

func (m *VoipManager) placeSIPOriginatedCall(ctx context.Context, req sipproxy.OutboundWhatsAppInviteRequest) (sipproxy.OutboundWhatsAppInviteResponse, error) {
	if m == nil || m.proxy == nil {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("voip manager: sipproxy not configured")
	}
	if strings.TrimSpace(req.CallID) == "" {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("missing SIP Call-ID")
	}
	if strings.TrimSpace(req.Target) == "" {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("missing WhatsApp target")
	}
	if strings.TrimSpace(req.RemoteRTPHost) == "" || req.RemoteRTPPort <= 0 {
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("missing remote RTP address")
	}

	callID := strings.TrimSpace(req.CallID)
	target := strings.TrimLeft(strings.TrimSpace(req.Target), "+")
	accountPhone := firstNonEmpty(m.getOwnPhone(), strings.Split(strings.TrimSpace(req.SessionID), ":")[0])

	stream, err := m.proxy.BridgeOutboundSIPCall(callID, accountPhone, target, req.RemoteRTPHost, req.RemoteRTPPort)
	if err != nil {
		return sipproxy.OutboundWhatsAppInviteResponse{}, err
	}
	sipConn := stream.WhatsAppConn
	if sipConn == nil {
		closeRTPStream(stream)
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("outbound SIP RTP connection is nil")
	}
	m.proxy.SetLocalRTPPort(callID, stream.WhatsAppPort)
	answerSDP, err := m.proxy.CreateSDPAnswer(callID, accountPhone)
	if err != nil {
		closeRTPStream(stream)
		return sipproxy.OutboundWhatsAppInviteResponse{}, err
	}

	call, err := m.PlaceCall(ctx, target)
	if err != nil {
		closeRTPStream(stream)
		return sipproxy.OutboundWhatsAppInviteResponse{}, err
	}

	ready := make(chan struct{}, 1)
	ended := make(chan string, 1)
	peer := &sipPeer{}
	if addr, err := net.ResolveUDPAddr("udp", req.RemoteRTPAddr); err == nil {
		peer.set(addr)
	}

	call.OnReady(func() {
		bridgeSource := NewVoipBridgeSource(sipConn, peer)
		bridgeSource.SetLogger(m.log)
		bridgeSource.StartReadLoop()

		player := calls.NewPlayer()
		call.Subscribe(player)
		player.Play(bridgeSource)

		bridgeSink := NewVoipBridgeSink(sipConn, peer, 0)
		call.Receive(bridgeSink)

		ctx := &bridgeContext{
			callID:     callID,
			sinkConn:   sipConn,
			sourceConn: sipConn,
			source:     bridgeSource,
			sink:       bridgeSink,
		}
		m.activeCalls.Store(callID, ctx)
		m.registerSIPRemoteTermination(call, callID)

		m.log.InfoE().
			Str("call_id", callID).
			Str("whatsapp_call_id", call.ID()).
			Str("target", target).
			Str("sip_rtp", req.RemoteRTPAddr).
			Msg("VoipManager: SIP-originated WhatsApp bridge ready")

		select {
		case ready <- struct{}{}:
		default:
		}
	})

	call.OnEnd(func(reason string) {
		m.log.InfoE().
			Str("call_id", callID).
			Str("whatsapp_call_id", call.ID()).
			Str("reason", reason).
			Msg("VoipManager: SIP-originated WhatsApp call ended")

		select {
		case ended <- reason:
		default:
		}
		m.cleanupCall(callID)
	})

	go func() {
		<-ctx.Done()
		_ = call.Hangup()
	}()

	timeout := 60 * time.Second
	select {
	case <-ready:
		return sipproxy.OutboundWhatsAppInviteResponse{SDP: answerSDP}, nil
	case reason := <-ended:
		closeRTPStream(stream)
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("WhatsApp call ended before media ready: %s", reason)
	case <-time.After(timeout):
		_ = call.Hangup()
		closeRTPStream(stream)
		return sipproxy.OutboundWhatsAppInviteResponse{}, fmt.Errorf("WhatsApp target did not answer within %s", timeout)
	case <-ctx.Done():
		_ = call.Hangup()
		closeRTPStream(stream)
		return sipproxy.OutboundWhatsAppInviteResponse{}, ctx.Err()
	}
}

func closeRTPStream(stream *sipproxy.RTPStream) {
	if stream == nil {
		return
	}
	if stream.WhatsAppConn != nil {
		_ = stream.WhatsAppConn.Close()
	}
	if stream.SIPConn != nil && stream.SIPConn != stream.WhatsAppConn {
		_ = stream.SIPConn.Close()
	}
}
