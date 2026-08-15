package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"medx-mock-sms-provider/config"
	"medx-mock-sms-provider/pkg/platformhttp"
	"medx-mock-sms-provider/pkg/provider"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	gin.SetMode(cfg.GinMode)

	r := gin.Default()
	r.Use(platformhttp.CorrelationMiddleware())
	platformhttp.MountOps(r, "medx-mock-sms-provider", openAPISpec())

	var live *provider.SynermaxxProvider
	if cfg.LiveSynermaxx() {
		if cfg.SynermaxxAppKey == "" {
			log.Fatalf("config: SYNERMAXX_SMS_APPKEY is required when SMS_PROVIDER=synermaxx")
		}
		live = provider.NewSynermaxxProvider(provider.SynermaxxConfig{
			URL:        cfg.SynermaxxURL,
			AppKey:     cfg.SynermaxxAppKey,
			Originator: cfg.SynermaxxOriginator,
		})
	}
	h := provider.NewHandler(live)
	h.RegisterRoutes(r.Group("/provider/v1"))

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	if live != nil {
		log.Printf("medx-mock-sms-provider listening on %s (Provider Protocol v1, live synermaxx)", addr)
	} else {
		log.Printf("medx-mock-sms-provider listening on %s (Provider Protocol v1)", addr)
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func openAPISpec() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "medx-mock-sms-provider",
			"version":     "v1",
			"description": "Mock SMS Provider — Provider Protocol v1 for PSB-1 Send SMS vertical slice.",
		},
		"paths": map[string]any{
			"/healthz":                 map[string]any{"get": map[string]any{"summary": "Liveness"}},
			"/readyz":                  map[string]any{"get": map[string]any{"summary": "Readiness"}},
			"/provider/v1/manifest":    map[string]any{"get": map[string]any{"summary": "Provider manifest"}},
			"/provider/v1/invocations": map[string]any{"post": map[string]any{"summary": "Invoke send-sms"}},
			"/provider/v1/test":        map[string]any{"post": map[string]any{"summary": "Design-time test invoke"}},
		},
	}
}
