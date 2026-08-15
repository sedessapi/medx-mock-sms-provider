package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseSynermaxxACK(t *testing.T) {
	body := []byte(`{"code":200,"status":"ACK","transid":"TX-live-1","mobilenum":"09662060323","originator":"22566100"}`)
	got := parseSynermaxxResponse(200, body)
	if !got.Accepted {
		t.Fatalf("ACK should be accepted: %+v", got)
	}
	if got.ProviderStatus != "ACK" {
		t.Fatalf("status=%q", got.ProviderStatus)
	}
	if got.ProviderMessageID != "TX-live-1" {
		t.Fatalf("transid=%q", got.ProviderMessageID)
	}
}

func TestParseSynermaxxNACK(t *testing.T) {
	body := []byte(`{"code":400,"status":"NACK","message":"invalid destination"}`)
	got := parseSynermaxxResponse(200, body)
	if got.Accepted {
		t.Fatalf("NACK should be rejected: %+v", got)
	}
}

func TestLiveInvokeMapsToSynermaxx(t *testing.T) {
	var gotBody map[string]any
	var gotAppKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAppKey = r.Header.Get("appkey")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"status":"ACK","transid":"TX-42"}`))
	}))
	t.Cleanup(srv.Close)

	h := NewHandler(NewSynermaxxProvider(SynermaxxConfig{
		URL:        srv.URL,
		AppKey:     "test-key",
		Originator: "22566100",
		HTTPClient: srv.Client(),
	}))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/provider/v1/invocations", strings.NewReader(
		`{"operation":"send-sms","input":{"to":"09662060323","message":"Assumpta Est Mariae"}}`,
	))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.Invoke(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["outcome"] != "accepted" {
		t.Fatalf("outcome=%v", resp["outcome"])
	}
	if resp["provider_ref"] != "TX-42" {
		t.Fatalf("provider_ref=%v", resp["provider_ref"])
	}
	if strings.HasPrefix(resp["provider_ref"].(string), "mock-sms-") {
		t.Fatal("live path must not emit mock-sms provider_ref")
	}
	if gotAppKey != "test-key" {
		t.Fatalf("appkey header=%q", gotAppKey)
	}
	if gotBody["mobilenum"] != "09662060323" {
		t.Fatalf("mobilenum=%v", gotBody["mobilenum"])
	}
	if gotBody["fullmesg"] != "Assumpta Est Mariae" {
		t.Fatalf("fullmesg=%v", gotBody["fullmesg"])
	}
}

func TestLiveInvokeRejectsNonPHNumber(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	h := NewHandler(NewSynermaxxProvider(SynermaxxConfig{
		URL:    srv.URL,
		AppKey: "test-key",
	}))
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/provider/v1/invocations", strings.NewReader(
		`{"operation":"send-sms","input":{"to":"+15559876110","message":"no"}}`,
	))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.Invoke(c)

	if called {
		t.Fatal("invalid number must not call Synermaxx")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
