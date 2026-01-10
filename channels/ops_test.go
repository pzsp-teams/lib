package channels

import (
	"context"
	"errors"
	"net/http"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	iapi "github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/search"
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
		require.Len(t, got.Messages, 2)
		assert.Equal(t, "m1", got.Messages[0].ID)
		assert.Equal(t, "b", got.Messages[1].Content)
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

	t.Run("sets NextLink from graph response", func(t *testing.T) {
		next := "https://graph.microsoft.com/v1.0/next"
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("m1"), Content: util.Ptr("a")}),
			})
			col.SetOdataNextLink(&next)

			d.channelAPI.EXPECT().
				ListMessages(gomock.Any(), "team-1", "chan-1", nil, false).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListMessages(ctx, "team-1", "chan-1", nil, false)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.NextLink)
		assert.Equal(t, next, *got.NextLink)
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
		require.Len(t, got.Messages, 1)
		assert.Equal(t, "r1", got.Messages[0].ID)
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

	t.Run("sets NextLink from graph response", func(t *testing.T) {
		next := "https://graph.microsoft.com/v1.0/next"
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("r1"), Content: util.Ptr("x")}),
			})
			col.SetOdataNextLink(&next)

			d.channelAPI.EXPECT().
				ListReplies(gomock.Any(), "team-1", "chan-1", "m-1", nil, false).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListReplies(ctx, "team-1", "chan-1", "m-1", nil, false)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.NextLink)
		assert.Equal(t, next, *got.NextLink)
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

func TestOps_ListMessagesNext(t *testing.T) {
	t.Run("passes nextLink and maps messages + nextLink", func(t *testing.T) {
		nextIn := "https://graph.microsoft.com/v1.0/whatever?$skiptoken=abc"
		nextOut := "https://graph.microsoft.com/v1.0/whatever?$skiptoken=def"

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("m1"), Content: util.Ptr("a")}),
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("m2"), Content: util.Ptr("b")}),
			})
			col.SetOdataNextLink(&nextOut)

			d.channelAPI.EXPECT().
				ListMessagesNext(gomock.Any(), "team-1", "chan-1", nextIn, false).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListMessagesNext(ctx, "team-1", "chan-1", nextIn, false)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.Messages, 2)
		assert.Equal(t, "m1", got.Messages[0].ID)
		assert.Equal(t, "b", got.Messages[1].Content)
		require.NotNil(t, got.NextLink)
		assert.Equal(t, nextOut, *got.NextLink)
	})

	t.Run("passes includeSystem through", func(t *testing.T) {
		nextIn := "next"
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{})

			d.channelAPI.EXPECT().
				ListMessagesNext(gomock.Any(), "team-1", "chan-1", nextIn, true).
				Return(col, nil).
				Times(1)
		})

		_, err := op.ListMessagesNext(ctx, "team-1", "chan-1", nextIn, true)
		require.NoError(t, err)
	})

	t.Run("maps api error via sender", func(t *testing.T) {
		nextIn := "next"
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				ListMessagesNext(gomock.Any(), "team-1", "chan-1", nextIn, false).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		got, err := op.ListMessagesNext(ctx, "team-1", "chan-1", nextIn, false)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")

		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})
}

func TestOps_ListRepliesNext(t *testing.T) {
	t.Run("passes nextLink and maps replies + nextLink", func(t *testing.T) {
		nextIn := "https://graph.microsoft.com/v1.0/replies?$skiptoken=abc"
		nextOut := "https://graph.microsoft.com/v1.0/replies?$skiptoken=def"

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{
				testutil.NewGraphMessage(&testutil.NewMessageParams{ID: util.Ptr("r1"), Content: util.Ptr("x")}),
			})
			col.SetOdataNextLink(&nextOut)

			d.channelAPI.EXPECT().
				ListRepliesNext(gomock.Any(), "team-1", "chan-1", "m-1", nextIn, false).
				Return(col, nil).
				Times(1)
		})

		got, err := op.ListRepliesNext(ctx, "team-1", "chan-1", "m-1", nextIn, false)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.Messages, 1)
		assert.Equal(t, "r1", got.Messages[0].ID)
		require.NotNil(t, got.NextLink)
		assert.Equal(t, nextOut, *got.NextLink)
	})

	t.Run("passes includeSystem through", func(t *testing.T) {
		nextIn := "next"
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			col := msmodels.NewChatMessageCollectionResponse()
			col.SetValue([]msmodels.ChatMessageable{})

			d.channelAPI.EXPECT().
				ListRepliesNext(gomock.Any(), "team-1", "chan-1", "m-1", nextIn, true).
				Return(col, nil).
				Times(1)
		})

		_, err := op.ListRepliesNext(ctx, "team-1", "chan-1", "m-1", nextIn, true)
		require.NoError(t, err)
	})

	t.Run("maps api error via sender (includes message id)", func(t *testing.T) {
		nextIn := "next"
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				ListRepliesNext(gomock.Any(), "team-1", "chan-1", "m-1", nextIn, false).
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
				Times(1)
		})

		got, err := op.ListRepliesNext(ctx, "team-1", "chan-1", "m-1", nextIn, false)
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

