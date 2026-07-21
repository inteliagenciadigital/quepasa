package sipproxy

import (
	"fmt"
	"time"

	qplog "github.com/inteliagenciadigital/quepasa/qplog"
)

// SIPProxyCallAnswerManager manages call answering operations
type SIPProxyCallAnswerManager struct {
	logger     qplog.Logger
	proxy      *SIPProxyManager
	timeout    time.Duration
	retryCount int
}

// NewSIPProxyCallAnswerManager creates a new call answer manager
func NewSIPProxyCallAnswerManager(proxy *SIPProxyManager) *SIPProxyCallAnswerManager {
	return &SIPProxyCallAnswerManager{
		logger:     qplog.New().WithField("component", "call_answer"),
		proxy:      proxy,
		timeout:    30 * time.Second,
		retryCount: 3,
	}
}

// AnswerCall answers an incoming WhatsApp call
// fromPhone = quem estÃ¡ ligando (ex: 557138388109)
// callID = identificador Ãºnico da chamada (ex: BE88BDBDA5C0C1E75D7BD0F8E0E10EBF)
// DEPRECATED: Use AnswerCallWithReceiver instead for correct toPhone parameter
func (cam *SIPProxyCallAnswerManager) AnswerCall(fromPhone, callID string) error {
	cam.logger.Errorf("âŒ DEPRECATED: AnswerCall called without receiver number!")
	cam.logger.Errorf("âŒ This method is deprecated - use AnswerCallWithReceiver instead")
	return fmt.Errorf("AnswerCall is deprecated - missing receiver phone number")
}

// AnswerCallWithReceiver answers an incoming WhatsApp call with explicit receiver number
// fromPhone = quem estÃ¡ ligando (ex: 557138388109)
// toPhone = nÃºmero do WhatsApp que estÃ¡ recebendo (ex: 5521967609494)
// callID = identificador Ãºnico da chamada (ex: BE88BDBDA5C0C1E75D7BD0F8E0E10EBF)
func (cam *SIPProxyCallAnswerManager) AnswerCallWithReceiver(fromPhone, toPhone, callID string) error {
	cam.logger.Infof("ðŸ“ž Answering call from %s to %s (CallID: %s)", fromPhone, toPhone, callID)

	// Send INVITE to SIP proxy with correct parameter order: callID, fromPhone, toPhone
	if err := cam.proxy.SendSIPInvite(callID, fromPhone, toPhone); err != nil {
		cam.logger.Errorf("âŒ Failed to send SIP INVITE: %v", err)
		return err
	}

	cam.logger.Infof("âœ… Call answered and forwarded to SIP proxy")
	return nil
}

// SetTimeout sets the timeout for call operations
func (cam *SIPProxyCallAnswerManager) SetTimeout(timeout time.Duration) {
	cam.timeout = timeout
}

// SetRetryCount sets the number of retries for failed operations
func (cam *SIPProxyCallAnswerManager) SetRetryCount(count int) {
	cam.retryCount = count
}
