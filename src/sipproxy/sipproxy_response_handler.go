package sipproxy

import (
	"strings"

	"github.com/emiago/sipgo/sip"
	qplog "github.com/inteliagenciadigital/quepasa/qplog"
)

// SIPResponseHandler handles SIP response processing
type SIPResponseHandler struct {
	logger             qplog.Logger
	transactionMonitor *SIPTransactionMonitor
}

// NewSIPResponseHandler creates a new SIP response handler
func NewSIPResponseHandler(logger qplog.Logger, transactionMonitor *SIPTransactionMonitor) *SIPResponseHandler {
	return &SIPResponseHandler{
		logger:             logger,
		transactionMonitor: transactionMonitor,
	}
}

// ProcessImmediateResponse processes immediate SIP responses
func (srh *SIPResponseHandler) ProcessImmediateResponse(statusCode, callID, fromPhone, toPhone, response string) {
	srh.logger.Infof("ðŸ” Processing immediate SIP response: %s for CallID: %s", statusCode, callID)

	// Parse status code for proper handling
	switch statusCode {
	case "200":
		srh.logger.Infof("âœ… Call %s immediately ACCEPTED (200 OK)", callID)
		if srh.transactionMonitor.callAcceptedHandler != nil {
			mockResponse := &sip.Response{
				StatusCode: 200,
				Reason:     "OK",
			}
			srh.logger.Infof("ðŸŽ‰ Calling onCallAccepted handler for immediate 200 OK")
			srh.transactionMonitor.callAcceptedHandler(callID, fromPhone, toPhone, mockResponse)
		}
	case "603":
		srh.logger.Infof("âŒ Call %s immediately REJECTED (603 Decline)", callID)
		if srh.transactionMonitor.callRejectedHandler != nil {
			mockResponse := &sip.Response{
				StatusCode: 603,
				Reason:     "Decline",
			}
			srh.logger.Infof("ðŸ’” Calling onCallRejected handler for immediate 603")
			srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
		}
	default:
		if statusCode[0] >= '4' { // 4xx, 5xx, 6xx
			srh.logger.Infof("âŒ Call %s immediately REJECTED (%s)", callID, statusCode)
			if srh.transactionMonitor.callRejectedHandler != nil {
				mockResponse := srh.createMockErrorResponse(statusCode)
				srh.logger.Infof("ðŸ’” Calling onCallRejected handler for %s", statusCode)
				srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
			}
		} else {
			srh.logger.Infof("ðŸ“ž Call %s PROGRESS (%s)", callID, statusCode)
		}
	}
}

