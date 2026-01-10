package api

import (
	"strings"

	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"
)

func extractMessages(resp graphsearch.QueryPostResponseable) []SearchEntity {
	if resp == nil || resp.GetValue() == nil {
		return nil
	}

	out := make([]SearchEntity, 0, 25)

	for _, sr := range resp.GetValue() {
		if sr == nil || sr.GetHitsContainers() == nil {
			continue
		}
		for _, hc := range sr.GetHitsContainers() {
			if hc == nil || hc.GetHits() == nil {
				continue
			}
			for _, hit := range hc.GetHits() {
				if hit == nil {
					continue
				}
				resource := hit.GetResource()
				if resource == nil {
					continue
				}
				msgID := resource.GetId()
				if msgID == nil || strings.TrimSpace(*msgID) == "" {
					continue
				}

				var teamID, channelID, chatID *string
				if ad := resource.GetAdditionalData(); ad != nil {
					if ciRaw, ok := ad["channelIdentity"]; ok {
						if ciMap, ok := ciRaw.(map[string]any); ok {
							teamID = asStringPtr(ciMap["teamId"])
							channelID = asStringPtr(ciMap["channelId"])
						}
					}
					if chatIDRaw, ok := ad["chatId"]; ok {
						chatID = asStringPtr(chatIDRaw)
					}
				}

				out = append(out, SearchEntity{
					MessageID: msgID,
					TeamID:    teamID,
					ChannelID: channelID,
					ChatID:    chatID,
				})
			}
		}
	}

	return out
}

func asStringPtr(v any) *string {
	switch t := v.(type) {
	case nil:
		return nil
	case *string:
		if t == nil {
			return nil
		}
		s := strings.TrimSpace(*t)
		if s == "" {
			return nil
		}
		return &s
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		return &s
	default:
		return nil
	}
}
