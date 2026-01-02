package channels

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opsSUTDeps struct {
	api *testutil.MockChannelAPI
}

func newOpsSUT(t *testing.T, setup func(d opsSUTDeps)) (channelOps, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	apiMock := testutil.NewMockChannelAPI(ctrl)
	if setup != nil {
		setup(opsSUTDeps{api: apiMock})
	}

	return NewOps(apiMock), context.Background()
}

func TestOps_Wait_NoPanic(t *testing.T) {
	op, _ := newOpsSUT(t, nil)
	require.NotPanics(t, func() { op.Wait() })
}

func TestOps_ListChannelsByTeamID(t *testing.T) {
	t.Run("maps channels", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChannelCollectionResponse()
			col.SetValue([]msmodels.Channelable{
				testutil.NewGraphChannel(&testutil.NewChannelParams{
					ID:   util.Ptr("1"),
					Name: util.Ptr("General"),
				}),
				testutil.NewGraphChannel(&testutil.NewChannelParams{
					ID:   util.Ptr("2"),
					Name: util.Ptr("Random"),
				}),
			})

			d.api.EXPECT().
				ListChannels(gomock.Any(), "team-1").
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListChannelsByTeamID(ctx, "team-1")
		require.Nil(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "1", got[0].ID)
		assert.Equal(t, "General", got[0].Name)
		assert.True(t, got[0].IsGeneral)
		assert.Equal(t, "2", got[1].ID)
		assert.Equal(t, "Random", got[1].Name)
		assert.False(t, got[1].IsGeneral)
	})

	t.Run("propagates request error", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				ListChannels(gomock.Any(), "team-1").
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.ListChannelsByTeamID(ctx, "team-1")
		require.Nil(t, got)
		require.NotNil(t, err)
		assert.Equal(t, 403, err.Code)
	})
}

func TestOps_GetChannelByID(t *testing.T) {
	t.Run("maps channel", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				GetChannel(gomock.Any(), "team-1", "chan-1").
				Return(testutil.NewGraphChannel(&testutil.NewChannelParams{
					ID:   util.Ptr("chan-1"),
					Name: util.Ptr("General"),
				}), nil).
				Times(1)
		})

		got, err := op.GetChannelByID(ctx, "team-1", "chan-1")
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "chan-1", got.ID)
		assert.Equal(t, "General", got.Name)
		assert.True(t, got.IsGeneral)
	})

	t.Run("propagates request error", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				GetChannel(gomock.Any(), "team-1", "chan-1").
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.GetChannelByID(ctx, "team-1", "chan-1")
		require.Nil(t, got)
		require.NotNil(t, err)
		assert.Equal(t, 404, err.Code)
	})
}

func TestOps_CreateStandardChannel(t *testing.T) {
	t.Run("sets displayName and maps channel", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				CreateStandardChannel(gomock.Any(), "team-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, ch msmodels.Channelable) (msmodels.Channelable, *snd.RequestError) {
					dn := ch.GetDisplayName()
					require.NotNil(t, dn)
					assert.Equal(t, "MyChannel", *dn)
					return testutil.NewGraphChannel(&testutil.NewChannelParams{
						ID:   util.Ptr("c-1"),
						Name: util.Ptr("MyChannel"),
					}), nil
				}).
				Times(1)
		})

		got, err := op.CreateStandardChannel(ctx, "team-1", "MyChannel")
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "c-1", got.ID)
		assert.Equal(t, "MyChannel", got.Name)
	})

	t.Run("propagates request error", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				CreateStandardChannel(gomock.Any(), "team-1", gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.CreateStandardChannel(ctx, "team-1", "MyChannel")
		require.Nil(t, got)
		require.NotNil(t, err)
		assert.Equal(t, 403, err.Code)
	})
}

func TestOps_CreatePrivateChannel(t *testing.T) {
	t.Run("passes memberIDs/ownerIDs and maps channel", func(t *testing.T) {
		members := []string{"u1", "u2"}
		owners := []string{"o1"}

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				CreatePrivateChannelWithMembers(gomock.Any(), "team-1", "Secret", members, owners).
				Return(testutil.NewGraphChannel(&testutil.NewChannelParams{
					ID:   util.Ptr("c-1"),
					Name: util.Ptr("Secret"),
				}), nil).
				Times(1)
		})

		got, err := op.CreatePrivateChannel(ctx, "team-1", "Secret", members, owners)
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "c-1", got.ID)
		assert.Equal(t, "Secret", got.Name)
	})
}

