package channels

import (
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"
)

func extractChatMessages(resp graphsearch.QueryPostResponseable) []msmodels.ChatMessageable {
	if resp == nil || resp.GetValue() == nil {
		return nil
	}

	out := make([]msmodels.ChatMessageable, 0, 32)

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
				if hit == nil || resource == nil {
					continue
				}
				if msg, ok := resource.(msmodels.ChatMessageable); ok {
					out = append(out, msg)
				}
			}
		}
	}

	return out
}
