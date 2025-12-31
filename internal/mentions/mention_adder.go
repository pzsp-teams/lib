// Package mentions provides helpers for building, validating, and mapping Microsoft Teams @mentions.
//
// It supports user mentions and conversation mentions (team/channel/everyone) by:
//   - collecting mentions with stable <at id="..."> numbering and optional de-duplication,
//   - validating HTML bodies for required <at> tags,
//   - converting internal mention models to Microsoft Graph ChatMessageMention payloads.
package mentions

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/models"
)

type MentionAdder struct {
	out      *[]models.Mention
	nextAtID int32
	seen     map[string]struct{}
}

func NewMentionAdder(out *[]models.Mention) *MentionAdder {
	return &MentionAdder{
		out:  out,
		seen: make(map[string]struct{}),
	}
}

func (a *MentionAdder) Add(kind models.MentionKind, targetID, text, dedupKey string) {
	if _, exists := a.seen[dedupKey]; exists {
		return
	}
	a.seen[dedupKey] = struct{}{}
	*a.out = append(*a.out, models.Mention{
		TargetID: targetID,
		Kind:     kind,
		AtID:     a.nextAtID,
		Text:     text,
	})
	a.nextAtID++
}

func ExtractUserIDAndDisplayName(user msmodels.Userable, raw string) (id, displayName string, err error) {
	if user == nil {
		return "", "", fmt.Errorf("resolved user is nil for mention reference: %s", raw)
	}
	idPtr := user.GetId()
	if idPtr == nil || strings.TrimSpace(*idPtr) == "" {
		return "", "", fmt.Errorf("resolved user has empty id for mention reference: %s", raw)
	}
	dnPtr := user.GetDisplayName()
	if dnPtr == nil || strings.TrimSpace(*dnPtr) == "" {
		return "", "", fmt.Errorf("resolved user has empty display name for mention reference: %s", raw)
	}
	return *idPtr, *dnPtr, nil
}

func (a *MentionAdder) AddUserMention(ctx context.Context, email string, userAPI api.UsersAPI) error {
	user, reqErr := userAPI.GetUserByEmailOrUPN(ctx, email)
	if reqErr != nil {
		return reqErr
	}

	id, dn, err := ExtractUserIDAndDisplayName(user, email)
	if err != nil {
		return err
	}

	a.Add(models.MentionUser, id, dn, "user:"+id)
	return nil
}
