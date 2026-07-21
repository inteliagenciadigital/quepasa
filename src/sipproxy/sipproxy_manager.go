package sipproxy

import (
	"fmt"
	"sync"

	qplog "github.com/inteliagenciadigital/quepasa/qplog"
)

// SIPProxyManager coordena os mÃ³dulos refatorados usando sipgo
type SIPProxyManager struct {
	mutex     sync.RWMutex
	logger    qplog.Logger
	config    SIPProxySettings
	isRunning bool

	// MÃ³dulos refatorados
	networkManager     *SIPProxyNetworkManager
	responseHandler    *SIPResponseHandler
	callManagerSipgo   *SIPCallManagerSipgo // sipgo-based call manager
	transactionMonitor *SIPTransactionMonitor

	// Componentes legados (mantidos para compatibilidade)
	upnpManager *UPnPManager
	sipListener *SIPListener

	// Rastreamento de chamadas
	activeCalls  map[string]*SIPProxyCallData
	callAttempts map[string]int
}

var (
	managerInstance *SIPProxyManager
	managerOnce     sync.Once
)

// GetSIPProxyManager retorna a instÃ¢ncia singleton do manager refatorado
func GetSIPProxyManager(settings SIPProxySettings) *SIPProxyManager {
	managerOnce.Do(func() {
		logentry := qplog.New().WithField("package", "sipproxy")

		// Inicializar componentes legados
		upnpManager := NewUPnPManager(logentry)
		sipListener := NewSIPListener(logentry)
		transactionMonitor := NewSIPTransactionMonitor(logentry)

		// Inicializar mÃ³dulos refatorados
		networkManager := NewSIPProxyNetworkManager(settings.SIPProxyNetworkManagerSettings, logentry)

		responseHandler := NewSIPResponseHandler(
			logentry.WithField("module", "response"),
			transactionMonitor,
		)

		// NEW: Initialize sipgo-based call manager
		callManagerSipgo := NewSIPCallManagerSipgo(
			logentry.WithField("module", "sipgo-call"),
			settings,
			networkManager,
		)

		managerInstance = &SIPProxyManager{
			logger:             logentry,
			config:             settings,
			activeCalls:        make(map[string]*SIPProxyCallData),
			callAttempts:       make(map[string]int),
			networkManager:     networkManager,
			responseHandler:    responseHandler,
			callManagerSipgo:   callManagerSipgo,
			transactionMonitor: transactionMonitor,
			upnpManager:        upnpManager,
			sipListener:        sipListener,
		}

		logentry.Infof("ðŸ—ï¸ SIP Proxy Manager inicializado com arquitetura modular usando sipgo")
	})
	return managerInstance
}

// SendSIPInvite inicia uma chamada SIP usando a arquitetura modular
func (m *SIPProxyManager) SendSIPInvite(callID, fromPhone, toPhone string) error {
	return m.SendSIPInviteWithHeaders(callID, fromPhone, toPhone, nil)
}

// SendSIPInviteWithHeaders inicia uma chamada SIP anexando headers adicionais
// ao INVITE. Headers vazios sao ignorados.
func (m *SIPProxyManager) SendSIPInviteWithHeaders(callID, fromPhone, toPhone string, headers map[string]string) error {
	m.logger.Infof("ðŸš€ DEBUG: SendSIPInvite recebeu parÃ¢metros:")
	m.logger.Infof("   ðŸ“ž CallID recebido: %s", callID)
	m.logger.Infof("   ðŸ”µ From recebido: %s", fromPhone)
	m.logger.Infof("   ðŸŸ¢ To recebido: %s", toPhone)

	m.logger.Infof("ðŸ†•ðŸ“ž Iniciando chamada SIP modular usando SIPGO: %s â†’ %s (CallID: %s)", fromPhone, toPhone, callID)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Rastrear tentativa de chamada
	m.callAttempts[callID]++
	attemptNumber := m.callAttempts[callID]

	m.logger.Infof("ðŸ”„ Tentativa #%d para CallID: %s", attemptNumber, callID)

	m.logger.Infof("ðŸ” DEBUG: Passando para InitiateCallSipgo (NEW):")
	m.logger.Infof("   ðŸ“ž CallID que serÃ¡ passado: %s", callID)
	m.logger.Infof("   ðŸ”µ From que serÃ¡ passado: %s", fromPhone)
	m.logger.Infof("   ðŸŸ¢ To que serÃ¡ passado: %s", toPhone)

	// Usar o call manager sipgo para iniciar a chamada
	return m.callManagerSipgo.InitiateCallSipgoWithHeaders(callID, fromPhone, toPhone, headers)
}

