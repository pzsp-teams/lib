// Package channels provides various channel-related operations. It abstracts the underlying Microsoft Graph API calls.
// Package provides two services implementations - one with cache and one without cache. Both can be instantiated and used interchangeably.
// If cache is enabled, the service will use a caching layer to store and retrieve channel/member references, improving performance and reducing API calls.
// Concepts:
//   - Channels belong to teams.
//   - Channels can be standard or private.
//   - Users are identified by userID or email.
//   - Some operations require messageID - these can be obtained via ListMessages.
//   - ChannelRef is a reference (display name or ID) to a channel used in method parameters.
//   - If teamRef or channelRef is a display and is not unique, an ambiguity error is returned.
//   - The authenticated user (derived from MSAL) is the one making the API calls (appropriate scopes must be granted).
//
// If an async cached service is used, call Wait() to ensure all background cache updates are finished.
package channels

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

// Service defines the interface for channel-related operations.
// It includes methods for managing channels, members, messages, and more.
type Service interface {
	// ListChannels returns all channels in a team.
	ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error)

	// Get retrieves a specific channel by its reference (ID or display name) within a team.
	Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error)

	// CreateStandardChannel creates a standard channel within a team.
	// Standard channels are open to all team members.
	CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error)

	// CreatePrivateChannel creates a private channel within a team.
	// Private channels are restricted to specific members.
	// At least one owner must be specified.
	CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error)

	// Delete removes a channel from a team.
	Delete(ctx context.Context, teamRef, channelRef string) error

	// SendMessage sends a message to a channel.
	// Body parameter is the body of the message. It includes:
	//   - Content: the text or html content of the message.
	//   - ContentType: the type of content (text or html).
	//   - Mentions: optional mentions to include in the message.
	SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error)

	// SendReply sends a reply to a specific message in a channel.
	// Body parameter is the body of the reply message. It includes:
	//   - Content: the text or html content of the message.
	//   - ContentType: the type of content (text or html).
	//   - Mentions: optional mentions to include in the message.
	SendReply(ctx context.Context, teamRef, channelRef, messageID string, body models.MessageBody) (*models.Message, error)

	// ListMessages returns one page of messages in a channel.
	//
	// NextLink in the returned MessageCollection can be used to retrieve the next page of messages.
	ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions, includeSystem bool, nextLink *string) (*models.MessageCollection, error)

	// GetMessage retrieves a specific message from a channel by its ID.
	GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error)

	// ListReplies returns one page of replies to a specific message in a channel.
	//
	// NextLink in the returned MessageCollection can be used to retrieve the next page of replies.
	ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32, includeSystem bool, nextLink *string) (*models.MessageCollection, error)

	// GetReply retrieves a specific reply to a message in a channel by its ID.
	GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error)

	// ListMembers returns all members of a channel.
	ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error)

	// AddMember adds a user to a channel.
	AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error)

	// UpdateMemberRoles updates the roles of a member in a channel.
	UpdateMemberRoles(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error)

	// RemoveMember removes a user from a channel.
	RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error

	// GetMentions resolves raw mention strings to Mention objects in the context of a channel. Raw mentions can be:
	//   - Emails
	//   - Channel (only the same channel as channelRef can be mentioned). It can be used by specifying "channel" or channel display name as raw mention.
	//   - Team (only the parent team of the channel can be mentioned). It can be used by specifying "team" or team display name as raw mention.
	//   - User IDs
	GetMentions(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error)

	// SearchMessagesInChannel searches for messages in a channel matching the specified query and options.
	SearchMessages(ctx context.Context, teamRef, channelRef string, opts *models.SearchMessagesOptions) ([]*models.Message, error)
}
