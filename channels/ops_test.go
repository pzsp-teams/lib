package channels

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type opsSUTDeps struct {
	channelAPI *testutil.MockChannelAPI
	userAPI    *testutil.MockUserAPI
}

func newOpsSUT(t *testing.T, setup func(d opsSUTDeps)) (channelOps, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	apiMock := testutil.NewMockChannelAPI(ctrl)
	userMock := testutil.NewMockUserAPI(ctrl)

	if setup != nil {
		setup(opsSUTDeps{channelAPI: apiMock, userAPI: userMock})
	}

	return NewOps(apiMock, userMock), context.Background()
}

func requireStatus(t *testing.T, err error, want int) {
	t.Helper()
	got, ok := snd.StatusCode(err)
	require.True(t, ok, "expected StatusCode() available for error: %T (%v)", err, err)
	require.Equal(t, want, got)
}

func requireErrDataHas(t *testing.T, err error, res resources.Resource, want string) {
	t.Helper()
	var afp *snd.ErrAccessForbidden
	if errors.As(err, &afp) {
		require.Contains(t, afp.ResourceRefs[res], want)
		return
	}
	var nfp *snd.ErrResourceNotFound
	if errors.As(err, &nfp) {
		require.Contains(t, nfp.ResourceRefs[res], want)
		return
	}

	var afv snd.ErrAccessForbidden
	if errors.As(err, &afv) {
		require.Contains(t, afv.ResourceRefs[res], want)
		return
	}
	var nfv snd.ErrResourceNotFound
	if errors.As(err, &nfv) {
		require.Contains(t, nfv.ResourceRefs[res], want)
		return
	}

	t.Fatalf("could not extract ErrData from error type %T (%v)", err, err)
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

			d.channelAPI.EXPECT().
				ListChannels(gomock.Any(), "team-1").
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListChannelsByTeamID(ctx, "team-1")
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "1", got[0].ID)
		assert.Equal(t, "General", got[0].Name)
		assert.True(t, got[0].IsGeneral)
		assert.Equal(t, "2", got[1].ID)
		assert.Equal(t, "Random", got[1].Name)
		assert.False(t, got[1].IsGeneral)
	})

	t.Run("maps request error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				ListChannels(gomock.Any(), "team-1").
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.ListChannelsByTeamID(ctx, "team-1")
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
		assert.Equal(t, 403, af.Code)
	})
}

func TestOps_GetChannelByID(t *testing.T) {
	t.Run("maps channel", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				GetChannel(gomock.Any(), "team-1", "chan-1").
				Return(testutil.NewGraphChannel(&testutil.NewChannelParams{
					ID:   util.Ptr("chan-1"),
					Name: util.Ptr("General"),
				}), nil).
				Times(1)
		})

		got, err := op.GetChannelByID(ctx, "team-1", "chan-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "chan-1", got.ID)
		assert.Equal(t, "General", got.Name)
		assert.True(t, got.IsGeneral)
	})

	t.Run("maps request error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				GetChannel(gomock.Any(), "team-1", "chan-1").
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.GetChannelByID(ctx, "team-1", "chan-1")
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
		assert.Equal(t, 404, nf.Code)
	})
}

func TestOps_CreateStandardChannel(t *testing.T) {
	t.Run("sets displayName and maps channel", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
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
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "c-1", got.ID)
		assert.Equal(t, "MyChannel", got.Name)
	})

	t.Run("maps request error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				CreateStandardChannel(gomock.Any(), "team-1", gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.CreateStandardChannel(ctx, "team-1", "MyChannel")
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "MyChannel")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_CreatePrivateChannel(t *testing.T) {
	t.Run("passes memberIDs/ownerIDs and maps channel", func(t *testing.T) {
		members := []string{"u1", "u2"}
		owners := []string{"o1"}

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				CreatePrivateChannelWithMembers(gomock.Any(), "team-1", "Secret", members, owners).
				Return(testutil.NewGraphChannel(&testutil.NewChannelParams{
					ID:   util.Ptr("c-1"),
					Name: util.Ptr("Secret"),
				}), nil).
				Times(1)
		})

		got, err := op.CreatePrivateChannel(ctx, "team-1", "Secret", members, owners)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "c-1", got.ID)
		assert.Equal(t, "Secret", got.Name)
	})

	t.Run("maps request error via sender", func(t *testing.T) {
		members := []string{"u1"}
		owners := []string{"o1"}

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				CreatePrivateChannelWithMembers(gomock.Any(), "team-1", "Secret", members, owners).
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.CreatePrivateChannel(ctx, "team-1", "Secret", members, owners)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "Secret")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
	})
}

