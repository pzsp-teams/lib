package api

import (
	"context"
	"net/http"
	"testing"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/search"
	"github.com/stretchr/testify/require"
)

func TestNewAadUserMemberBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userRef  string
		roles    []string
		wantBind string
	}{
		{
			name:     "sets roles and user odata bind",
			userRef:  "user-id-123",
			roles:    []string{"owner"},
			wantBind: "https://graph.microsoft.com/v1.0/users('user-id-123')",
		},
		{
			name:     "empty roles allowed",
			userRef:  "someone@example.com",
			roles:    nil,
			wantBind: "https://graph.microsoft.com/v1.0/users('someone@example.com')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newAadUserMemberBody(tt.userRef, tt.roles)
			require.NotNil(t, got)

			require.Equal(t, tt.roles, got.GetRoles())

			ad := got.GetAdditionalData()
			require.NotNil(t, ad)
			v, ok := ad[graphUserBindKey]
			require.True(t, ok)
			require.Equal(t, tt.wantBind, v)
		})
	}
}

func TestNewRolesPatchBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		roles []string
	}{
		{name: "owner role", roles: []string{"owner"}},
		{name: "empty roles", roles: []string{}},
		{name: "nil roles", roles: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newRolesPatchBody(tt.roles)
			require.NotNil(t, got)
			require.Equal(t, tt.roles, got.GetRoles())

			_ = got.GetAdditionalData()
		})
	}
}

func TestAddToMembers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initialSize int
		userRefs    []string
		roles       []string
		wantAdded   int
	}{
		{
			name:        "adds multiple members",
			initialSize: 0,
			userRefs:    []string{"u1", "u2", "u3"},
			roles:       []string{"owner"},
			wantAdded:   3,
		},
		{
			name:        "no-op when no user refs",
			initialSize: 1,
			userRefs:    nil,
			roles:       []string{"owner"},
			wantAdded:   0,
		},
		{
			name:        "keeps existing members and appends",
			initialSize: 2,
			userRefs:    []string{"uX"},
			roles:       nil,
			wantAdded:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			members := make([]msmodels.ConversationMemberable, 0, tt.initialSize+len(tt.userRefs))
			for i := 0; i < tt.initialSize; i++ {
				members = append(members, newAadUserMemberBody("existing", nil))
			}

			before := len(members)
			addToMembers(&members, tt.userRefs, tt.roles)

			require.Equal(t, before+tt.wantAdded, len(members))

			for i, userRef := range tt.userRefs {
				got := members[before+i]
				require.NotNil(t, got)
				require.Equal(t, tt.roles, got.GetRoles())

				ad := got.GetAdditionalData()
				require.NotNil(t, ad)
				v, ok := ad[graphUserBindKey]
				require.True(t, ok)
				require.Equal(t, "https://graph.microsoft.com/v1.0/users('"+userRef+"')", v)
			}
		})
	}
}

func TestMessageToGraph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		contentType string
		wantType    msmodels.BodyType
	}{
		{
			name:        "default to text when contentType not html",
			content:     "hello",
			contentType: "text",
			wantType:    msmodels.TEXT_BODYTYPE,
		},
		{
			name:        "html sets HTML body type",
			content:     "<b>hi</b>",
			contentType: "html",
			wantType:    msmodels.HTML_BODYTYPE,
		},
		{
			name:        "empty content type -> text",
			content:     "x",
			contentType: "",
			wantType:    msmodels.TEXT_BODYTYPE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := messageToGraph(tt.content, tt.contentType)
			require.NotNil(t, got)

			require.NotNil(t, got.GetContent())
			require.Equal(t, tt.content, *got.GetContent())

			require.NotNil(t, got.GetContentType())
			require.Equal(t, tt.wantType, *got.GetContentType())
		})
	}
}

func TestNewTypeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
		wantMsg  string
	}{
		{
			name:     "builds unprocessable entity error",
			expected: "msmodels.Userable",
			wantMsg:  "Expected msmodels.Userable",
		},
		{
			name:     "empty expected still formats",
			expected: "",
			wantMsg:  "Expected ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := newTypeError(tt.expected)
			require.NotNil(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, err.Code)
			require.Equal(t, tt.wantMsg, err.Message)

			_ = err
		})
	}
}

