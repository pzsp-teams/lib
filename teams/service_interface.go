package teams

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/models"
)

type Service interface {
	Get(ctx context.Context, teamRef string) (*models.Team, error)
	ListMyJoined(ctx context.Context) ([]*models.Team, error)
	Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*models.Team, error)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error)
	CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error)
	Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error
	Unarchive(ctx context.Context, teamRef string) error
	Delete(ctx context.Context, teamRef string) error
	RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error)
}
