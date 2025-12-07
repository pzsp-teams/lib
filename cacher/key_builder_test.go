package cacher

import "testing"

func TestTeamKeyBuilder_ToString(t *testing.T) {
	kb := &TeamKeyBuilder{
		Type: Team,
		Name: "my-team",
	}

	got := kb.ToString()
	want := "$team$:my-team"

	if got != want {
		t.Fatalf("TeamKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestChannelKeyBuilder_ToString(t *testing.T) {
	kb := &ChannelKeyBuilder{
		Type:   Channel,
		TeamID: "team-123",
		Name:   "general",
	}

	got := kb.ToString()
	want := "$channel$:team-123:general"

	if got != want {
		t.Fatalf("ChannelKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestMemberKeyBuilder_ToString(t *testing.T) {
	kb := &MemberKeyBuilder{
		Type:      Member,
		TeamID:    "team-123",
		ChannelID: "chan-456",
		Ref:       "user@example.com",
	}

	got := kb.ToString()
	want := "$member$:team-123:chan-456:user@example.com"

	if got != want {
		t.Fatalf("MemberKeyBuilder.ToString() = %q, want %q", got, want)
	}
}

func TestNewTeamKeyBuilder_SetsFieldsAndImplementsInterface(t *testing.T) {
	var kb = NewTeamKeyBuilder("team-x")
	tkb, ok := kb.(*TeamKeyBuilder)
	if !ok {
		t.Fatalf("NewTeamKeyBuilder() should return *TeamKeyBuilder, got %T", kb)
	}

	if tkb.Type != Team {
		t.Errorf("TeamKeyBuilder.Type = %q, want %q", tkb.Type, Team)
	}
	if tkb.Name != "team-x" {
		t.Errorf("TeamKeyBuilder.Name = %q, want %q", tkb.Name, "team-x")
	}
}

func TestNewChannelKeyBuilder_SetsFieldsAndImplementsInterface(t *testing.T) {
	var kb = NewChannelKeyBuilder("team-1", "chan-x")
	ckb, ok := kb.(*ChannelKeyBuilder)
	if !ok {
		t.Fatalf("NewChannelKeyBuilder() should return *ChannelKeyBuilder, got %T", kb)
	}

	if ckb.Type != Channel {
		t.Errorf("ChannelKeyBuilder.Type = %q, want %q", ckb.Type, Channel)
	}
	if ckb.TeamID != "team-1" {
		t.Errorf("ChannelKeyBuilder.TeamID = %q, want %q", ckb.TeamID, "team-1")
	}
	if ckb.Name != "chan-x" {
		t.Errorf("ChannelKeyBuilder.Name = %q, want %q", ckb.Name, "chan-x")
	}
}

func TestNewMemberKeyBuilder_SetsFieldsAndImplementsInterface(t *testing.T) {
	var kb = NewMemberKeyBuilder("user-ref", "team-1", "chan-1")
	mkb, ok := kb.(*MemberKeyBuilder)
	if !ok {
		t.Fatalf("NewMemberKeyBuilder() should return *MemberKeyBuilder, got %T", kb)
	}

	if mkb.Type != Member {
		t.Errorf("MemberKeyBuilder.Type = %q, want %q", mkb.Type, Member)
	}
	if mkb.Ref != "user-ref" {
		t.Errorf("MemberKeyBuilder.Ref = %q, want %q", mkb.Ref, "user-ref")
	}
	if mkb.TeamID != "team-1" {
		t.Errorf("MemberKeyBuilder.TeamID = %q, want %q", mkb.TeamID, "team-1")
	}
	if mkb.ChannelID != "chan-1" {
		t.Errorf("MemberKeyBuilder.ChannelID = %q, want %q", mkb.ChannelID, "chan-1")
	}
}