// ProcessContinuousResponse processes responses in continuous monitoring
func (srh *SIPResponseHandler) ProcessContinuousResponse(response string, callID, fromPhone, toPhone string) (shouldStop bool) {
	srh.logger.Infof("ðŸ”„ ProcessContinuousResponse called for CallID: %s", callID)
	srh.logger.Infof("ðŸŒŸ SIP SERVER RESPONSE RECEIVED! Processing response for call %s â†’ %s", fromPhone, toPhone)

	// ðŸ†• CHECK FOR SIP REQUESTS (BYE, CANCEL, etc.) - not just responses
	if strings.Contains(response, "BYE ") || strings.Contains(response, "CANCEL ") {
		srh.logger.Infof("ðŸš¨ðŸš¨ðŸš¨ SIP REQUEST RECEIVED FROM SERVER! ðŸš¨ðŸš¨ðŸš¨")
		srh.logger.Infof("ðŸ“¡ THIS IS A SIP REQUEST (not response) - server is asking us to do something!")
		srh.handleSIPRequest(response, callID, fromPhone, toPhone)
		// Continue monitoring for more messages
		return false
	}

	// Parse the status code and handle the response
	if strings.Contains(response, "SIP/2.0") {
		lines := strings.Split(response, "\n")
		if len(lines) > 0 {
			statusLine := strings.TrimSpace(lines[0])
			if strings.HasPrefix(statusLine, "SIP/2.0 ") {
				parts := strings.Fields(statusLine)
				if len(parts) >= 3 {
					statusCode := parts[1]
					srh.logger.Infof("ðŸŽ¯ Processing SIP response: %s for CallID: %s", statusCode, callID)

					// Handle different response types
					switch statusCode {
					case "100":
						srh.logger.Infof("ðŸ“ž 100 Trying received - call is being processed")
						srh.logger.Infof("â³ SIP server is processing the call...")
						// Continue monitoring for final response
						return false
					case "180":
						srh.logger.Infof("ðŸ“ž 180 Ringing received - destination is ringing")
						srh.logger.Infof("ðŸ”” Phone is ringing on SIP server side...")
						// Continue monitoring for final response
						return false
					case "183":
						srh.logger.Infof("ðŸ“ž 183 Session Progress received - early media")
						srh.logger.Infof("ðŸŽµ Session progress with early media...")
						// Continue monitoring for final response
						return false
					case "200":
						srh.logger.Infof("âœ… 200 OK received - CALL ACCEPTED!")
						srh.logger.Infof("ðŸŽ‰ðŸŽ‰ðŸŽ‰ ACCEPT RESPONSE DETECTED! WHATSAPP CALL WILL CONTINUE! ðŸŽ‰ðŸŽ‰ðŸŽ‰")
						srh.logger.Infof("ðŸ“ž Call %s has been ACCEPTED by SIP server", callID)
						if srh.transactionMonitor.callAcceptedHandler != nil {
							mockResponse := &sip.Response{
								StatusCode: 200,
								Reason:     "OK",
							}
							srh.logger.Infof("ðŸŽ‰ Calling onCallAccepted handler for 200 OK response")
							srh.transactionMonitor.callAcceptedHandler(callID, fromPhone, toPhone, mockResponse)
						}

						// ðŸ†• CONTINUE monitoring after 200 OK to capture post-accept messages
						srh.logger.Infof("ðŸ‘‚ CONTINUING to monitor SIP server for post-accept messages (BYE, CANCEL, etc.)")
						srh.logger.Infof("ðŸ” Will log ANY message from SIP server after this 200 OK")
						return false // Changed from true to false - KEEP MONITORING!
					case "603":
						srh.logger.Infof("âŒ 603 Decline received - CALL REJECTED!")
						srh.logger.Infof("ðŸ”¥ðŸ”¥ðŸ”¥ REJECT RESPONSE DETECTED! WILL TERMINATE WHATSAPP CALL! ðŸ”¥ðŸ”¥ðŸ”¥")
						if srh.transactionMonitor.callRejectedHandler != nil {
							mockResponse := &sip.Response{
								StatusCode: 603,
								Reason:     "Decline",
							}
							srh.logger.Infof("ðŸ’” Calling onCallRejected handler for 603 - this should terminate WhatsApp call")
							srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
						} else {
							srh.logger.Errorf("âŒ callRejectedHandler is nil! Cannot process rejection!")
						}
						// Final response - stop monitoring
						return true
					default:
						if statusCode[0] >= '4' { // 4xx, 5xx, 6xx
							srh.logger.Infof("âŒ %s received - CALL REJECTED!", statusCode)
							srh.logger.Infof("ðŸ’” SIP server rejected the call with status: %s", statusCode)
							if srh.transactionMonitor.callRejectedHandler != nil {
								mockResponse := srh.createMockErrorResponse(statusCode)
								srh.logger.Infof("ðŸ’” Calling onCallRejected handler for status: %s", statusCode)
								srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
							}
							// Final response - stop monitoring
							return true
						} else {
							srh.logger.Infof("ðŸ“ž %s received - call progress", statusCode)
							srh.logger.Infof("ðŸ“¡ SIP server sent progress response: %s - continuing to monitor...", statusCode)
							// Continue monitoring for final response
							return false
						}
					}
				}
			}
		}
	} else {
		srh.logger.Warnf("âš ï¸ Non-SIP response received for CallID: %s", callID)
		srh.logger.Warnf("ðŸ¤” Response doesn't contain 'SIP/2.0' - might be malformed or non-SIP")
		srh.logger.Warnf("ðŸ“„ Raw response content: %s", response)
	}

	// Continue monitoring by default
	srh.logger.Infof("ðŸ”„ Continuing to monitor for more responses...")
	return false
}