// SetLocalRTPPort registers the local UDP RTP port to advertise in the SDP
// offer for callID. The audio bridge calls this with the port of the socket it
// listens on (and sends from), so the SIP server's RTP reaches the bridge.
// Must be called before SendSIPInvite for that call.
func (m *SIPProxyManager) SetLocalRTPPort(callID string, port int) {
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetLocalRTPPort(callID, port)
	}
}

// GetRemoteRTPAddr returns the SIP server's RTP address ("ip:port") parsed from
// the 200 OK SDP answer for callID, once the call has been accepted.
func (m *SIPProxyManager) GetRemoteRTPAddr(callID string) (string, bool) {
	if m.callManagerSipgo != nil {
		return m.callManagerSipgo.GetRemoteRTPAddr(callID)
	}
	return "", false
}

// SetOutboundWhatsAppInviteHandler registers the VoIP-layer callback used when
// this SIP proxy receives an INVITE that should originate a WhatsApp call.
func (m *SIPProxyManager) SetOutboundWhatsAppInviteHandler(handler OutboundWhatsAppInviteHandler) {
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetOutboundWhatsAppInviteHandler(handler)
	}
}

// CreateSDPAnswer builds the SDP body advertised to a SIP caller after the RTP
// port for callID has been registered with SetLocalRTPPort.
func (m *SIPProxyManager) CreateSDPAnswer(callID, fromPhone string) (string, error) {
	if m.callManagerSipgo == nil {
		return "", fmt.Errorf("sipgo call manager is not available")
	}
	return m.callManagerSipgo.CreateSDPOffer(callID, fromPhone)
}

// HangupCall sends a SIP BYE to the server to tear down the call leg. It is
// called when the WhatsApp side ends so the SIP server (asterisk) hangs up too.
// Safe to call for an already-removed call (no-op).
func (m *SIPProxyManager) HangupCall(callID string) {
	if m.callManagerSipgo == nil {
		return
	}
	if err := m.callManagerSipgo.CancelCall(callID); err != nil {
		m.logger.Errorf("HangupCall: failed to send SIP BYE for %s: %v", callID, err)
	}
}

// SetCallAcceptedHandler define o callback para chamadas aceitas
func (m *SIPProxyManager) SetCallAcceptedHandler(handler SIPCallAcceptedCallback) {
	m.logger.Infof("ðŸ“ž Configurando handler para chamadas aceitas")
	m.transactionMonitor.SetCallbacks(handler, m.transactionMonitor.callRejectedHandler)

	// NOVO: Configurar o callback tambÃ©m no sipgo call manager
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallAcceptedHandler(handler)
		m.logger.Infof("âœ… Handler de aceitaÃ§Ã£o tambÃ©m configurado no sipgo call manager")
	}
}

// SetCallRejectedHandler define o callback para chamadas rejeitadas
func (m *SIPProxyManager) SetCallRejectedHandler(handler SIPCallRejectedCallback) {
	m.logger.Infof("âŒ Configurando handler para chamadas rejeitadas")
	m.transactionMonitor.SetCallbacks(m.transactionMonitor.callAcceptedHandler, handler)

	// NOVO: Configurar o callback tambÃ©m no sipgo call manager
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallRejectedHandler(handler)
		m.logger.Infof("âŒ Handler de rejeiÃ§Ã£o tambÃ©m configurado no sipgo call manager")
	}
}

// SetCallTerminatedHandler define o callback para quando o lado SIP remoto
// encerra uma chamada ja estabelecida.
func (m *SIPProxyManager) SetCallTerminatedHandler(handler SIPCallTerminatedCallback) {
	m.logger.Infof("ðŸ“žâ¬…ï¸ Configurando handler para encerramento remoto de chamadas")
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallTerminatedHandler(handler)
	}
}

// Start inicializa e inicia o SIP proxy manager
func (m *SIPProxyManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("SIP proxy manager jÃ¡ estÃ¡ rodando")
	}

	m.logger.Infof("ðŸš€ Iniciando SIP Proxy Manager com arquitetura modular...")

	// Configurar rede (descoberta STUN, etc.)
	if err := m.networkManager.ConfigureNetwork(); err != nil {
		return fmt.Errorf("falha ao configurar rede: %v", err)
	}

	if m.callManagerSipgo != nil {
		if err := m.callManagerSipgo.StartListener(); err != nil {
			return fmt.Errorf("falha ao iniciar listener SIP: %v", err)
		}
	}

	m.isRunning = true
	m.logger.Infof("âœ… SIP Proxy Manager iniciado com sucesso")

	return nil
}