func TestOps_DeleteChannel(t *testing.T) {
	t.Run("calls api and returns nil", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				DeleteChannel(gomock.Any(), "team-1", "chan-1").
				Return(nil).
				Times(1)
		})

		err := op.DeleteChannel(ctx, "team-1", "chan-1", "ignored-ref")
		require.NoError(t, err)
	})

	t.Run("maps request error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				DeleteChannel(gomock.Any(), "team-1", "chan-1").
				Return(&snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		err := op.DeleteChannel(ctx, "team-1", "chan-1", "ignored-ref")
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "ignored-ref")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_SendMessage(t *testing.T) {
	t.Run("calls api and maps message", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SendMessage(gomock.Any(), "team-1", "chan-1", "hello", "text", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, _ string, ments []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *snd.RequestError) {
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
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m-1", got.ID)
		assert.Equal(t, "hello", got.Content)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SendMessage(gomock.Any(), "team-1", "chan-1", gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		body := models.MessageBody{Content: "hello", ContentType: models.MessageContentTypeText}
		got, err := op.SendMessage(ctx, "team-1", "chan-1", body)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})

	t.Run("returns 400 when mentions cannot be prepared", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
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
		require.Error(t, err)

		requireStatus(t, err, 400)

		var re *snd.RequestError
		require.ErrorAs(t, err, &re)
		assert.Equal(t, 400, re.Code)
		assert.Contains(t, re.Message, "Failed to prepare mentions")
	})
}

func TestOps_SendReply(t *testing.T) {
	t.Run("calls api and maps reply", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
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
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "r-1", got.ID)
		assert.Equal(t, "hi", got.Content)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SendReply(gomock.Any(), "team-1", "chan-1", "parent-1", gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		body := models.MessageBody{Content: "hi", ContentType: models.MessageContentTypeText}
		got, err := op.SendReply(ctx, "team-1", "chan-1", "parent-1", body)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.Message, "parent-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})

	t.Run("returns 400 when mentions cannot be prepared", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SendReply(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)
		})

		body := models.MessageBody{
			Content:     "hi",
			ContentType: models.MessageContentTypeText,
			Mentions: []models.Mention{
				{Kind: models.MentionTeam, TargetID: "not-a-guid", Text: "team", AtID: 0},
			},
		}

		got, err := op.SendReply(ctx, "team-1", "chan-1", "parent-1", body)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 400)

		var re *snd.RequestError
		require.ErrorAs(t, err, &re)
		assert.Contains(t, re.Message, "Failed to prepare mentions")
	})
}

func TestOps_ListMessages(t *testing.T) {
	t.Run("passes nil top when opts=nil and maps messages", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("m1"), Content: util.Ptr("a")}),
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("m2"), Content: util.Ptr("b")}),
			})
			d.channelAPI.EXPECT().
				ListMessages(gomock.Any(), "team-1", "chan-1", nil, false).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListMessages(ctx, "team-1", "chan-1", nil, false)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "m1", got[0].ID)
		assert.Equal(t, "b", got[1].Content)
	})

	t.Run("passes top when provided", func(t *testing.T) {
		var top int32 = 10
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{})
			d.channelAPI.EXPECT().
				ListMessages(gomock.Any(), "team-1", "chan-1", &top, false).
				Return(col, nil).
				Times(1)
		})

		_, err := op.ListMessages(ctx, "team-1", "chan-1", &models.ListMessagesOptions{Top: &top}, false)
		require.NoError(t, err)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		var top int32 = 5
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				ListMessages(gomock.Any(), "team-1", "chan-1", &top, false).
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.ListMessages(ctx, "team-1", "chan-1", &models.ListMessagesOptions{Top: &top}, false)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
	})
}

