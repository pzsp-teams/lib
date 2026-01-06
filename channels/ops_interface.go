package channels

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

type channelOps interface {
	ListChannelsByTeamID(ctx context.Context, teamID string) ([]*models.Channel, error)
	GetChannelByID(ctx context.Context, teamID, channelID string) (*models.Channel, error)
	CreateStandardChannel(ctx context.Context, teamID, name string) (*models.Channel, error)
	CreatePrivateChannel(ctx context.Context, teamID, name string, memberIDs, ownerIDs []string) (*models.Channel, error)
	DeleteChannel(ctx context.Context, teamID, channelID, channelRef string) error
	SendMessage(ctx context.Context, teamID, channelID string, body models.MessageBody) (*models.Message, error)
	SendReply(ctx context.Context, teamID, channelID, messageID string, body models.MessageBody) (*models.Message, error)
	ListMessages(ctx context.Context, teamID, channelID string, opts *models.ListMessagesOptions, includeSystem bool) (*models.MessageCollection, error)
	GetMessage(ctx context.Context, teamID, channelID, messageID string) (*models.Message, error)
	ListReplies(ctx context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions, includeSystem bool) (*models.MessageCollection, error)
	GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (*models.Message, error)
	ListMembers(ctx context.Context, teamID, channelID string) ([]*models.Member, error)
	AddMember(ctx context.Context, teamID, channelID, userID string, isOwner bool) (*models.Member, error)
	UpdateMemberRoles(ctx context.Context, teamID, channelID, memberID string, isOwner bool) (*models.Member, error)
	RemoveMember(ctx context.Context, teamID, channelID, memberID, userRef string) error
	GetMentions(ctx context.Context, teamID, teamRef, channelRef, channelID string, rawMentions []string) ([]models.Mention, error)
	ListMessagesNext(ctx context.Context, teamID, channelID, nextLink string, includeSystem bool) (*models.MessageCollection, error)
	ListRepliesNext(ctx context.Context, teamID, channelID, messageID, nextLink string, includeSystem bool) (*models.MessageCollection, error)
	SearchMessagesInChannel(ctx context.Context, teamID, channelID string, opts *models.SearchMessagesOptions) ([]*models.Message, error)
}