func TestOps_DeleteChannel(t *testing.T) {
	t.Run("calls api and returns nil", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				DeleteChannel(gomock.Any(), "team-1", "chan-1").
				Return(nil).
				Times(1)
		})

		err := op.DeleteChannel(ctx, "team-1", "chan-1", "ignored-ref")
		require.Nil(t, err)
	})

	t.Run("propagates request error", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				DeleteChannel(gomock.Any(), "team-1", "chan-1").
				Return(&snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		err := op.DeleteChannel(ctx, "team-1", "chan-1", "ignored-ref")
		require.NotNil(t, err)
		assert.Equal(t, 403, err.Code)
	})
}

func TestOps_SendMessage(t *testing.T) {
	t.Run("calls api and maps message", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				SendMessage(gomock.Any(), "team-1", "chan-1", "hello", "text", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, _ string, ments []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *snd.RequestError) {
					// body has no mentions => should be empty/nil
					assert.Len(t, ments, 0)
					return testutil.NewGraphMessage(&testutil.NewMessageParams{
						ID:      util.Ptr("m-1"),
						Content: util.Ptr("hello"),
					}), nil
				}).
				Times(1)
		})

		body := models.MessageBody{Content: "hello", ContentType: models.MessageContentTypeText}
		got, err := op.SendMessage(ctx, "team-1", "chan-1", body)
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m-1", got.ID)
		assert.Equal(t, "hello", got.Content)
	})

	t.Run("propagates api error", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				SendMessage(gomock.Any(), "team-1", "chan-1", gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		body := models.MessageBody{Content: "hello", ContentType: models.MessageContentTypeText}
		got, err := op.SendMessage(ctx, "team-1", "chan-1", body)
		require.Nil(t, got)
		require.NotNil(t, err)
		assert.Equal(t, 403, err.Code)
	})

	t.Run("returns 400 when mentions cannot be prepared", func(t *testing.T) {
		// If this test fails because PrepareMentions doesn't validate this payload anymore,
		// feel free to delete/adjust it (the main behavior is covered above).
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				SendMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)
		})

		body := models.MessageBody{
			Content:     "hello",
			ContentType: models.MessageContentTypeText,
			Mentions: []models.Mention{
				{Kind: models.MentionTeam, TargetID: "not-a-guid", Text: "team", AtID: 0},
			},
		}

		got, err := op.SendMessage(ctx, "team-1", "chan-1", body)
		require.Nil(t, got)
		require.NotNil(t, err)
		assert.Equal(t, 400, err.Code)
		assert.Contains(t, err.Message, "Failed to prepare mentions")
	})
}

func TestOps_SendReply(t *testing.T) {
	t.Run("calls api and maps reply", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				SendReply(gomock.Any(), "team-1", "chan-1", "parent-1", "hi", "text", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, _ string, _ string, ments []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *snd.RequestError) {
					assert.Len(t, ments, 0)
					return testutil.NewGraphMessage(&testutil.NewMessageParams{
						ID:      util.Ptr("r-1"),
						Content: util.Ptr("hi"),
					}), nil
				}).
				Times(1)
		})

		body := models.MessageBody{Content: "hi", ContentType: models.MessageContentTypeText}
		got, err := op.SendReply(ctx, "team-1", "chan-1", "parent-1", body)
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "r-1", got.ID)
		assert.Equal(t, "hi", got.Content)
	})
}

