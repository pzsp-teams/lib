package api

import (
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/require"
)

type normEntity struct {
	msgID     string
	teamID    string
	channelID string
	chatID    string
}

func normEntities(in []SearchEntity) []normEntity {
	out := make([]normEntity, 0, len(in))
	for _, e := range in {
		out = append(out, normEntity{
			msgID:     util.Deref(e.MessageID),
			teamID:    util.Deref(e.TeamID),
			channelID: util.Deref(e.ChannelID),
			chatID:    util.Deref(e.ChatID),
		})
	}
	return out
}

func newResource(id string, additional map[string]any) msmodels.Entityable {
	m := msmodels.NewChatMessage()
	m.SetId(&id)
	if additional != nil {
		m.SetAdditionalData(additional)
	}
	return m
}

func newHit(res msmodels.Entityable) msmodels.SearchHitable {
	h := msmodels.NewSearchHit()
	h.SetResource(res)
	return h
}

func newHitsContainer(hits ...msmodels.SearchHitable) msmodels.SearchHitsContainerable {
	hc := msmodels.NewSearchHitsContainer()
	hc.SetHits(hits)
	return hc
}

func newSearchResponse(containers ...msmodels.SearchHitsContainerable) msmodels.SearchResponseable {
	sr := msmodels.NewSearchResponse()
	sr.SetHitsContainers(containers)
	return sr
}

func newQueryPostResponse(responses ...msmodels.SearchResponseable) graphsearch.QueryPostResponseable {
	resp := graphsearch.NewQueryPostResponse()
	resp.SetValue(responses)
	return resp
}

func TestAsStringPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want *string
	}{
		{name: "nil -> nil", in: nil, want: nil},
		{name: "*string nil -> nil", in: (*string)(nil), want: nil},
		{name: "*string blank -> nil", in: util.Ptr("   "), want: nil},
		{name: "*string trims", in: util.Ptr("  abc  "), want: util.Ptr("abc")},
		{name: "string blank -> nil", in: "  ", want: nil},
		{name: "string trims", in: "  xyz ", want: util.Ptr("xyz")},
		{name: "unsupported type -> nil", in: 123, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := asStringPtr(tt.in)

			if tt.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tt.want, *got)
		})
	}
}

func TestPrepareIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		res     msmodels.Entityable
		wantMsg *string
		wantT   *string
		wantC   *string
		wantCh  *string
	}{
		{
			name:    "nil resource -> all nil",
			res:     nil,
			wantMsg: nil, wantT: nil, wantC: nil, wantCh: nil,
		},
		{
			name:    "nil id -> all nil",
			res:     func() msmodels.Entityable { m := msmodels.NewChatMessage(); return m }(),
			wantMsg: nil, wantT: nil, wantC: nil, wantCh: nil,
		},
		{
			name:    "blank id -> all nil",
			res:     newResource("   ", nil),
			wantMsg: nil, wantT: nil, wantC: nil, wantCh: nil,
		},
		{
			name:    "valid id + no additionalData -> only msgID set",
			res:     newResource("m1", nil),
			wantMsg: util.Ptr("m1"), wantT: nil, wantC: nil, wantCh: nil,
		},
		{
			name: "valid id + channelIdentity map trims + chatId trims",
			res: newResource("m2", map[string]any{
				"channelIdentity": map[string]any{
					"teamId":    "  t1  ",
					"channelId": util.Ptr("  c1 "),
				},
				"chatId": "  ch1 ",
			}),
			wantMsg: util.Ptr("m2"),
			wantT:   util.Ptr("t1"),
			wantC:   util.Ptr("c1"),
			wantCh:  util.Ptr("ch1"),
		},
		{
			name: "channelIdentity present but wrong type -> ignored",
			res: newResource("m3", map[string]any{
				"channelIdentity": "not-a-map",
				"chatId":          "chat-x",
			}),
			wantMsg: util.Ptr("m3"),
			wantT:   nil,
			wantC:   nil,
			wantCh:  util.Ptr("chat-x"),
		},
		{
			name: "chatId blank -> nil",
			res: newResource("m4", map[string]any{
				"chatId": "   ",
			}),
			wantMsg: util.Ptr("m4"),
			wantT:   nil,
			wantC:   nil,
			wantCh:  nil,
		},
		{
			name: "channelIdentity map but blanks -> nil ids",
			res: newResource("m5", map[string]any{
				"channelIdentity": map[string]any{
					"teamId":    " ",
					"channelId": "",
				},
			}),
			wantMsg: util.Ptr("m5"),
			wantT:   nil,
			wantC:   nil,
			wantCh:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			teamID, channelID, chatID, msgID := prepareIDs(tt.res)

			if tt.wantMsg == nil {
				require.Nil(t, msgID)
			} else {
				require.NotNil(t, msgID)
				require.Equal(t, *tt.wantMsg, *msgID)
			}

			if tt.wantT == nil {
				require.Nil(t, teamID)
			} else {
				require.NotNil(t, teamID)
				require.Equal(t, *tt.wantT, *teamID)
			}

			if tt.wantC == nil {
				require.Nil(t, channelID)
			} else {
				require.NotNil(t, channelID)
				require.Equal(t, *tt.wantC, *channelID)
			}

			if tt.wantCh == nil {
				require.Nil(t, chatID)
			} else {
				require.NotNil(t, chatID)
				require.Equal(t, *tt.wantCh, *chatID)
			}
		})
	}
}

func TestExtractMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		resp  graphsearch.QueryPostResponseable
		want  []normEntity
		wantN int
	}{
		{
			name:  "nil resp -> nil",
			resp:  nil,
			want:  nil,
			wantN: 0,
		},
		{
			name:  "resp value nil -> nil",
			resp:  newQueryPostResponse(nil),
			want:  []normEntity{},
			wantN: 0,
		},
		{
			name: "skips nil search response / nil hitsContainers / nil hits / nil hit",
			resp: newQueryPostResponse(
				nil,
				func() msmodels.SearchResponseable {
					sr := msmodels.NewSearchResponse()
					sr.SetHitsContainers(nil)
					return sr
				}(),
				newSearchResponse(
					func() msmodels.SearchHitsContainerable {
						hc := msmodels.NewSearchHitsContainer()
						hc.SetHits(nil)
						return hc
					}(),
				),
			),
			want:  []normEntity{},
			wantN: 0,
		},
		{
			name: "skips hits with nil resource id / blank id, keeps valid",
			resp: newQueryPostResponse(
				newSearchResponse(
					newHitsContainer(
						nil,
						newHit(newResource("", nil)),
						newHit(newResource("  ", nil)),
						newHit(newResource("m1", nil)),
						newHit(newResource("m2", map[string]any{
							"channelIdentity": map[string]any{
								"teamId":    "t1",
								"channelId": "c1",
							},
							"chatId": "chat-1",
						})),
					),
				),
			),
			want: []normEntity{
				{msgID: "m1", teamID: "", channelID: "", chatID: ""},
				{msgID: "m2", teamID: "t1", channelID: "c1", chatID: "chat-1"},
			},
			wantN: 2,
		},
		{
			name: "preserves nested order across multiple responses/containers",
			resp: newQueryPostResponse(
				newSearchResponse(
					newHitsContainer(
						newHit(newResource("m1", nil)),
						newHit(newResource("m2", map[string]any{
							"chatId": "c2",
						})),
					),
					newHitsContainer(
						newHit(newResource("m3", map[string]any{
							"channelIdentity": map[string]any{
								"teamId":    "t3",
								"channelId": "ch3",
							},
						})),
					),
				),
				newSearchResponse(
					newHitsContainer(
						newHit(newResource("m4", nil)),
					),
				),
			),
			want: []normEntity{
				{msgID: "m1"},
				{msgID: "m2", chatID: "c2"},
				{msgID: "m3", teamID: "t3", channelID: "ch3"},
				{msgID: "m4"},
			},
			wantN: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractMessages(tt.resp)

			if tt.resp == nil || tt.resp.GetValue() == nil {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Len(t, got, tt.wantN)
			require.Equal(t, tt.want, normEntities(got))
		})
	}
}