// Stop para graciosamente o SIP proxy manager
func (m *SIPProxyManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("SIP proxy manager nÃ£o estÃ¡ rodando")
	}

	m.logger.Infof("ðŸ›‘ Parando SIP Proxy Manager...")

	// Cancelar todas as chamadas ativas
	for _, callID := range m.callManagerSipgo.GetActiveCalls() {
		if err := m.callManagerSipgo.CancelCall(callID); err != nil {
			m.logger.Errorf("Falha ao cancelar chamada %s: %v", callID, err)
		}
	}

	if m.callManagerSipgo != nil {
		m.callManagerSipgo.StopListener()
	}

	m.isRunning = false
	m.logger.Infof("âœ… SIP Proxy Manager parado com sucesso")

	return nil
}

// IsRunning retorna se o SIP proxy manager estÃ¡ rodando
func (m *SIPProxyManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isRunning
}

// GetActiveCallCount retorna o nÃºmero de chamadas ativas
func (m *SIPProxyManager) GetActiveCallCount() int {
	return len(m.callManagerSipgo.GetActiveCalls())
}

// GetNetworkInfo retorna informaÃ§Ãµes da configuraÃ§Ã£o de rede atual
func (m *SIPProxyManager) GetNetworkInfo() map[string]interface{} {
	return map[string]interface{}{
		"public_ip":     m.networkManager.GetPublicIP(),
		"local_ip":      m.networkManager.GetLocalIP(),
		"local_port":    m.networkManager.GetLocalPort(),
		"sip_server":    m.networkManager.GetSIPServerEndpoint(),
		"is_configured": m.networkManager.IsConfigured(),
	}
}

// GetStats retorna estatÃ­sticas do manager
func (m *SIPProxyManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"is_running":          m.isRunning,
		"active_calls":        m.GetActiveCallCount(),
		"total_call_attempts": len(m.callAttempts),
		"network_configured":  m.networkManager.IsConfigured(),
	}
}

// MÃ©todos de compatibilidade legada
func (m *SIPProxyManager) GetConfig() SIPProxySettings {
	return m.config
}

func (m *SIPProxyManager) GetPublicIP() string {
	return m.networkManager.GetPublicIP()
}

func (m *SIPProxyManager) SetConfig(config SIPProxySettings) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.config = config
	m.logger.Infof("ðŸ“‹ ConfiguraÃ§Ã£o SIP Proxy atualizada")
}

// MÃ©todos avanÃ§ados de gerenciamento de chamadas
func (m *SIPProxyManager) CancelCall(callID string) error {
	return m.callManagerSipgo.CancelCall(callID)
}

func (m *SIPProxyManager) GetCallInfo(callID string) (map[string]interface{}, bool) {
	// sipgo handles call info internally, for now return basic info
	return map[string]interface{}{
		"call_id": callID,
		"state":   "UNKNOWN",
	}, false
}

// Initialize inicializa o SIP proxy manager (compatibilidade legada)
func (m *SIPProxyManager) Initialize() error {
	m.logger.Infof("ðŸ”§ Inicializando SIP Proxy Manager...")
	return m.Start()
}

// RemoveCall remove uma chamada ativa (compatibilidade legada)
func (m *SIPProxyManager) RemoveCall(callID string) {
	m.logger.Infof("ðŸ—‘ï¸ Removendo chamada: %s", callID)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if call exists in active calls map
	if _, exists := m.activeCalls[callID]; exists {
		delete(m.activeCalls, callID)
		m.logger.Infof("âœ… Chamada %s removida do mapeamento ativo", callID)
	} else {
		m.logger.Infof("ðŸ“žâ„¹ï¸ Chamada %s nÃ£o encontrada no mapeamento ativo - pode ter sido removida anteriormente", callID)
	}

	// =========================================================================
	// ðŸš« DUPLICATE BYE PREVENTION: Don't send BYE automatically here
	// =========================================================================
	// The CancelCall() method already handles BYE sending and cleanup
	// RemoveCall() should only remove from tracking, not send SIP messages

	// COMMENTED OUT: Automatic CancelCall (causes duplicate BYEs)
	// Cancel the call in call manager (now handles missing calls gracefully)
	// if err := m.callManagerSipgo.CancelCall(callID); err != nil {
	//	m.logger.Errorf("Erro ao cancelar chamada %s: %v", callID, err)
	// } else {
	//	m.logger.Infof("âœ… Processo de cancelamento de chamada %s concluÃ­do", callID)
	// }

	m.logger.Infof("âœ… Chamada %s removida do rastreamento (sem envio de BYE adicional)", callID)
}

// GetSipgoCallManager returns the underlying SIPGO call manager for advanced operations
// like per-call handler registration
func (m *SIPProxyManager) GetSipgoCallManager() *SIPCallManagerSipgo {
	return m.callManagerSipgo
}

