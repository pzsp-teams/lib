package mapper

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
)

type ChannelMapper interface {
	MapChannelRefToChannelID(ctx context.Context, teamID, channelName string) (string, error)
	MapUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error)
}

type channelMapper struct {
	channelsAPI api.ChannelAPI
}

func NewChannelMapper(channelsAPI api.ChannelAPI) ChannelMapper {
	return &channelMapper{channelsAPI: channelsAPI}
}

// MapChannelNameToChannelID will be used later
func (m *channelMapper) MapChannelRefToChannelID(ctx context.Context, teamID, channelRef string) (string, error) {
	ref := strings.TrimSpace(channelRef)
	if ref == "" {
		return "", fmt.Errorf("empty channel reference")
	}
	if isLikelyChannelID(ref) {
		return ref, nil
	}
	chans, err := m.channelsAPI.ListChannels(ctx, teamID)
	if err != nil {
		return "", err
	}
	return resolveChannelIDByName(teamID, ref, chans)
}

func resolveChannelIDByName(teamID, ref string, chans msmodels.ChannelCollectionResponseable) (string, error) {
	if chans == nil || chans.GetValue() == nil || len(chans.GetValue()) == 0 {
		return "", fmt.Errorf("no channels available in team %q", teamID)
	}
	matches := make([]msmodels.Channelable, 0, len(chans.GetValue()))
	for _, c := range chans.GetValue() {
		if c == nil {
			continue
		}
		if deref(c.GetDisplayName()) == ref {
			matches = append(matches, c)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("channel with name %q not found in team %q", ref, teamID)
	case 1:
		id := deref(matches[0].GetId())
		if id == "" {
			return "", fmt.Errorf("channel %q has nil id in team %q", ref, teamID)
		}
		return id, nil
	default:
		var options []string
		for _, c := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", deref(c.GetDisplayName()), deref(c.GetId())))
		}
		return "", fmt.Errorf(
			"multiple channels named %q found in team %q: \n%s.\nPlease use one of the IDs instead",
			ref, teamID, strings.Join(options, ";\n"),
		)
	}
}

func (m *channelMapper) MapUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	resp, err := m.channelsAPI.ListMembers(ctx, teamID, channelID)
	if err != nil {
		return "", err
	}
	if resp == nil || resp.GetValue() == nil || len(resp.GetValue()) == 0 {
		return "", fmt.Errorf("no members found in channel %q", channelID)
	}

	for _, member := range resp.GetValue() {
		if member == nil {
			continue
		}
		um, ok := member.(msmodels.AadUserConversationMemberable)
		if !ok {
			continue
		}
		if matchesUserRef(um, userRef) {
			return deref(member.GetId()), nil
		}
	}

	return "", fmt.Errorf("member %q not found in channel %q", userRef, channelID)
}

func isLikelyChannelID(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "19:") && strings.Contains(s, "@thread.")
}

func matchesUserRef(um msmodels.AadUserConversationMemberable, userRef string) bool {
	if userRef == "" {
		return false
	}
	if deref(um.GetUserId()) == userRef {
		return true
	}
	if deref(um.GetDisplayName()) == userRef {
		return true
	}
	ad := um.GetAdditionalData()
	if ad == nil {
		return false
	}
	for _, key := range []string{"userPrincipalName", "mail", "email"} {
		raw, ok := ad[key]
		if !ok {
			continue
		}
		if v, ok := raw.(string); ok && strings.EqualFold(v, userRef) {
			return true
		}
	}
	return false
}
