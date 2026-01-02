package channels

import (
	"context"

	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type channelOps interface {
	Wait()
	ListChannelsByTeamID(ctx context.Context, teamID string) ([]*models.Channel, *snd.RequestError)
	GetChannelByID(ctx context.Context, teamID, channelID string) (*models.Channel, *snd.RequestError)
	CreateStandardChannel(ctx context.Context, teamID, name string) (*models.Channel, *snd.RequestError)
	CreatePrivateChannel(ctx context.Context, teamID, name string, memberIDs, ownerIDs []string) (*models.Channel, *snd.RequestError)
	DeleteChannel(ctx context.Context, teamID, channelID, channelRef string) *snd.RequestError
	SendMessage(ctx context.Context, teamID, channelID string, body models.MessageBody) (*models.Message, *snd.RequestError)
	SendReply(ctx context.Context, teamID, channelID, messageID string, body models.MessageBody) (*models.Message, *snd.RequestError)
	ListMessages(ctx context.Context, teamID, channelID string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError)
	GetMessage(ctx context.Context, teamID, channelID, messageID string) (*models.Message, *snd.RequestError)
	ListReplies(ctx context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError)
	GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*models.Message, *snd.RequestError)
	ListMembers(ctx context.Context, teamID, channelID string) ([]*models.Member, *snd.RequestError)
	AddMember(ctx context.Context, teamID, channelID, userID string, isOwner bool) (*models.Member, *snd.RequestError)
	UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, isOwner bool) (*models.Member, *snd.RequestError)
	RemoveMember(ctx context.Context, teamID, channelID, memberID, userRef string) *snd.RequestError
	// GetMentions(ctx context.Context, teamID, channelID, rawMentions []string) ([]*models.Mention, error)
}