func TestOps_GetMessage(t *testing.T) {
	t.Run("maps message", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				GetMessage(gomock.Any(), "team-1", "chan-1", "m-1").
				Return(testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("m-1"),
					Content: util.Ptr("hello"),
				}), nil).
				Times(1)
		})

		got, err := op.GetMessage(ctx, "team-1", "chan-1", "m-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m-1", got.ID)
		assert.Equal(t, "hello", got.Content)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				GetMessage(gomock.Any(), "team-1", "chan-1", "m-1").
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.GetMessage(ctx, "team-1", "chan-1", "m-1")
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.Message, "m-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_ListReplies(t *testing.T) {
	t.Run("passes nil top when opts=nil and maps replies", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("r1"), Content: util.Ptr("x")}),
			})
			d.channelAPI.EXPECT().
				ListReplies(gomock.Any(), "team-1", "chan-1", "m-1", nil, false).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListReplies(ctx, "team-1", "chan-1", "m-1", nil, false)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "r1", got[0].ID)
	})

	t.Run("passes top when provided", func(t *testing.T) {
		var top int32 = 5
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{})
			d.channelAPI.EXPECT().
				ListReplies(gomock.Any(), "team-1", "chan-1", "m-1", &top, false).
				Return(col, nil).
				Times(1)
		})

		_, err := op.ListReplies(ctx, "team-1", "chan-1", "m-1", &models.ListMessagesOptions{Top: &top}, false)
		require.NoError(t, err)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		var top int32 = 5
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				ListReplies(gomock.Any(), "team-1", "chan-1", "m-1", &top, false).
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.ListReplies(ctx, "team-1", "chan-1", "m-1", &models.ListMessagesOptions{Top: &top}, false)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.Message, "m-1")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
	})
}

func TestOps_GetReply(t *testing.T) {
	t.Run("maps reply", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				GetReply(gomock.Any(), "team-1", "chan-1", "m-1", "r-1").
				Return(testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("r-1"),
					Content: util.Ptr("reply"),
				}), nil).
				Times(1)
		})

		got, err := op.GetReply(ctx, "team-1", "chan-1", "m-1", "r-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "r-1", got.ID)
		assert.Equal(t, "reply", got.Content)
	})

	t.Run("maps api error via sender (includes reply id in errdata)", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				GetReply(gomock.Any(), "team-1", "chan-1", "m-1", "r-1").
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.GetReply(ctx, "team-1", "chan-1", "m-1", "r-1")
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.Message, "m-1")
		requireErrDataHas(t, err, resources.Message, "r-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_ListMembers(t *testing.T) {
	t.Run("maps members", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewConversationMemberCollectionResponse()
			col.SetValue([]msmodels.ConversationMemberable{
				testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:          util.Ptr("m1"),
					UserID:      util.Ptr("u1"),
					DisplayName: util.Ptr("Alice"),
					Roles:       []string{"owner"},
				}),
			})
			d.channelAPI.EXPECT().
				ListMembers(gomock.Any(), "team-1", "chan-1").
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListMembers(ctx, "team-1", "chan-1")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "m1", got[0].ID)
		assert.Equal(t, "u1", got[0].UserID)
		assert.Equal(t, "Alice", got[0].DisplayName)
		assert.Equal(t, "owner", got[0].Role)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				ListMembers(gomock.Any(), "team-1", "chan-1").
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.ListMembers(ctx, "team-1", "chan-1")
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
	})
}

