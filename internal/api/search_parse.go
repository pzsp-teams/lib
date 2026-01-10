package api

import (
	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"
)

func extractMessages(resp graphsearch.QueryPostResponseable) []SearchEntity {
	if resp == nil || resp.GetValue() == nil {
		return nil
	}

	out := make([]SearchEntity, 0, len(resp.GetValue()))

	for _, sr := range resp.GetValue() {
		if sr == nil || sr.GetHitsContainers() == nil {
			continue
		}
		for _, hc := range sr.GetHitsContainers() {
			if hc == nil || hc.GetHits() == nil {
				continue
			}
			for _, hit := range hc.GetHits() {
				resource := hit.GetResource()
				if resource == nil {
					continue
				}
				channelIdentity, ok := resource.GetAdditionalData()["channelIdentity"]
				if ok {
					if teamID, ok := channelIdentity.(map[string]any)["teamId"]; ok && teamID != nil {
						out = append(out, SearchEntity{
							ChannelID: channelIdentity.(map[string]any)["channelId"].(*string),
							TeamID:    channelIdentity.(map[string]any)["teamId"].(*string),
							MessageID: resource.GetId(),
						})
					} else {
						if chatID, ok := resource.GetAdditionalData()["chatId"]; ok {
							out = append(out, SearchEntity{
								ChatID:    chatID.(*string),
								MessageID: resource.GetId(),
							})
						}
					}
				}
			}
		}
	}

	return out
}
