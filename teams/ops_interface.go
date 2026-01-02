package teams

import (
	"context"

	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type teamsOps interface {
	Wait()
	GetTeamByID(ctx context.Context, teamID string) (*models.Team, *snd.RequestError)
	ListMyJoinedTeams(ctx context.Context) ([]*models.Team, *snd.RequestError)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, *snd.RequestError)
	CreateFromTemplate(ctx context.Context, displayName, description string, ownerIDs []string) (string, *snd.RequestError)
	Archive(ctx context.Context, teamID string) *snd.RequestError
	Unarchive(ctx context.Context, teamID string) *snd.RequestError
	DeleteTeam(ctx context.Context, teamID, teamRef string) *snd.RequestError
	RestoreDeletedTeam(ctx context.Context, teamID string) (string, *snd.RequestError)
	ListMembers(ctx context.Context, teamID string) ([]*models.Member, *snd.RequestError)
	GetMemberByID(ctx context.Context, teamID, memberID string) (*models.Member, *snd.RequestError)
	AddMember(ctx context.Context, teamID, userRef string, isOwner bool) (*models.Member, *snd.RequestError)
	UpdateMemberRoles(ctx context.Context, teamID, memberID string, isOwner bool) (*models.Member, *snd.RequestError)
	RemoveMember(ctx context.Context, teamID, memberID, userRef string) *snd.RequestError
}
