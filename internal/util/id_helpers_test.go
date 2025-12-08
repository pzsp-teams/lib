package util

import "testing"

func TestIsLikelyGUID_Positive(t *testing.T) {
	s := "123e4567-e89b-12d3-a456-426614174000"
	if !IsLikelyGUID(s) {
		t.Fatalf("expected IsLikelyGUID(%q)=true, got false", s)
	}
}

func TestIsLikelyGUID_Negative(t *testing.T) {
	for _, s := range []string{
		"", "not-a-guid", "123e4567-e89b-12d3-a456-42661417400", "zzze4567-e89b-12d3-a456-426614174000",
	} {
		if IsLikelyGUID(s) {
			t.Fatalf("expected IsLikelyGUID(%q)=false, got true", s)
		}
	}
}

func TestIsLikelyChannelID_Positive(t *testing.T) {
	s := "19:xxx@thread."
	if !IsLikelyChannelID(s) {
		t.Fatalf("expected IsLikelyChannelID(%q)=true, got false", s)
	}
}

func TestIsLikelyChannelID_Negative(t *testing.T) {
	for _, s := range []string{
		"", "not-a-channel-id", "20:xxx@thread", "19:xxx@notthread",
	} {
		if IsLikelyChannelID(s) {
			t.Fatalf("expected IsLikelyChannelID(%q)=false, got true", s)
		}
	}
}
