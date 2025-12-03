package teams

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type TeamsAPIInterface interface {
	CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError)
	Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError)
	ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError)
	Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError)
	Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError
	Unarchive(ctx context.Context, teamID string) *sender.RequestError
	Delete(ctx context.Context, teamID string) *sender.RequestError
	RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError)
}
