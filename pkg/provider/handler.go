package provider

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderSimulationMode = "X-Simulation-Mode"
	HeaderCorrelationID  = "X-Correlation-ID"

	ModeAccepted     = "accepted"
	ModeRejected     = "rejected"
	ModeUnavailable  = "unavailable"
	ModeTimeout      = "timeout"
	ModeMalformed    = "malformed"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/manifest", h.Manifest)
	r.POST("/invocations", h.Invoke)
	r.POST("/test", h.Test)
}

func (h *Handler) Manifest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"protocol":    "provider/v1",
		"provider":    "mock-sms-provider",
		"operations":  []string{"send-sms"},
		"simulations": []string{ModeAccepted, ModeRejected, ModeUnavailable, ModeTimeout, ModeMalformed},
	})
}

type invokeRequest struct {
	Operation string         `json:"operation"`
	Input     map[string]any `json:"input"`
}

func (h *Handler) Invoke(c *gin.Context) {
	h.handle(c, false)
}

func (h *Handler) Test(c *gin.Context) {
	h.handle(c, true)
}

func (h *Handler) handle(c *gin.Context, isTest bool) {
	var req invokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}
	op := strings.TrimSpace(req.Operation)
	if op == "" {
		op = "send-sms"
	}
	if op != "send-sms" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_operation", "detail": op})
		return
	}
	to, _ := req.Input["to"].(string)
	message, _ := req.Input["message"].(string)
	if strings.TrimSpace(to) == "" || strings.TrimSpace(message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_input",
			"detail": "input.to and input.message are required",
		})
		return
	}

	mode := strings.ToLower(strings.TrimSpace(c.GetHeader(HeaderSimulationMode)))
	if mode == "" {
		mode = ModeAccepted
	}
	corr := c.GetHeader(HeaderCorrelationID)
	if corr == "" {
		if v, ok := c.Get("correlation_id"); ok {
			corr, _ = v.(string)
		}
	}

	switch mode {
	case ModeTimeout:
		time.Sleep(3 * time.Second)
		c.JSON(http.StatusGatewayTimeout, gin.H{
			"outcome":        "timeout",
			"provider_ref":   "",
			"correlation_id": corr,
			"detail":         gin.H{"simulation": ModeTimeout},
		})
		return
	case ModeUnavailable:
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"outcome":        "provider_unavailable",
			"provider_ref":   "",
			"correlation_id": corr,
			"detail":         gin.H{"simulation": ModeUnavailable},
		})
		return
	case ModeMalformed:
		c.Data(http.StatusOK, "text/plain", []byte("not-json"))
		return
	case ModeRejected:
		c.JSON(http.StatusOK, gin.H{
			"outcome":        "rejected",
			"provider_ref":   "",
			"correlation_id": corr,
			"detail": gin.H{
				"simulation": ModeRejected,
				"reason":     "simulated rejection",
				"to":         to,
				"test":       isTest,
			},
		})
		return
	case ModeAccepted:
		c.JSON(http.StatusOK, gin.H{
			"outcome":        "accepted",
			"provider_ref":   "mock-sms-" + uuid.NewString(),
			"correlation_id": corr,
			"detail": gin.H{
				"simulation": ModeAccepted,
				"to":         to,
				"message":    message,
				"test":       isTest,
			},
		})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid_simulation_mode",
			"detail": mode,
		})
	}
}