func TestOps_SearchChannelMessages(t *testing.T) {
	t.Run("opts nil -> returns error and does not call api", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SearchChannelMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)
		})

		got, err := op.SearchChannelMessages(ctx, util.Ptr("team-1"), util.Ptr("chan-1"), nil, nil)
		require.Nil(t, got)
		require.Error(t, err)
		require.Equal(t, "missing opts", err.Error())
	})

	t.Run("maps request error without resources when teamID/channelID are nil", func(t *testing.T) {
		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SearchChannelMessages(gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
				Return(nil, &snd.RequestError{Code: 403, Message: "nope"}, nil).
				Times(1)
		})

		opts := &search.SearchMessagesOptions{}
		got, err := op.SearchChannelMessages(ctx, nil, nil, opts, nil)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 403)
		var af *snd.ErrAccessForbidden
		require.ErrorAs(t, err, &af)
	})

	t.Run("maps request error with team/channel resources when IDs present", func(t *testing.T) {
		teamID := util.Ptr("team-1")
		channelID := util.Ptr("chan-1")

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SearchChannelMessages(gomock.Any(), teamID, channelID, gomock.Any(), gomock.Any()).
				Return(nil, &snd.RequestError{Code: 404, Message: "missing"}, nil).
				Times(1)
		})

		opts := &search.SearchMessagesOptions{}
		got, err := op.SearchChannelMessages(ctx, teamID, channelID, opts, nil)
		require.Nil(t, got)
		require.Error(t, err)

		requireStatus(t, err, 404)
		requireErrDataHas(t, err, resources.Team, "team-1")
		requireErrDataHas(t, err, resources.Channel, "chan-1")

		var nf *snd.ErrResourceNotFound
		require.ErrorAs(t, err, &nf)
	})

	t.Run("success maps messages and nextFrom", func(t *testing.T) {
		teamID := util.Ptr("team-1")
		channelID := util.Ptr("chan-1")
		chatID := util.Ptr("chat-1")
		next := int32(42)

		apiResp := []*iapi.SearchMessage{
			{
				Message: testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("m1"),
					Content: util.Ptr("hello"),
				}),
				TeamID:    teamID,
				ChannelID: channelID,
				ChatID:    chatID,
			},
			{
				Message: testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("m2"),
					Content: util.Ptr("world"),
				}),
				TeamID:    teamID,
				ChannelID: channelID,
				ChatID:    nil,
			},
		}

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().
				SearchChannelMessages(gomock.Any(), teamID, channelID, gomock.Any(), gomock.Any()).
				Return(apiResp, nil, &next).
				Times(1)
		})

		opts := &search.SearchMessagesOptions{Query: util.Ptr("q")}
		got, err := op.SearchChannelMessages(ctx, teamID, channelID, opts, nil)
		require.NoError(t, err)
		require.NotNil(t, got)

		require.NotNil(t, got.NextFrom)
		require.Equal(t, int32(42), *got.NextFrom)

		require.Len(t, got.Messages, 2)

		require.NotNil(t, got.Messages[0].Message)
		assert.Equal(t, "m1", got.Messages[0].Message.ID)
		assert.Equal(t, "hello", got.Messages[0].Message.Content)
		assert.Equal(t, teamID, got.Messages[0].TeamID)
		assert.Equal(t, channelID, got.Messages[0].ChannelID)
		assert.Equal(t, chatID, got.Messages[0].ChatID)

		require.NotNil(t, got.Messages[1].Message)
		assert.Equal(t, "m2", got.Messages[1].Message.ID)
		assert.Equal(t, "world", got.Messages[1].Message.Content)
		assert.Equal(t, teamID, got.Messages[1].TeamID)
		assert.Equal(t, channelID, got.Messages[1].ChannelID)
		assert.Nil(t, got.Messages[1].ChatID)
	})

	t.Run("success allows nil nextFrom", func(t *testing.T) {
		teamID := util.Ptr("team-1")
		channelID := util.Ptr("chan-1")

		apiResp := []*iapi.SearchMessage{
			{
				Message: testutil.NewGraphMessage(&testutil.NewMessageParams{
					ID:      util.Ptr("m1"),
					Content: util.Ptr("x"),
				}),
				TeamID:    teamID,
				ChannelID: channelID,
			},
		}

		op, ctx := newOpsSUT(t, func(d opsSUTDeps) {
			d.channelAPI.EXPECT().	
				SearchChannelMessages(gomock.Any(), teamID, channelID, gomock.Any(), gomock.Any()).
				Return(apiResp, nil, nil).
				Times(1)
		})

		opts := &search.SearchMessagesOptions{}
		got, err := op.SearchChannelMessages(ctx, teamID, channelID, opts, nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Nil(t, got.NextFrom)
		require.Len(t, got.Messages, 1)
	})
}

