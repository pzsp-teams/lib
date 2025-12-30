package mentions

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type fakeUsersAPI struct {
	userByKey map[string]msmodels.Userable
	errByKey  map[string]*sender.RequestError

	lastKey string
}

func (f *fakeUsersAPI) GetUserByEmailOrUPN(ctx context.Context, emailOrUPN string) (msmodels.Userable, *sender.RequestError) {
	f.lastKey = emailOrUPN
	if f.errByKey != nil {
		if err := f.errByKey[emailOrUPN]; err != nil {
			return nil, err
		}
	}
	if f.userByKey != nil {
		if u, ok := f.userByKey[emailOrUPN]; ok {
			return u, nil
		}
	}
	return nil, &sender.RequestError{Message: "not found"}
}

var _ api.UsersAPI = (*fakeUsersAPI)(nil)

func newGraphUser(id, displayName string) msmodels.Userable {
	u := msmodels.NewUser()
	if id != "" {
		u.SetId(&id)
	}
	if displayName != "" {
		u.SetDisplayName(&displayName)
	}
	return u
}

func TestMentionAdder_Add_AppendsAndIncrementsAtID(t *testing.T) {
	out := []models.Mention{}
	a := NewMentionAdder(&out)

	a.Add(models.MentionUser, "u-1", "Alice", "user:u-1")
	a.Add(models.MentionTeam, "t-1", "Team", "team:t-1")

	if len(out) != 2 {
		t.Fatalf("expected 2 mentions, got %d", len(out))
	}

	if out[0].Kind != models.MentionUser || out[0].TargetID != "u-1" || out[0].Text != "Alice" || out[0].AtID != 0 {
		t.Fatalf("unexpected first mention: %+v", out[0])
	}
	if out[1].Kind != models.MentionTeam || out[1].TargetID != "t-1" || out[1].Text != "Team" || out[1].AtID != 1 {
		t.Fatalf("unexpected second mention: %+v", out[1])
	}
}

func TestMentionAdder_Add_DeduplicatesByKey(t *testing.T) {
	out := []models.Mention{}
	a := NewMentionAdder(&out)

	a.Add(models.MentionUser, "u-1", "Alice", "user:u-1")
	a.Add(models.MentionUser, "u-1", "Alice (dup)", "user:u-1") 

	if len(out) != 1 {
		t.Fatalf("expected 1 mention after dedup, got %d", len(out))
	}

	if out[0].Text != "Alice" {
		t.Fatalf("expected first value to stay, got %+v", out[0])
	}
	if out[0].AtID != 0 {
		t.Fatalf("expected AtID=0, got %d", out[0].AtID)
	}
}

func TestExtractUserIDAndDisplayName_OK(t *testing.T) {
	u := newGraphUser("id-1", "Alice")
	id, dn, err := ExtractUserIDAndDisplayName(u, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "id-1" || dn != "Alice" {
		t.Fatalf("unexpected values: id=%q dn=%q", id, dn)
	}
}

func TestExtractUserIDAndDisplayName_UserNil(t *testing.T) {
	_, _, err := ExtractUserIDAndDisplayName(nil, "alice@example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestExtractUserIDAndDisplayName_EmptyID(t *testing.T) {
	u := newGraphUser("", "Alice")
	_, _, err := ExtractUserIDAndDisplayName(u, "alice@example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestExtractUserIDAndDisplayName_EmptyDisplayName(t *testing.T) {
	u := newGraphUser("id-1", "")
	_, _, err := ExtractUserIDAndDisplayName(u, "alice@example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMentionAdder_AddUserMention_Success_AddsMention(t *testing.T) {
	ctx := context.Background()
	out := []models.Mention{}
	a := NewMentionAdder(&out)

	fu := &fakeUsersAPI{
		userByKey: map[string]msmodels.Userable{
			"alice@example.com": newGraphUser("id-1", "Alice"),
		},
	}

	if err := a.AddUserMention(ctx, "alice@example.com", fu); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(out))
	}

	got := out[0]
	if got.Kind != models.MentionUser {
		t.Fatalf("expected MentionUser kind, got %q", got.Kind)
	}
	if got.TargetID != "id-1" {
		t.Fatalf("expected TargetID=id-1, got %q", got.TargetID)
	}
	if got.Text != "Alice" {
		t.Fatalf("expected Text=Alice, got %q", got.Text)
	}
	if got.AtID != 0 {
		t.Fatalf("expected AtID=0, got %d", got.AtID)
	}

	if err := a.AddUserMention(ctx, "alice@example.com", fu); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected still 1 mention after second add, got %d", len(out))
	}
}

func TestMentionAdder_AddUserMention_PropagatesUsersAPIError(t *testing.T) {
	ctx := context.Background()
	out := []models.Mention{}
	a := NewMentionAdder(&out)

	reqErr := &sender.RequestError{Message: "boom"}
	fu := &fakeUsersAPI{
		errByKey: map[string]*sender.RequestError{
			"alice@example.com": reqErr,
		},
	}

	err := a.AddUserMention(ctx, "alice@example.com", fu)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, reqErr) && err.Error() != reqErr.Error() {
		t.Fatalf("expected request error, got %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no mentions appended on error, got %d", len(out))
	}
}

func TestMentionAdder_AddUserMention_UserMissingFields_ReturnsError(t *testing.T) {
	ctx := context.Background()
	out := []models.Mention{}
	a := NewMentionAdder(&out)

	fu := &fakeUsersAPI{
		userByKey: map[string]msmodels.Userable{
			"alice@example.com": newGraphUser("id-1", ""),
		},
	}

	err := a.AddUserMention(ctx, "alice@example.com", fu)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if len(out) != 0 {
		t.Fatalf("expected no mentions appended on invalid user, got %d", len(out))
	}
}
