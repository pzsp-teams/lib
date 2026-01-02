package channels

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type ops struct {
	userAPI   api.UserAPI
	channelAPI api.ChannelAPI
}

func NewOps(channelAPI api.ChannelAPI, userAPI api.UserAPI) channelOps {
	return &ops{
		channelAPI: channelAPI,
		userAPI:    userAPI,
	}
}



func (o *ops) ListChannelsByTeamID(ctx context.Context, teamID string) ([]*models.Channel, error) {
	resp, requestErr := o.channelAPI.ListChannels(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphChannel), nil
}

func (o *ops) GetChannelByID(ctx context.Context, teamID, channelID string) (*models.Channel, error) {
	resp, requestErr := o.channelAPI.GetChannel(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelID))
	}
	return adapter.MapGraphChannel(resp), nil
}

func (o *ops) CreateStandardChannel(ctx context.Context, teamID, name string) (*models.Channel, error) {
	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)
	resp, requestErr := o.channelAPI.CreateStandardChannel(ctx, teamID, newChannel)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, name))
	}
	return adapter.MapGraphChannel(resp), nil
}

func (o *ops) CreatePrivateChannel(ctx context.Context, teamID, name string, memberIDs, ownerIDs []string) (*models.Channel, error) {
	resp, requestErr := o.channelAPI.CreatePrivateChannelWithMembers(ctx, teamID, name, memberIDs, ownerIDs)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, name))
	}
	return adapter.MapGraphChannel(resp), nil
}

func (o *ops) DeleteChannel(ctx context.Context, teamID, channelID, channelRef string) error {
	return snd.MapError(o.channelAPI.DeleteChannel(ctx, teamID, channelID), snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelRef))
}

func (o *ops) SendMessage(ctx context.Context, teamID, channelID string, body models.MessageBody) (*models.Message, error) {
	ments, err := mentions.PrepareMentions(&body)
	if err != nil {
		return nil, snd.MapError(&snd.RequestError{
			Code:    http.StatusBadRequest,
			Message: "Failed to prepare mentions: " + err.Error(),
		}, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelID))
	}
	resp, requestErr := o.channelAPI.SendMessage(ctx, teamID, channelID, body.Content, string(body.ContentType), ments)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelID))
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) SendReply(ctx context.Context, teamID, channelID, messageID string, body models.MessageBody) (*models.Message, error) {
	ments, err := mentions.PrepareMentions(&body)
	if err != nil {
		return nil, snd.MapError(&snd.RequestError{
			Code:    http.StatusBadRequest,
			Message: "Failed to prepare mentions: " + err.Error(),
		}, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelID), snd.WithResource(resources.Message, messageID))
	}
	resp, requestErr := o.channelAPI.SendReply(ctx, teamID, channelID, messageID, body.Content, string(body.ContentType), ments)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelID), snd.WithResource(resources.Message, messageID))
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListMessages(ctx context.Context, teamID, channelID string, opts *models.ListMessagesOptions) ([]*models.Message, error) {
	var top *int32
	if opts != nil && opts.Top != nil {
		top = opts.Top
	}
	resp, requestErr := o.channelAPI.ListMessages(ctx, teamID, channelID, top)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.Channel, channelID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) GetMessage(ctx context.Context, teamID, channelID, messageID string) (*models.Message, error) {
	resp, requestErr := o.channelAPI.GetMessage(ctx, teamID, channelID, messageID)
	if requestErr != nil {
		return nil, snd.MapError(
			requestErr, 
			snd.WithResource(resources.Team, teamID),
			snd.WithResource(resources.Channel, channelID), 
			snd.WithResource(resources.Message, messageID),
		)
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListReplies(ctx context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions) ([]*models.Message, error) {
	var top *int32
	if opts != nil && opts.Top != nil {
		top = opts.Top
	}
	resp, requestErr := o.channelAPI.ListReplies(ctx, teamID, channelID, messageID, top)
	if requestErr != nil {
		return nil, snd.MapError(
			requestErr, 
			snd.WithResource(resources.Team, teamID),
			snd.WithResource(resources.Channel, channelID), 
			snd.WithResource(resources.Message, messageID),
		)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*models.Message, error) {
	resp, requestErr := o.channelAPI.GetReply(ctx, teamID, channelID, messageID, replyID)
	if requestErr != nil {
		return nil, snd.MapError(
			requestErr, 
			snd.WithResource(resources.Team, teamID),
			snd.WithResource(resources.Channel, channelID), 
			snd.WithResource(resources.Message, messageID),
			snd.WithResource(resources.Message, replyID),
		)
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListMembers(ctx context.Context, teamID, channelID string) ([]*models.Member, error) {
	resp, requestErr := o.channelAPI.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, snd.MapError(
			requestErr,
			snd.WithResource(resources.Team, teamID),
			snd.WithResource(resources.Channel, channelID),
		)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (o *ops) AddMember(ctx context.Context, teamID, channelID, userID string, isOwner bool) (*models.Member, error) {
	roles := util.MemberRole(isOwner)
	resp, requestErr := o.channelAPI.AddMember(ctx, teamID, channelID, userID, roles)
	if requestErr != nil {
		return nil, snd.MapError(
			requestErr,
			snd.WithResource(resources.Team, teamID),
			snd.WithResource(resources.Channel, channelID),
			snd.WithResource(resources.User, userID),
		)
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, isOwner bool) (*models.Member, error) {
	roles := util.MemberRole(isOwner)
	resp, requestErr := o.channelAPI.UpdateMemberRoles(ctx, teamID, channelID, memberID, roles)
	if requestErr != nil {
		return nil, snd.MapError(
			requestErr,
			snd.WithResource(resources.Team, teamID),
			snd.WithResource(resources.Channel, channelID),
			snd.WithResource(resources.User, memberID),
		)
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) RemoveMember(ctx context.Context, teamID, channelID, memberID, userRef string) error {
	return snd.MapError(o.channelAPI.RemoveMember(ctx, teamID, channelID, memberID),
		snd.WithResource(resources.Team, teamID),
		snd.WithResource(resources.Channel, channelID),
		snd.WithResource(resources.User, memberID),
	)
}

func (o *ops) GetMentions(ctx context.Context, teamID, teamRef, channelRef, channelID string, rawMentions []string) ([]models.Mention, error) {
	out := make([]models.Mention, 0, len(rawMentions))
	adder := mentions.NewMentionAdder(&out)

	for _, raw := range rawMentions {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if tryAddTeamOrChannelMention(adder, raw, teamRef, teamID, channelRef, channelID) {
			continue
		}

		if util.IsLikelyEmail(raw) {
			if err := adder.AddUserMention(ctx, raw, o.userAPI); err != nil {
				return nil, err
			}
			continue
		}

		return nil, fmt.Errorf("cannot resolve mention reference: %s", raw)
	}

	return out, nil
}