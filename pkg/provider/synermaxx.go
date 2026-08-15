package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultSynermaxxURL = "https://svr36.synermaxx.asia/api/sendmsg.php"

type SynermaxxConfig struct {
	URL        string
	AppKey     string
	Originator string
	HTTPClient *http.Client
}

type SynermaxxProvider struct {
	url        string
	appKey     string
	originator string
	client     *http.Client
}

type synermaxxRequest struct {
	MobileNum  string `json:"mobilenum"`
	Originator string `json:"originator"`
	FullMesg   string `json:"fullmesg"`
	Format     string `json:"format"`
	Blackhole  string `json:"blackhole"`
}

type SendResult struct {
	Accepted          bool
	ProviderStatus    string
	ProviderMessageID string
	ErrorCode         string
	ErrorMessage      string
	HTTPStatus        int
}

func NewSynermaxxProvider(cfg SynermaxxConfig) *SynermaxxProvider {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		url = defaultSynermaxxURL
	}
	originator := strings.TrimSpace(cfg.Originator)
	if originator == "" {
		originator = "22566100"
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &SynermaxxProvider{
		url:        url,
		appKey:     strings.TrimSpace(cfg.AppKey),
		originator: originator,
		client:     client,
	}
}

func (p *SynermaxxProvider) Send(ctx context.Context, mobile, message string) (SendResult, error) {
	if p.appKey == "" {
		return SendResult{}, fmt.Errorf("synermaxx appkey not configured")
	}

	payload := synermaxxRequest{
		MobileNum:  mobile,
		Originator: p.originator,
		FullMesg:   message,
		Format:     "json",
		Blackhole:  "0",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return SendResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("appkey", p.appKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return SendResult{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return SendResult{}, err
	}
	return parseSynermaxxResponse(resp.StatusCode, respBody), nil
}

func parseSynermaxxResponse(statusCode int, body []byte) SendResult {
	result := SendResult{
		HTTPStatus:     statusCode,
		ProviderStatus: "unknown",
	}

	if statusCode < 200 || statusCode >= 300 {
		result.Accepted = false
		result.ErrorCode = "http_error"
		result.ErrorMessage = fmt.Sprintf("synermaxx HTTP %d: %s", statusCode, strings.TrimSpace(string(body)))
		result.ProviderStatus = "failed"
		return result
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		result.Accepted = true
		result.ProviderStatus = "sent"
		result.ProviderMessageID = strings.TrimSpace(string(body))
		return result
	}

	result.ProviderMessageID = firstString(parsed,
		"transid", "messageid", "message_id", "msgid", "msg_id", "id", "transaction_id",
	)
	result.ProviderStatus = firstString(parsed, "status", "result", "response", "state")
	if result.ProviderStatus == "" {
		if code := firstString(parsed, "code"); code == "200" {
			result.ProviderStatus = "ACK"
		} else {
			result.ProviderStatus = "sent"
		}
	}

	if synermaxxIndicatesFailure(parsed, result.ProviderStatus) {
		result.Accepted = false
		result.ErrorCode = "provider_rejected"
		result.ErrorMessage = firstString(parsed, "error", "err", "message", "description", "reason")
		if result.ErrorMessage == "" {
			result.ErrorMessage = strings.TrimSpace(string(body))
		}
		return result
	}

	result.Accepted = true
	return result
}

func synermaxxIndicatesFailure(parsed map[string]any, status string) bool {
	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch statusLower {
	case "ack", "ok", "success", "sent", "accepted", "queued", "delivered", "0":
		return false
	case "nack", "error", "failed", "fail", "rejected", "invalid":
		return true
	}

	if code := firstString(parsed, "code"); code != "" && code != "200" {
		return true
	}
	return false
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case string:
				if s := strings.TrimSpace(t); s != "" {
					return s
				}
			case float64:
				return fmt.Sprintf("%.0f", t)
			}
		}
	}
	return ""
}