// GetActiveCalls retorna todas as chamadas ativas (compatibilidade legada)
func (m *SIPProxyManager) GetActiveCalls() map[string]*SIPProxyCallData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Retorna uma cÃ³pia do mapa para evitar modificaÃ§Ãµes concorrentes
	activeCalls := make(map[string]*SIPProxyCallData)
	for callID, callData := range m.activeCalls {
		activeCalls[callID] = callData
	}

	return activeCalls
}

// BridgeInboundWhatsAppCall handles an incoming WhatsApp VoIP call by:
//  1. Creating a raw RTP stream (two UDP sockets, no automatic forwarders)
//  2. Storing call state for tracking
//
// The SIP INVITE to the configured SIP server is sent separately by the
// SIPCallManagerSipgo. This method only prepares the RTP media path so the
// VoIPBridge can read/write Î¼-law RTP packets directly.
//
// The caller receives the *RTPStream handle and owns all I/O on its sockets.
func (m *SIPProxyManager) BridgeInboundWhatsAppCall(
	callID string,
	fromPhone string,
	toPhone string,
) (*RTPStream, error) {

	m.logger.Infof("ðŸŒ‰ BridgeInboundWhatsAppCall: CallID=%s, From=%s, To=%s", callID, fromPhone, toPhone)

	// Get local and public IPs from the network manager
	localIP := m.networkManager.GetLocalIP()
	if localIP == "" {
		localIP = "0.0.0.0"
	}
	publicIP := m.networkManager.GetPublicIP()

	// Create a dedicated RTPProxy instance for this bridge call.
	// We use a fresh instance because the bridge owns the sockets and does
	// not share the legacy forwarding model.
	baseLogger := qplog.New()
	rtpProxy := NewRTPProxy(baseLogger, localIP, publicIP)

	// The SIP server host/port come from SIPProxyManager config.
	sipHost := m.config.ServerHost
	sipPort := m.config.ServerPort
	if sipHost == "" {
		return nil, fmt.Errorf("BridgeInboundWhatsAppCall: SIP server host not configured")
	}

	// Create raw RTP stream (no forwarding goroutines).
	stream, err := rtpProxy.CreateRTPStreamRaw(callID, sipHost, sipPort)
	if err != nil {
		return nil, fmt.Errorf("BridgeInboundWhatsAppCall: failed to create RTP stream: %v", err)
	}

	// Store call data for tracking.
	callData := &SIPProxyCallData{
		CallID: callID,
		From:   fromPhone,
		To:     toPhone,
		Status: "initiated",
	}

	m.mutex.Lock()
	m.activeCalls[callID] = callData
	m.mutex.Unlock()

	m.logger.Infof("ðŸŒ‰âœ… Bridge RTP stream ready: CallID=%s, WAPort=%d, SIPPort=%d â†’ %s:%d",
		callID, stream.WhatsAppPort, stream.SIPPort, sipHost, sipPort)

	return stream, nil
}

// BridgeOutboundSIPCall creates a raw RTP stream for a SIP-originated call. The
// remote RTP address is supplied by the caller's SDP offer.
func (m *SIPProxyManager) BridgeOutboundSIPCall(
	callID string,
	fromPhone string,
	toPhone string,
	remoteHost string,
	remotePort int,
) (*RTPStream, error) {
	m.logger.Infof("ðŸŒ‰ BridgeOutboundSIPCall: CallID=%s, From=%s, To=%s, Remote=%s:%d", callID, fromPhone, toPhone, remoteHost, remotePort)

	localIP := m.networkManager.GetLocalIP()
	if localIP == "" {
		localIP = "0.0.0.0"
	}
	publicIP := m.networkManager.GetPublicIP()

	baseLogger := qplog.New()
	rtpProxy := NewRTPProxy(baseLogger, localIP, publicIP)
	stream, err := rtpProxy.CreateRTPStreamRaw(callID, remoteHost, remotePort)
	if err != nil {
		return nil, fmt.Errorf("BridgeOutboundSIPCall: failed to create RTP stream: %v", err)
	}

	callData := &SIPProxyCallData{
		CallID: callID,
		From:   fromPhone,
		To:     toPhone,
		Status: "initiated",
	}

	m.mutex.Lock()
	m.activeCalls[callID] = callData
	m.mutex.Unlock()

	m.logger.Infof("ðŸŒ‰âœ… Outbound SIP RTP stream ready: CallID=%s, WAPort=%d, SIPPort=%d â†’ %s:%d",
		callID, stream.WhatsAppPort, stream.SIPPort, remoteHost, remotePort)

	return stream, nil
}