// LogResponse logs a SIP response with proper formatting
func (srh *SIPResponseHandler) LogResponse(response string, callID string, responseNumber int, bytesReceived int) {
	srh.logger.Infof("ðŸ“¨ SIP Response #%d received (%d bytes) for CallID: %s:", responseNumber, bytesReceived, callID)
	srh.logger.Infof("â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”")

	// Parse and log the response
	for _, line := range strings.Split(response, "\n") {
		if strings.TrimSpace(line) != "" {
			srh.logger.Infof("%s", line)
		}
	}
	srh.logger.Infof("â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”")
}

// createMockErrorResponse creates a mock SIP error response based on status code
func (srh *SIPResponseHandler) createMockErrorResponse(statusCode string) *sip.Response {
	mockResponse := &sip.Response{
		StatusCode: 400, // Default
		Reason:     "Client Error",
	}

	// Parse actual status code if possible
	if len(statusCode) >= 3 {
		switch statusCode {
		case "404":
			mockResponse.StatusCode = 404
			mockResponse.Reason = "Not Found"
		case "486":
			mockResponse.StatusCode = 486
			mockResponse.Reason = "Busy Here"
		case "480":
			mockResponse.StatusCode = 480
			mockResponse.Reason = "Temporarily Unavailable"
		default:
			if statusCode[0] == '5' {
				mockResponse.StatusCode = 500
				mockResponse.Reason = "Server Error"
			} else if statusCode[0] == '6' {
				mockResponse.StatusCode = 600
				mockResponse.Reason = "Global Failure"
			}
		}
	}

	return mockResponse
}

// handleSIPRequest processes SIP requests received from the server (BYE, CANCEL, etc.)
func (srh *SIPResponseHandler) handleSIPRequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("ðŸ” Analyzing SIP request from server for CallID: %s", callID)

	// Log the complete request with detailed formatting
	srh.logSIPRequestDetails(request, callID)

	// Parse the request line to determine the method
	lines := strings.Split(request, "\n")
	if len(lines) > 0 {
		requestLine := strings.TrimSpace(lines[0])
		srh.logger.Infof("ðŸ“‹ SIP Request Line: %s", requestLine)

		// Check for specific SIP methods
		if strings.HasPrefix(requestLine, "BYE ") {
			srh.handleBYERequest(request, callID, fromPhone, toPhone)
		} else if strings.HasPrefix(requestLine, "CANCEL ") {
			srh.handleCANCELRequest(request, callID, fromPhone, toPhone)
		} else if strings.HasPrefix(requestLine, "INVITE ") {
			srh.handleINVITERequest(request, callID, fromPhone, toPhone)
		} else if strings.HasPrefix(requestLine, "ACK ") {
			srh.handleACKRequest(request, callID, fromPhone, toPhone)
		} else {
			srh.handleOtherSIPRequest(requestLine, request, callID, fromPhone, toPhone)
		}
	}
}

// logSIPRequestDetails logs the complete SIP request with formatting
func (srh *SIPResponseHandler) logSIPRequestDetails(request string, callID string) {
	srh.logger.Infof("â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”")
	srh.logger.Infof("ðŸ“¨ COMPLETE SIP REQUEST FROM SERVER (CallID: %s):", callID)
	srh.logger.Infof("â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”")

	lines := strings.Split(request, "\n")
	for i, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			srh.logger.Infof("ðŸ“„ Line %d: %s", i+1, cleanLine)
		}
	}

	srh.logger.Infof("â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”â”")
}