func TestOps_ListMessages_GetMessage_ListReplies_GetReply(t *testing.T) {
	t.Run("ListMessages passes nil top when opts=nil", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{testutil.NewGraphMessage(&testutil.NewMessageParams{
				ID:      util.Ptr("m1"),
				Content: util.Ptr("m1"),
			}), testutil.NewGraphMessage(&testutil.NewMessageParams{
				ID:      util.Ptr("b"),
				Content: util.Ptr("b"),
			})})
			d.api.EXPECT().
				ListMessages(gomock.Any(), "team-1", "chan-1", nil).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListMessages(ctx, "team-1", "chan-1", nil)
		require.Nil(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "m1", got[0].ID)
		assert.Equal(t, "b", got[1].Content)
	})

	t.Run("ListMessages passes top when provided", func(t *testing.T) {
		var top int32 = 10
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{})

			d.api.EXPECT().
				ListMessages(gomock.Any(), "team-1", "chan-1", &top).
				Return(col, nil).
				Times(1)
		})

		_, err := op.ListMessages(ctx, "team-1", "chan-1", &models.ListMessagesOptions{Top: &top})
		require.Nil(t, err)
	})

	t.Run("GetMessage maps message", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				GetMessage(gomock.Any(), "team-1", "chan-1", "m-1").
				Return(testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("m-1"),
					Content: util.Ptr("hello"),
				}), nil).
				Times(1)
		})

		got, err := op.GetMessage(ctx, "team-1", "chan-1", "m-1")
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m-1", got.ID)
		assert.Equal(t, "hello", got.Content)
	})

	t.Run("ListReplies passes top", func(t *testing.T) {
		var top int32 = 5
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{testutil.NewGraphMessage(&testutil.NewMessageParams{
				ID:      util.Ptr("r1"),
				Content: util.Ptr("x"),
			})})

			d.api.EXPECT().
				ListReplies(gomock.Any(), "team-1", "chan-1", "m-1", &top).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListReplies(ctx, "team-1", "chan-1", "m-1", &models.ListMessagesOptions{Top: &top})
		require.Nil(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "r1", got[0].ID)
	})

	t.Run("GetReply maps reply", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				GetReply(gomock.Any(), "team-1", "chan-1", "m-1", "r-1").
				Return(testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("r-1"),
					Content: util.Ptr("reply"),
				}), nil).
				Times(1)
		})

		got, err := op.GetReply(ctx, "team-1", "chan-1", "m-1", "r-1")
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "r-1", got.ID)
		assert.Equal(t, "reply", got.Content)
	})
}

func TestOps_ListMembers_AddMember_UpdateMemberRoles_RemoveMember(t *testing.T) {
	t.Run("ListMembers maps members", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewConversationMemberCollectionResponse()
			col.SetValue([]msmodels.ConversationMemberable{
				testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:          util.Ptr("m1"),
					UserID:      util.Ptr("u1"),
					DisplayName: util.Ptr("Alice"),
					Roles:       util.Ptr([]string{"owner"}),
				}),
				testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:          util.Ptr("m2"),
					UserID:      util.Ptr("u2"),
					DisplayName: util.Ptr("Bob"),
					Roles:       util.Ptr([]string{}),
				}),
			})
			d.api.EXPECT().
				ListMembers(gomock.Any(), "team-1", "chan-1").
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListMembers(ctx, "team-1", "chan-1")
		require.Nil(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "m1", got[0].ID)
		assert.Equal(t, "u1", got[0].UserID)
		assert.Equal(t, "Alice", got[0].DisplayName)
		assert.Equal(t, "owner", got[0].Role)
	})

	t.Run("AddMember passes owner role", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				AddMember(gomock.Any(), "team-1", "chan-1", "user-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *snd.RequestError) {
					require.Equal(t, []string{"owner"}, roles)
					return testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("m1"),
						UserID:      util.Ptr("user-1"),
						DisplayName: util.Ptr("X"),
						Roles:       util.Ptr([]string{"owner"}),
					}), nil
				}).
				Times(1)
		})

		got, err := op.AddMember(ctx, "team-1", "chan-1", "user-1", true)
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m1", got.ID)
	})

	t.Run("UpdateMemberRoles passes empty roles when not owner", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				UpdateMemberRoles(gomock.Any(), "team-1", "chan-1", "member-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *snd.RequestError) {
					require.Len(t, roles, 0)
					return testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("member-1"),
						UserID:      util.Ptr("user-1"),
						DisplayName: util.Ptr("X"),
						Roles:       util.Ptr([]string{}),
					}), nil
				}).
				Times(1)
		})

		got, err := op.UpdateMemberRoles(ctx, "team-1", "chan-1", "member-1", false)
		require.Nil(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "member-1", got.ID)
	})

	t.Run("RemoveMember calls api", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.api.EXPECT().
				RemoveMember(gomock.Any(), "team-1", "chan-1", "member-1").
				Return(nil).
				Times(1)
		})

		err := op.RemoveMember(ctx, "team-1", "chan-1", "member-1", "ignored-user-ref")
		require.Nil(t, err)
	})
}
