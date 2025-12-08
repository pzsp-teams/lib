package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pzsp-teams/lib"
)

type Request struct {
	Type   string                 `json:"type"`
	Method string                 `json:"method,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

var client *lib.Client
var initialized bool

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()

		var req Request
		err := json.Unmarshal([]byte(line), &req)
		if err != nil {
			respondError(writer, fmt.Errorf("invalid json: %w", err))
			continue
		}

		if req.Type == "init" {
			if initialized {
				respondError(writer, fmt.Errorf("client already initialized"))
				continue
			}

			configMap := req.Config

			scopes, err := safeScopes(configMap)
			if err != nil {
				respondError(writer, err)
				continue
			}
			authMethod, err := validateAuthMethod(safeString(configMap, "authMethod"))
			if err != nil {
				respondError(writer, err)
				continue
			}
			authConfig := lib.AuthConfig{
				ClientID:     safeString(configMap, "clientID"),
				Tenant:     safeString(configMap, "tenant"),
				Email:		safeString(configMap, "email"),
				Scopes:     scopes,
				AuthMethod: authMethod,
			}

			senderConfig := lib.SenderConfig{
				MaxRetries: safeInt(configMap, "maxRetries"),
				NextRetryDelay: safeInt(configMap, "nextRetryDelay"),
				Timeout: safeInt(configMap, "timeout"),
			}

			c, err := lib.NewClient(context.TODO(), &authConfig, &senderConfig)
			if err != nil {
				respondError(writer, err)
				continue
			}
			client = c
			initialized = true
			respondResult(writer, "initialized")
			continue
		}

		if req.Type == "request" {
			if client == nil {
				respondError(writer, fmt.Errorf("client not initialized"))
				continue
			}

			switch req.Method {
			case "listChannels":
				teamRef := safeString(req.Params, "teamRef")
				if teamRef == "" {
					respondError(writer, fmt.Errorf("invalid teamRef parameter"))
					continue
				}
				channels, err := client.Channels.ListChannels(context.TODO(), teamRef)
				if err != nil {
					respondError(writer, err)
				} else {
					respondResult(writer, channels)
				}
			default:
				respondError(writer, fmt.Errorf("unknown method"))
			}
			continue
		}
		respondError(writer, fmt.Errorf("unknown request type"))
	}
}
