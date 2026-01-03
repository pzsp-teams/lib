package mentions

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMentionAdder_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		calls []struct {
			kind     models.MentionKind
			targetID string
			text     string
		}
		want []models.Mention
	}{
		{
			name: "appends mentions and increments AtID",
			calls: []struct {
				kind     models.MentionKind
				targetID string
				text     string
			}{
				{kind: models.MentionUser, targetID: "u-1", text: "Alice"},
				{kind: models.MentionTeam, targetID: "t-1", text: "Team"},
			},
			want: []models.Mention{
				{Kind: models.MentionUser, TargetID: "u-1", Text: "Alice", AtID: 0},
				{Kind: models.MentionTeam, TargetID: "t-1", Text: "Team", AtID: 1},
			},
		},
		{
			name:  "no calls -> no output",
			calls: nil,
			want:  nil,
		},
		{
			name: "only one call -> AtID=0",
			calls: []struct {
				kind     models.MentionKind
				targetID string
				text     string
			}{
				{kind: models.MentionEveryone, targetID: "chat-1", text: "Everyone"},
			},
			want: []models.Mention{
				{Kind: models.MentionEveryone, TargetID: "chat-1", Text: "Everyone", AtID: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out []models.Mention
			a := NewMentionAdder(&out)

			for _, c := range tt.calls {
				a.Add(c.kind, c.targetID, c.text)
			}

			require.Equal(t, tt.want, out)
		})
	}
}

func TestExtractUserIDAndDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		user            msmodels.Userable
		raw             string
		wantID          string
		wantDisplayName string
		wantErrContains string
	}{
		{
			name: "ok",
			user: testutil.NewGraphUser(&testutil.NewUserParams{
				ID:          util.Ptr("id-1"),
				DisplayName: util.Ptr("Alice"),
			}),
			raw:             "alice@example.com",
			wantID:          "id-1",
			wantDisplayName: "Alice",
		},
		{
			name:            "user nil",
			user:            nil,
			raw:             "alice@example.com",
			wantErrContains: "resolved user is nil",
		},
		{
			name: "empty id",
			user: testutil.NewGraphUser(&testutil.NewUserParams{
				ID:          util.Ptr(""),
				DisplayName: util.Ptr("Alice"),
			}),
			raw:             "alice@example.com",
			wantErrContains: "resolved user has empty id",
		},
		{
			name: "id is whitespace",
			user: testutil.NewGraphUser(&testutil.NewUserParams{
				ID:          util.Ptr("   "),
				DisplayName: util.Ptr("Alice"),
			}),
			raw:             "alice@example.com",
			wantErrContains: "resolved user has empty id",
		},
		{
			name: "empty display name",
			user: testutil.NewGraphUser(&testutil.NewUserParams{
				ID:          util.Ptr("id-1"),
				DisplayName: util.Ptr(""),
			}),
			raw:             "alice@example.com",
			wantErrContains: "resolved user has empty display name",
		},
		{
			name: "display name is whitespace",
			user: testutil.NewGraphUser(&testutil.NewUserParams{
				ID:          util.Ptr("id-1"),
				DisplayName: util.Ptr("   "),
			}),
			raw:             "alice@example.com",
			wantErrContains: "resolved user has empty display name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id, dn, err := ExtractUserIDAndDisplayName(tt.user, tt.raw)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				require.Empty(t, id)
				require.Empty(t, dn)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantID, id)
			require.Equal(t, tt.wantDisplayName, dn)
		})
	}
}

func TestMentionAdder_AddUserMention(t *testing.T) {
	t.Parallel()

	type step struct {
		userRef         string
		retUser         msmodels.Userable
		retErr          *sender.RequestError
		wantErrContains string
		wantErrIs       *sender.RequestError
		wantAppend      bool
		wantTargetID    string
		wantText        string
	}

	tests := []struct {
		name  string
		steps []step
		want  []models.Mention
	}{
		{
			name: "success -> appends MentionUser, increments AtID",
			steps: []step{
				{
					userRef: "alice@example.com",
					retUser: testutil.NewGraphUser(&testutil.NewUserParams{
						ID:          util.Ptr("id-1"),
						DisplayName: util.Ptr("Alice"),
					}),
					wantAppend:   true,
					wantTargetID: "id-1",
					wantText:     "Alice",
				},
				{
					userRef: "alice@example.com",
					retUser: testutil.NewGraphUser(&testutil.NewUserParams{
						ID:          util.Ptr("id-1"),
						DisplayName: util.Ptr("Alice"),
					}),
					wantAppend:   true,
					wantTargetID: "id-1",
					wantText:     "Alice",
				},
			},
			want: []models.Mention{
				{Kind: models.MentionUser, TargetID: "id-1", Text: "Alice", AtID: 0},
				{Kind: models.MentionUser, TargetID: "id-1", Text: "Alice", AtID: 1},
			},
		},
		{
			name: "propagates UserAPI error, appends nothing",
			steps: []step{
				{
					userRef:   "alice@example.com",
					retErr:    &sender.RequestError{Message: "boom"},
					wantErrIs: &sender.RequestError{Message: "boom"},
				},
			},
			want: nil,
		},
		{
			name: "user missing displayName -> returns error, appends nothing",
			steps: []step{
				{
					userRef: "alice@example.com",
					retUser: testutil.NewGraphUser(&testutil.NewUserParams{
						ID:          util.Ptr("id-1"),
						DisplayName: util.Ptr(""),
					}),
					wantErrContains: "resolved user has empty display name",
				},
			},
			want: nil,
		},
		{
			name: "user is nil -> returns error, appends nothing",
			steps: []step{
				{
					userRef:         "alice@example.com",
					retUser:         nil,
					wantErrContains: "resolved user is nil",
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			userAPI := testutil.NewMockUserAPI(ctrl)

			var out []models.Mention
			a := NewMentionAdder(&out)

			for i, st := range tt.steps {
				var reqErrPtr *sender.RequestError
				if st.retErr != nil {
					reqErrPtr = st.retErr
				}

				userAPI.
					EXPECT().
					GetUserByEmailOrUPN(gomock.Any(), st.userRef).
					Return(st.retUser, reqErrPtr).
					Times(1)

				err := a.AddUserMention(ctx, st.userRef, userAPI)

				switch {
				case st.retErr != nil:
					require.Error(t, err, "step %d", i)
					require.True(t, errors.Is(err, reqErrPtr), "step %d: expected ErrorIs(requestErr)", i)
				case st.wantErrContains != "":
					require.Error(t, err, "step %d", i)
					require.Contains(t, err.Error(), st.wantErrContains, "step %d", i)
				default:
					require.NoError(t, err, "step %d", i)
				}

				if st.wantAppend {
					require.GreaterOrEqual(t, len(out), 1, "step %d: expected append", i)
					last := out[len(out)-1]
					require.Equal(t, models.MentionUser, last.Kind, "step %d", i)
					require.Equal(t, st.wantTargetID, last.TargetID, "step %d", i)
					require.Equal(t, st.wantText, last.Text, "step %d", i)
				}
			}
			require.Equal(t, tt.want, out)
		})
	}
}
