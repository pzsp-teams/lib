// Package teams provides team-related operations and abstracts the underlying Microsoft Graph API calls.
//
// The package exposes two interchangeable service implementations: one without cache and one with cache.
// When cache is enabled, the service stores and reuses team references (e.g. display name -> team ID),
// reducing the number of resolver/API calls. The cache may be cleared on request errors.
//
// Concepts:
//   - teamRef is a team reference (ID or display name) used in method parameters.
//   - Operations are executed on behalf of the authenticated user (derived from MSAL); required scopes must be granted.
//   - Some operations accept a Graph patch object (msmodels.Team) for updates.
//   - Archived teams can be archived/unarchived via dedicated operations.
//   - Deleted teams can be restored using a deleted group ID.
//
// If an async cached service is used, call Wait() to ensure all background cache updates are finished.
package teams

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/models"
)

// Service defines the interface for team-related operations.
// It includes methods for retrieving, creating, updating, archiving, unarchiving, deleting, and restoring teams.
type Service interface {
	// Get retrieves a specific team by its reference (ID or display name).
	Get(ctx context.Context, teamRef string) (*models.Team, error)
	
	// ListMyJoined returns all teams the authenticated user has joined.
	ListMyJoined(ctx context.Context) ([]*models.Team, error)
	
	// Update modifies a team's properties using a Graph patch object.
	Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*models.Team, error)
	
	// CreateViaGroup creates a new team associated with a Microsoft 365 group.
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error)
	
	// CreateFromTemplate creates a new team from a template.
	CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error)
	
	// Archive archives a team, optionally making SharePoint read-only for members.
	Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error
	
	// Unarchive restores an archived team.
	Unarchive(ctx context.Context, teamRef string) error
	
	// Delete removes a team.
	Delete(ctx context.Context, teamRef string) error
	
	// RestoreDeleted restores a deleted team using the deleted group ID.
	RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error)
}