func TestIsSystemEvent(t *testing.T) {
	t.Parallel()

	t.Run("returns true when eventDetail is present", func(t *testing.T) {
		t.Parallel()

		m := msmodels.NewChatMessage()
		m.SetEventDetail(msmodels.NewChannelAddedEventMessageDetail())

		require.True(t, isSystemEvent(m))
	})

	t.Run("returns true when messageType is chatEvent", func(t *testing.T) {
		t.Parallel()

		m := msmodels.NewChatMessage()
		mt := msmodels.CHATEVENT_CHATMESSAGETYPE
		m.SetMessageType(&mt)

		require.True(t, isSystemEvent(m))
	})

	t.Run("returns false for normal message (no eventDetail and no chatEvent type)", func(t *testing.T) {
		t.Parallel()

		m := msmodels.NewChatMessage()
		require.False(t, isSystemEvent(m))
	})
}

func TestFilterOutSystemEvents(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when value is nil", func(t *testing.T) {
		t.Parallel()

		resp := msmodels.NewChatMessageCollectionResponse()
		got := filterOutSystemEvents(resp)

		require.Nil(t, got)
	})

	t.Run("filters out nil entries and system events, keeps order", func(t *testing.T) {
		t.Parallel()

		normal1 := msmodels.NewChatMessage()
		id1 := "m1"
		normal1.SetId(&id1)

		sysByType := msmodels.NewChatMessage()
		sysType := msmodels.CHATEVENT_CHATMESSAGETYPE
		sysByType.SetMessageType(&sysType)

		sysByDetail := msmodels.NewChatMessage()
		sysByDetail.SetEventDetail(msmodels.NewChannelAddedEventMessageDetail())

		normal2 := msmodels.NewChatMessage()
		id2 := "m2"
		normal2.SetId(&id2)

		resp := msmodels.NewChatMessageCollectionResponse()
		resp.SetValue([]msmodels.ChatMessageable{
			nil,
			normal1,
			sysByType,
			sysByDetail,
			normal2,
		})

		got := filterOutSystemEvents(resp)

		require.Len(t, got, 2)
		require.NotNil(t, got[0].GetId())
		require.NotNil(t, got[1].GetId())
		require.Equal(t, "m1", *got[0].GetId())
		require.Equal(t, "m2", *got[1].GetId())
	})

	t.Run("returns empty slice when all messages are filtered out", func(t *testing.T) {
		t.Parallel()

		sysByType := msmodels.NewChatMessage()
		sysType := msmodels.CHATEVENT_CHATMESSAGETYPE
		sysByType.SetMessageType(&sysType)

		resp := msmodels.NewChatMessageCollectionResponse()
		resp.SetValue([]msmodels.ChatMessageable{sysByType})

		got := filterOutSystemEvents(resp)
		require.NotNil(t, got)
		require.Len(t, got, 0)
	})
}

func TestNormalizeVisibilityForGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "private lower", in: "private", want: "Private"},
		{name: "private mixed + spaces", in: "  PrIvAtE  ", want: "Private"},
		{name: "public lower", in: "public", want: "Public"},
		{name: "public upper", in: "PUBLIC", want: "Public"},
		{name: "empty -> public", in: "", want: "Public"},
		{name: "spaces only -> public", in: "   ", want: "Public"},
		{name: "passthrough unknown", in: "HiddenMembership", want: "HiddenMembership"},
		{name: "passthrough already proper", in: "Private", want: "Private"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeVisibilityForGroup(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseTeamIDFromHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		contentLocation string
		location        string
		wantID          string
		wantOK          bool
	}{
		{
			name:            "extracts from Content-Location quoted format",
			contentLocation: "/teams('TEAM-ID-1')",
			location:        "",
			wantID:          "TEAM-ID-1",
			wantOK:          true,
		},
		{
			name:            "extracts from Location quoted format",
			contentLocation: "",
			location:        "/teams('TEAM-ID-2')",
			wantID:          "TEAM-ID-2",
			wantOK:          true,
		},
		{
			name:            "extracts from Location when contains operations segment",
			contentLocation: "",
			location:        "/teams('TEAM-ID-3')/operations('OP-ID')",
			wantID:          "TEAM-ID-3",
			wantOK:          true,
		},
		{
			name:            "extracts from Content-Location when contains operations segment",
			contentLocation: "/teams('TEAM-ID-4')/operations('OP-ID')",
			location:        "",
			wantID:          "TEAM-ID-4",
			wantOK:          true,
		},
		{
			name:            "extracts from slash format /teams/{id}",
			contentLocation: "/teams/TEAM-ID-5",
			location:        "",
			wantID:          "TEAM-ID-5",
			wantOK:          true,
		},
		{
			name:            "extracts from full-ish path with query (slash format stops before ?)",
			contentLocation: "/teams/TEAM-ID-6?$select=id",
			location:        "",
			wantID:          "TEAM-ID-6",
			wantOK:          true,
		},
		{
			name:            "trims spaces around header values",
			contentLocation: "   /teams('TEAM-ID-7')   ",
			location:        "",
			wantID:          "TEAM-ID-7",
			wantOK:          true,
		},
		{
			name:            "prefers Content-Location over Location when both are present",
			contentLocation: "/teams('TEAM-ID-CL')",
			location:        "/teams('TEAM-ID-LOC')",
			wantID:          "TEAM-ID-CL",
			wantOK:          true,
		},
		{
			name:            "returns false when both headers empty",
			contentLocation: "",
			location:        "",
			wantID:          "",
			wantOK:          false,
		},
		{
			name:            "returns false when headers do not contain a team id",
			contentLocation: "/groups('G1')",
			location:        "/operations('OP1')",
			wantID:          "",
			wantOK:          false,
		},
		{
			name:            "returns false when only operations without teams",
			contentLocation: "/teamsTemplates('standard')/operations('OP1')",
			location:        "",
			wantID:          "",
			wantOK:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseTeamIDFromHeaders(tt.contentLocation, tt.location)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantID, got)
		})
	}
}

func TestFilterTrimNonEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "nil -> empty slice",
			in:   nil,
			want: []string{},
		},
		{
			name: "filters blanks and trims",
			in:   []string{" a ", "", "   ", "\t", "b", "  c  "},
			want: []string{"a", "b", "c"},
		},
		{
			name: "all blanks -> empty",
			in:   []string{"", " ", "\n", "\t"},
			want: []string{},
		},
		{
			name: "keeps order",
			in:   []string{" z ", " y", "x "},
			want: []string{"z", "y", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := filterTrimNonEmpty(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidateCreateFromTemplate(t *testing.T) {
	t.Parallel()

	t.Run("empty -> bad request", func(t *testing.T) {
		t.Parallel()

		err := validateCreateFromTemplate("")
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.Code)
		require.Equal(t, "displayName cannot be empty", err.Message)
	})

	t.Run("non-empty -> ok", func(t *testing.T) {
		t.Parallel()

		err := validateCreateFromTemplate("Team A")
		require.Nil(t, err)
	})

	t.Run("spaces only -> ok (current behavior)", func(t *testing.T) {
		t.Parallel()

		err := validateCreateFromTemplate("   ")
		require.Nil(t, err)
	})
}

func TestBuildTeamFromTemplateBody(t *testing.T) {
	t.Parallel()

	t.Run("private + description -> sets expected fields and primary owner", func(t *testing.T) {
		t.Parallel()

		body := buildTeamFromTemplateBody("Team A", "Desc", "private", "owner-id-1")
		require.NotNil(t, body)

		require.NotNil(t, body.GetDisplayName())
		require.Equal(t, "Team A", *body.GetDisplayName())

		require.NotNil(t, body.GetDescription())
		require.Equal(t, "Desc", *body.GetDescription())

		require.NotNil(t, body.GetVisibility())
		require.Equal(t, msmodels.PRIVATE_TEAMVISIBILITYTYPE, *body.GetVisibility())

		require.NotNil(t, body.GetFirstChannelName())
		require.Equal(t, "General", *body.GetFirstChannelName())

		ad := body.GetAdditionalData()
		require.NotNil(t, ad)
		require.Equal(t, templateBindValue, ad[templateBindKey])

		members := body.GetMembers()
		require.NotNil(t, members)
		require.Len(t, members, 1)

		m := members[0]
		require.NotNil(t, m)
		require.Equal(t, []string{roleOwner}, m.GetRoles())

		mad := m.GetAdditionalData()
		require.NotNil(t, mad)
		require.Equal(t, "https://graph.microsoft.com/v1.0/users('owner-id-1')", mad[graphUserBindKey])
	})

	t.Run("public default + empty description -> description not set", func(t *testing.T) {
		t.Parallel()

		body := buildTeamFromTemplateBody("Team A", "", "PUBLIC", "owner-id-1")
		require.NotNil(t, body)

		require.NotNil(t, body.GetVisibility())
		require.Equal(t, msmodels.PUBLIC_TEAMVISIBILITYTYPE, *body.GetVisibility())

		require.Nil(t, body.GetDescription())
	})
}

func TestTeamAPI_NormalizeOwners_WithoutIncludeMe(t *testing.T) {
	t.Parallel()

	t.Run("filters blanks and returns owners", func(t *testing.T) {
		t.Parallel()

		tapi := &teamAPI{}
		got, err := tapi.normalizeOwners(context.Background(), []string{" a ", "", "  ", "b"}, false)

		require.Nil(t, err)
		require.Equal(t, []string{"a", "b"}, got)
	})

	t.Run("no owners -> bad request", func(t *testing.T) {
		t.Parallel()

		tapi := &teamAPI{}
		got, err := tapi.normalizeOwners(context.Background(), []string{"", "   "}, false)

		require.Nil(t, got)
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.Code)
		require.Equal(t, "at least one owner is required", err.Message)
	})
}

func TestTeamAPI_AddMembersInBulk_Empty_NoOps(t *testing.T) {
	t.Parallel()

	tapi := &teamAPI{}
	err := tapi.addMembersInBulk(context.Background(), "team-id", nil)
	require.Nil(t, err)

	err = tapi.addMembersInBulk(context.Background(), "team-id", []string{})
	require.Nil(t, err)
}

func TestTeamAPI_AddPostCreateMembersAndOwners_NoOps(t *testing.T) {
	t.Parallel()

	tapi := &teamAPI{}
	err := tapi.addPostCreateMembersAndOwners(context.Background(), "team-id", nil, nil)
	require.Nil(t, err)

	err = tapi.addPostCreateMembersAndOwners(context.Background(), "team-id", []string{"  "}, []string{"\t"})
	require.Nil(t, err)
}

func TestGetHeaderValue(t *testing.T) {
	t.Parallel()

	t.Run("nil headers -> empty", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "", getHeaderValue(nil, "location"))
	})

	t.Run("wrong type -> empty", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "", getHeaderValue(struct{}{}, "location"))
	})

	t.Run("nil *ResponseHeaders -> empty", func(t *testing.T) {
		t.Parallel()
		var rh *abstractions.ResponseHeaders
		require.Equal(t, "", getHeaderValue(rh, "location"))
	})

	t.Run("missing key -> empty (no panic)", func(t *testing.T) {
		t.Parallel()

		rh := abstractions.NewResponseHeaders()
		rh.Add("some-other-header", "x")

		require.NotPanics(t, func() {
			_ = getHeaderValue(rh, "location")
		})
		require.Equal(t, "", getHeaderValue(rh, "location"))
	})

	t.Run("present key -> returns first value", func(t *testing.T) {
		t.Parallel()

		rh := abstractions.NewResponseHeaders()
		rh.Add("location", "LOC-1", "LOC-2")
		rh.Add("content-location", "CL-1")

		require.Contains(t, []string{"LOC-1", "LOC-2"}, getHeaderValue(rh, "location"))
		require.Equal(t, "CL-1", getHeaderValue(rh, "content-location"))
	})

	t.Run("key is normalized to lower-case", func(t *testing.T) {
		t.Parallel()

		rh := abstractions.NewResponseHeaders()
		rh.Add("Location", "LOC-1")

		require.Equal(t, "LOC-1", getHeaderValue(rh, "LOCATION"))
		require.Equal(t, "LOC-1", getHeaderValue(rh, "location"))
	})
}

func TestCalcNextSearchFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     *search.SearchMessagesOptions
		returned int
		want     int32
	}{
		{
			name:     "nil opts -> from=0",
			opts:     nil,
			returned: 0,
			want:     0,
		},
		{
			name:     "nil opts + returned -> increments from 0",
			opts:     nil,
			returned: 7,
			want:     7,
		},
		{
			name: "nil SearchPage -> from=0",
			opts: &search.SearchMessagesOptions{
				SearchPage: nil,
			},
			returned: 3,
			want:     3,
		},
		{
			name: "SearchPage but nil From -> from=0",
			opts: &search.SearchMessagesOptions{
				SearchPage: &search.SearchPage{From: nil},
			},
			returned: 5,
			want:     5,
		},
		{
			name: "From set + returned 0",
			opts: &search.SearchMessagesOptions{
				SearchPage: &search.SearchPage{From: util.Ptr[int32](10)},
			},
			returned: 0,
			want:     10,
		},
		{
			name: "From set + returned adds",
			opts: &search.SearchMessagesOptions{
				SearchPage: &search.SearchPage{From: util.Ptr[int32](10)},
			},
			returned: 6,
			want:     16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := calcNextSearchFrom(tt.opts, tt.returned)
			require.NotNil(t, got)
			require.Equal(t, tt.want, *got)
		})
	}
}

func TestParseTeamIDFromHeaders_ExtraCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		contentLocation string
		location        string
		wantID          string
		wantOK          bool
	}{
		{
			name:            "full URL quoted format",
			contentLocation: "https://graph.microsoft.com/v1.0/teams('TEAM-ID-1')",
			location:        "",
			wantID:          "TEAM-ID-1",
			wantOK:          true,
		},
		{
			name:            "full URL slash format",
			contentLocation: "https://graph.microsoft.com/v1.0/teams/TEAM-ID-2",
			location:        "",
			wantID:          "TEAM-ID-2",
			wantOK:          true,
		},
		{
			name:            "full URL with operations quoted format",
			contentLocation: "https://graph.microsoft.com/v1.0/teams('TEAM-ID-3')/operations('OP')",
			location:        "",
			wantID:          "TEAM-ID-3",
			wantOK:          true,
		},
		{
			name:            "full URL with operations slash format",
			contentLocation: "https://graph.microsoft.com/v1.0/teams/TEAM-ID-4/operations/OP",
			location:        "",
			wantID:          "TEAM-ID-4",
			wantOK:          true,
		},
		{
			name:            "unknown header content -> false",
			contentLocation: "https://graph.microsoft.com/v1.0/groups('G1')",
			location:        "",
			wantID:          "",
			wantOK:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseTeamIDFromHeaders(tt.contentLocation, tt.location)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantID, got)
		})
	}
}

func TestFilterOutSystemEvents_EmptySliceValue(t *testing.T) {
	t.Parallel()

	resp := msmodels.NewChatMessageCollectionResponse()
	resp.SetValue([]msmodels.ChatMessageable{})

	got := filterOutSystemEvents(resp)
	require.NotNil(t, got)
	require.Len(t, got, 0)
}

func TestMessageToGraph_ContentTypeIsCaseSensitive(t *testing.T) {
	t.Parallel()

	body := messageToGraph("<b>hi</b>", "HTML") 
	require.NotNil(t, body)

	require.NotNil(t, body.GetContentType())
	require.Equal(t, msmodels.TEXT_BODYTYPE, *body.GetContentType())
}
