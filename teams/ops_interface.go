package teams

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

type teamsOps interface {
	GetTeamByID(ctx context.Context, teamID string) (*models.Team, error)
	ListMyJoinedTeams(ctx context.Context) ([]*models.Team, error)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error)
	CreateFromTemplate(ctx context.Context, displayName, description string, ownerIDs []string) (string, error)
	Archive(ctx context.Context, teamID, teamRef string, spoReadOnlyForMembers *bool) error
	Unarchive(ctx context.Context, teamID string) error
	DeleteTeam(ctx context.Context, teamID, teamRef string) error
	RestoreDeletedTeam(ctx context.Context, deletedGroupID string) (string, error)
	ListMembers(ctx context.Context, teamID string) ([]*models.Member, error)
	GetMemberByID(ctx context.Context, teamID, memberID string) (*models.Member, error)
	AddMember(ctx context.Context, teamID, userRef string, isOwner bool) (*models.Member, error)
	UpdateMemberRoles(ctx context.Context, teamID, memberID string, isOwner bool) (*models.Member, error)
	RemoveMember(ctx context.Context, teamID, memberID, userRef string) error
}
