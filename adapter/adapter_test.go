package adapter

import (
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/stretchr/testify/assert"
)

type NewTeamParams struct {
	ID          string
	DisplayName string
	Description string
	IsArchived  bool
	Visibilitiy msmodels.TeamVisibilityType
}

func newGraphTeam(params *NewTeamParams) msmodels.Teamable {
	t := msmodels.NewTeam()
	t.SetId(&params.ID)
	t.SetDisplayName(&params.DisplayName)
	t.SetDescription(&params.Description)
	t.SetIsArchived(&params.IsArchived)
	t.SetVisibility(&params.Visibilitiy)
	return t
}

func TestMapGraphTeam_NilInput(t *testing.T) {
	team := MapGraphTeam(nil)

	assert.Nil(t, team)
}

func TestMapGraphTeam_AllFieldsPresent(t *testing.T) {
	teamParams := &NewTeamParams{
		ID:          "team-id",
		DisplayName: "Team Name",
		Description: "A sample team",
		IsArchived:  true,
		Visibilitiy: msmodels.PUBLIC_TEAMVISIBILITYTYPE,
	}
	graphTeam := newGraphTeam(teamParams)

	team := MapGraphTeam(graphTeam)

	assert.Equal(t, team.ID, *graphTeam.GetId())
	assert.Equal(t, team.DisplayName, *graphTeam.GetDisplayName())
	assert.Equal(t, team.Description, *graphTeam.GetDescription())
	assert.Equal(t, team.IsArchived, *graphTeam.GetIsArchived())
	assert.Equal(t, team.Visibility, "public")
}

func TestMapGraphTeam_MissingFields(t *testing.T) {
	teamParams := &NewTeamParams{}
	graphTeam := newGraphTeam(teamParams)

	team := MapGraphTeam(graphTeam)

	assert.Equal(t, team.ID, "")
	assert.Equal(t, team.DisplayName, "")
	assert.Equal(t, team.Description, "")
	assert.Equal(t, team.IsArchived, false)
	assert.Equal(t, team.Visibility, "private")
}
