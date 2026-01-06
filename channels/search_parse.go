package channels

import (
	"strings"

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
				if hit == nil || hit.GetResource() == nil {
					continue
				}
				if msg, ok := hit.GetResource().(msmodels.ChatMessageable); ok {
					out = append(out, msg)
				}
			}
		}
	}

	return out
}

func isFromChannel(msg msmodels.ChatMessageable, teamID, channelID string) bool {
	if msg == nil {
		return false
	}
	ci := msg.GetChannelIdentity()
	if ci == nil {
		return false
	}
	tid := ci.GetTeamId()
	cid := ci.GetChannelId()
	return tid != nil && cid != nil && *tid == teamID && *cid == channelID
}

func isAuthoredBy(msg msmodels.ChatMessageable, userID string) bool {
	if msg == nil {
		return false
	}
	from := msg.GetFrom()
	if from == nil || from.GetUser() == nil || from.GetUser().GetId() == nil {
		return false
	}
	return *from.GetUser().GetId() == userID
}

func isSystemEvent(msg msmodels.ChatMessageable) bool {
	if msg == nil || msg.GetMessageType() == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace((msg.GetMessageType()).String()), "systemEventMessage")
}
