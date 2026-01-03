package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pzsp-teams/lib/config"
)

func newSenderConfig() *config.SenderConfig {
	return &config.SenderConfig{
		MaxRetries:     3,
		NextRetryDelay: 2,
		Timeout:        10,
	}
}

func loadAuthConfig() *config.AuthConfig {
	_ = godotenv.Load()
	cfg := &config.AuthConfig{
		ClientID:   getEnv("CLIENT_ID", ""),
		Tenant:     getEnv("TENANT_ID", ""),
		Email:      getEnv("EMAIL", ""),
		Scopes:     strings.Split(getEnv("SCOPES", "https://graph.microsoft.com/.default"), ","),
		AuthMethod: config.Method(getEnv("AUTH_METHOD", "DEVICE_CODE")),
	}
	validate(cfg)
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func validate(cfg *config.AuthConfig) {
	if cfg.ClientID == "" {
		log.Fatal("Missing CLIENT ID")
	}
	if cfg.Tenant == "" {
		log.Fatal("Missing TENANT ID")
	}
	if cfg.Email == "" {
		log.Fatal("Missing EMAIL")
	}
	if cfg.AuthMethod != "DEVICE_CODE" && cfg.AuthMethod != "INTERACTIVE" {
		log.Fatal("AUTH METHOD must be either DEVICE_CODE or INTERACTIVE")
	}
}