func TestOpsWithCache_SearchChannelMessages(t *testing.T) {
	teamID := util.Ptr("team-1")
	channelID := util.Ptr("chan-1")
	opts := &search.SearchMessagesOptions{Query: util.Ptr("hello")}

	t.Run("success passes through result and does not touch cache", func(t *testing.T) {
		var next int32 = 123
		want := &search.SearchResults{
			Messages: []*search.SearchResult{
				{Message: &models.Message{ID: "m1"}},
			},
			NextFrom: &next,
		}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				SearchChannelMessages(gomock.Any(), teamID, channelID, opts, nil).
				Return(want, nil).
				Times(1)

			d.runner.EXPECT().Run(gomock.Any()).Times(0)
			d.cacher.EXPECT().Clear().Times(0)
			d.cacher.EXPECT().Set(gomock.Any(), gomock.Any()).Times(0)
			d.cacher.EXPECT().Invalidate(gomock.Any()).Times(0)
		})

			got, err := sut.SearchChannelMessages(ctx, teamID, channelID, opts, nil)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error clears cache (WithErrorClear)", func(t *testing.T) {
		err400 := testutil.ReqErr(http.StatusBadRequest)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				SearchChannelMessages(gomock.Any(), teamID, channelID, opts, nil).
				Return(nil, err400).
				Times(1)

			expectClearNow(d)
		})

		got, err := sut.SearchChannelMessages(ctx, teamID, channelID, opts, nil)
		require.Nil(t, got)
		require.Error(t, err)
		require.True(t, err == err400)
	})
}

func TestOpsWithCache_GetChannelByID_BlankName_NoSet(t *testing.T) {
	teamID := "team-1"
	channelID := "chan-1"

	t.Run("channel returned but blank name -> task runs, helper skips Set", func(t *testing.T) {
		out := &models.Channel{ID: "c1", Name: "   "}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				GetChannelByID(gomock.Any(), teamID, channelID).
				Return(out, nil).
				Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(gomock.Any(), gomock.Any()).Times(0)
		})

		got, err := sut.GetChannelByID(ctx, teamID, channelID)
		require.NoError(t, err)
		require.Equal(t, out, got)
	})
}

func TestOpsWithCache_AddMember_BlankEmail_NoSet(t *testing.T) {
	teamID := "team-1"
	channelID := "chan-1"

	t.Run("member returned but blank email -> task runs, helper skips Set", func(t *testing.T) {
		out := &models.Member{ID: "m1", Email: "   "}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				AddMember(gomock.Any(), teamID, channelID, "u1", false).
				Return(out, nil).
				Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(gomock.Any(), gomock.Any()).Times(0)
		})

		got, err := sut.AddMember(ctx, teamID, channelID, "u1", false)
		require.NoError(t, err)
		require.Equal(t, out, got)
	})
}
