package api

import (
	"net/http"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
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