func TestOps_AddMember(t *testing.T) {
	t.Run("passes owner role when isOwner=true", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				AddMember(gomock.Any(), "team-1", "chan-1", "user-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *snd.RequestError) {
					require.Equal(t, []string{"owner"}, roles)
					return testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("m1"),
						UserID:      util.Ptr("user-1"),
						DisplayName: util.Ptr("X"),
						Roles:       []string{"owner"},
					}), nil
				}).
				Times(1)
		})

		got, err := op.AddMember(ctx, "team-1", "chan-1", "user-1", true)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("passes empty roles when isOwner=false", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				AddMember(gomock.Any(), "team-1", "chan-1", "user-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *snd.RequestError) {
					require.Len(t, roles, 0)
					return testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("m1"),
						UserID:      util.Ptr("user-1"),
						DisplayName: util.Ptr("X"),
						Roles:       []string{},
					}), nil
				}).
				Times(1)
		})

		_, err := op.AddMember(ctx, "team-1", "chan-1", "user-1", false)
		require.NoError(t, err)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				AddMember(gomock.Any(), "team-1", "chan-1", "user-1", gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.AddMember(ctx, "team-1", "chan-1", "user-1", true)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.User, "user-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_UpdateMemberRoles(t *testing.T) {
	t.Run("passes owner role when isOwner=true", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				UpdateMemberRoles(gomock.Any(), "team-1", "chan-1", "member-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *snd.RequestError) {
					require.Equal(t, []string{"owner"}, roles)
					return testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("member-1"),
						UserID:      util.Ptr("user-1"),
						DisplayName: util.Ptr("X"),
						Roles:       []string{"owner"},
					}), nil
				}).
				Times(1)
		})

		_, err := op.UpdateMemberRoles(ctx, "team-1", "chan-1", "member-1", true)
		require.NoError(t, err)
	})

	t.Run("passes empty roles when isOwner=false", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				UpdateMemberRoles(gomock.Any(), "team-1", "chan-1", "member-1", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *snd.RequestError) {
					require.Len(t, roles, 0)
					return testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("member-1"),
						UserID:      util.Ptr("user-1"),
						DisplayName: util.Ptr("X"),
						Roles:       []string{},
					}), nil
				}).
				Times(1)
		})

		_, err := op.UpdateMemberRoles(ctx, "team-1", "chan-1", "member-1", false)
		require.NoError(t, err)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				UpdateMemberRoles(gomock.Any(), "team-1", "chan-1", "member-1", gomock.Any()).
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.UpdateMemberRoles(ctx, "team-1", "chan-1", "member-1", true)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.User, "member-1")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
	})
}

func TestOps_RemoveMember(t *testing.T) {
	t.Run("calls api and returns nil", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				RemoveMember(gomock.Any(), "team-1", "chan-1", "member-1").
				Return(nil).
				Times(1)
		})

		err := op.RemoveMember(ctx, "team-1", "chan-1", "member-1", "ignored-user-ref")
		require.NoError(t, err)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				RemoveMember(gomock.Any(), "team-1", "chan-1", "member-1").
				Return(&snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		err := op.RemoveMember(ctx, "team-1", "chan-1", "member-1", "ignored-user-ref")
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")
		requireErrDataHas(t, err, resources.User, "member-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_GetMentions(t *testing.T) {
	t.Run("returns error for unknown mention reference", func(t *testing.T) {
		op, ctx := newOpsSUT(t, nil)
		_, err := op.GetMentions(ctx, "teamID", "teamRef", "chanRef", "chanID", []string{"???", "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot resolve mention reference")
	})

	t.Run("skips empty/whitespace mentions", func(t *testing.T) {
		op, ctx := newOpsSUT(t, nil)
		got, err := op.GetMentions(ctx, "teamID", "teamRef", "chanRef", "chanID", []string{" ", "", "\t"})
		require.NoError(t, err)
		require.Len(t, got, 0)
	})
}