// handleBYERequest processes BYE requests from the server
func (srh *SIPResponseHandler) handleBYERequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("ðŸ›‘ðŸ›‘ðŸ›‘ BYE REQUEST RECEIVED FROM SIP SERVER! ðŸ›‘ðŸ›‘ðŸ›‘")
	srh.logger.Infof("ðŸ“ž SIP server is requesting to TERMINATE the call!")
	srh.logger.Infof("ðŸ’¡ CallID: %s | Call: %s â†’ %s", callID, fromPhone, toPhone)
	srh.logger.Infof("ðŸŽ¯ ACTION NEEDED: Should terminate WhatsApp call and respond with 200 OK")

	// TODO: Implement call termination logic here
	// For now, just log what should happen
	srh.logger.Infof("ðŸ“‹ Next steps (TODO):")
	srh.logger.Infof("   1. ðŸ“žâŒ Terminate WhatsApp call for CallID: %s", callID)
	srh.logger.Infof("   2. ðŸ“¤ Send 200 OK response to SIP server")
	srh.logger.Infof("   3. ðŸ§¹ Clean up call resources")
}

// handleCANCELRequest processes CANCEL requests from the server
func (srh *SIPResponseHandler) handleCANCELRequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("âŒâŒâŒ CANCEL REQUEST RECEIVED FROM SIP SERVER! âŒâŒâŒ")
	srh.logger.Infof("ðŸ“ž SIP server is requesting to CANCEL the ongoing invitation!")
	srh.logger.Infof("ðŸ’¡ CallID: %s | Call: %s â†’ %s", callID, fromPhone, toPhone)
	srh.logger.Infof("ðŸŽ¯ ACTION NEEDED: Should cancel WhatsApp call and respond with 200 OK")

	// TODO: Implement call cancellation logic here
	srh.logger.Infof("ðŸ“‹ Next steps (TODO):")
	srh.logger.Infof("   1. ðŸ“žâŒ Cancel WhatsApp call invitation for CallID: %s", callID)
	srh.logger.Infof("   2. ðŸ“¤ Send 200 OK response to SIP server")
	srh.logger.Infof("   3. ðŸ§¹ Clean up invitation resources")
}

// handleINVITERequest processes INVITE requests from the server
func (srh *SIPResponseHandler) handleINVITERequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("ðŸ“žðŸ“žðŸ“ž INVITE REQUEST RECEIVED FROM SIP SERVER! ðŸ“žðŸ“žðŸ“ž")
	srh.logger.Infof("ðŸ”„ SIP server is sending a new or re-INVITE!")
	srh.logger.Infof("ðŸ’¡ CallID: %s | Call: %s â†’ %s", callID, fromPhone, toPhone)
	srh.logger.Infof("ðŸŽ¯ This might be a call modification or re-negotiation")
}

// handleACKRequest processes ACK requests from the server
func (srh *SIPResponseHandler) handleACKRequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("âœ…âœ…âœ… ACK REQUEST RECEIVED FROM SIP SERVER! âœ…âœ…âœ…")
	srh.logger.Infof("ðŸ¤ SIP server is acknowledging our response!")
	srh.logger.Infof("ðŸ’¡ CallID: %s | Call: %s â†’ %s", callID, fromPhone, toPhone)
	srh.logger.Infof("ðŸŽ¯ This confirms the SIP transaction is complete")
}

// handleOtherSIPRequest processes other SIP requests
func (srh *SIPResponseHandler) handleOtherSIPRequest(requestLine, request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("â“â“â“ UNKNOWN SIP REQUEST RECEIVED FROM SERVER! â“â“â“")
	srh.logger.Infof("ðŸ“‹ Request Line: %s", requestLine)
	srh.logger.Infof("ðŸ’¡ CallID: %s | Call: %s â†’ %s", callID, fromPhone, toPhone)
	srh.logger.Infof("ðŸŽ¯ This is an unrecognized SIP method - logging for analysis")
}
