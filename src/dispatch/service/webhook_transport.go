package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	whatsapp "github.com/inteliagenciadigital/quepasa/whatsapp"
	log "github.com/inteliagenciadigital/quepasa/qplog"
)

// WebhookRequest is the outbound HTTP contract used by the dispatch module
// to deliver events/messages to external webhook endpoints.
type WebhookRequest struct {
	ConnectionString string
	Wid              string
	Extra            interface{}
	Timeout          time.Duration
}

type WebhookResponse struct {
	StatusCode int
	Duration   time.Duration
	TimedOut   bool
}

type webhookPayload struct {
	*whatsapp.WhatsappMessage
	Extra interface{} `json:"extra,omitempty"`
}

// SendWebhook performs external HTTP delivery for one message.
// The caller owns domain-level metrics/state updates.
func SendWebhook(message *whatsapp.WhatsappMessage, request *WebhookRequest, logger log.Logger) (*WebhookResponse, error) {
	if request == nil {
		return &WebhookResponse{}, nil
	}

	startTime := time.Now()

	payload := &webhookPayload{
		WhatsappMessage: message,
		Extra:           request.Extra,
	}

	payloadJSON, err := json.Marshal(&payload)
	if err != nil {
		return &WebhookResponse{}, err
	}

	if logger != nil {
		logger.Debugf("posting webhook payload: %s", payloadJSON)
	}

	client := &http.Client{}
	if request.Timeout > 0 {
		client.Timeout = request.Timeout
	}

	result := &WebhookResponse{}

	operation := func() error {
		req, reqErr := http.NewRequest("POST", request.ConnectionString, bytes.NewReader(payloadJSON))
		if reqErr != nil {
			return backoff.Permanent(reqErr)
		}
		req.Header.Set("User-Agent", "Quepasa")
		req.Header.Set("X-QUEPASA-WID", request.Wid)
		req.Header.Set("Content-Type", "application/json")

		resp, doErr := client.Do(req)
		if doErr != nil {
			if netErr, ok := doErr.(interface{ Timeout() bool }); ok {
				result.TimedOut = netErr.Timeout()
			}
			return doErr
		}
		defer resp.Body.Close()

		result.StatusCode = resp.StatusCode
		// Only retry on 5xx errors or network errors. 4xx errors should be permanent.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return backoff.Permanent(fmt.Errorf("invalid webhook response status (client error): %d", resp.StatusCode))
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("invalid webhook response status: %d", resp.StatusCode)
		}

		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.Multiplier = 1.5
	b.MaxInterval = 10 * time.Second
	b.MaxElapsedTime = 60 * time.Second

	err = backoff.RetryNotify(operation, b, func(err error, t time.Duration) {
		if logger != nil {
			logger.Warnf("Retrying webhook send... Error: %v, Next attempt in: %v", err, t)
		}
	})

	result.Duration = time.Since(startTime)

	return result, err
}
