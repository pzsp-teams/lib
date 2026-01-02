package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mentions"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type ops struct {
	channelAPI api.ChannelAPI
}

func NewOps(channelAPI api.ChannelAPI) channelOps {
	return &ops{
		channelAPI: channelAPI,
	}
}

func (o *ops) Wait() {}

func (o *ops) ListChannelsByTeamID(ctx context.Context, teamID string) ([]*models.Channel, *snd.RequestError) {
	resp, requestErr := o.channelAPI.ListChannels(ctx, teamID)
	if requestErr != nil {
		return nil, requestErr
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphChannel), nil
}

func (o *ops) GetChannelByID(ctx context.Context, teamID, channelID string) (*models.Channel, *snd.RequestError) {
	resp, requestErr := o.channelAPI.GetChannel(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphChannel(resp), nil
}

func (o *ops) CreateStandardChannel(ctx context.Context, teamID, name string) (*models.Channel, *snd.RequestError) {
	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)
	resp, requestErr := o.channelAPI.CreateStandardChannel(ctx, teamID, newChannel)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphChannel(resp), nil
}

func (o *ops) CreatePrivateChannel(ctx context.Context, teamID, name string, memberIDs, ownerIDs []string) (*models.Channel, *snd.RequestError) {
	resp, requestErr := o.channelAPI.CreatePrivateChannelWithMembers(ctx, teamID, name, memberIDs, ownerIDs)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphChannel(resp), nil
}

func (o *ops) DeleteChannel(ctx context.Context, teamID, channelID, channelRef string) *snd.RequestError {
	return o.channelAPI.DeleteChannel(ctx, teamID, channelID)
}

func (o *ops) SendMessage(ctx context.Context, teamID, channelID string, body models.MessageBody) (*models.Message, *snd.RequestError) {
	ments, err := mentions.PrepareMentions(&body)
	if err != nil {
		return nil, &snd.RequestError{
			Code:    400,
			Message: "Failed to prepare mentions: " + err.Error(),
		}
	}
	resp, requestErr := o.channelAPI.SendMessage(ctx, teamID, channelID, body.Content, string(body.ContentType), ments)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) SendReply(ctx context.Context, teamID, channelID, messageID string, body models.MessageBody) (*models.Message, *snd.RequestError) {
	ments, err := mentions.PrepareMentions(&body)
	if err != nil {
		return nil, &snd.RequestError{
			Code:    400,
			Message: "Failed to prepare mentions: " + err.Error(),
		}
	}
	resp, requestErr := o.channelAPI.SendReply(ctx, teamID, channelID, messageID, body.Content, string(body.ContentType), ments)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListMessages(ctx context.Context, teamID, channelID string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError) {
	var top *int32
	if opts != nil && opts.Top != nil {
		top = opts.Top
	}
	resp, requestErr := o.channelAPI.ListMessages(ctx, teamID, channelID, top)
	if requestErr != nil {
		return nil, requestErr
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) GetMessage(ctx context.Context, teamID, channelID, messageID string) (*models.Message, *snd.RequestError) {
	resp, requestErr := o.channelAPI.GetMessage(ctx, teamID, channelID, messageID)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListReplies(ctx context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError) {
	var top *int32
	if opts != nil && opts.Top != nil {
		top = opts.Top
	}
	resp, requestErr := o.channelAPI.ListReplies(ctx, teamID, channelID, messageID, top)
	if requestErr != nil {
		return nil, requestErr
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMessage), nil
}

func (o *ops) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*models.Message, *snd.RequestError) {
	resp, requestErr := o.channelAPI.GetReply(ctx, teamID, channelID, messageID, replyID)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMessage(resp), nil
}

func (o *ops) ListMembers(ctx context.Context, teamID, channelID string) ([]*models.Member, *snd.RequestError) {
	resp, requestErr := o.channelAPI.ListMembers(ctx, teamID, channelID)
	if requestErr != nil {
		return nil, requestErr
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (o *ops) AddMember(ctx context.Context, teamID, channelID, userID string, isOwner bool) (*models.Member, *snd.RequestError) {
	roles := util.MemberRole(isOwner)
	resp, requestErr := o.channelAPI.AddMember(ctx, teamID, channelID, userID, roles)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, isOwner bool) (*models.Member, *snd.RequestError) {
	roles := util.MemberRole(isOwner)
	resp, requestErr := o.channelAPI.UpdateMemberRoles(ctx, teamID, channelID, memberID, roles)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) RemoveMember(ctx context.Context, teamID, channelID, memberID, userRef string) *snd.RequestError {
	return o.channelAPI.RemoveMember(ctx, teamID, channelID, memberID)
}
