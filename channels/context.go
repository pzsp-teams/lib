package channels

import "context"

type contextKey int

const (
	resolvedTeamIDKey contextKey = iota
	resolvedChannelIDKey
)

func withResolvedTeamID(ctx context.Context, teamID string) context.Context {
	return context.WithValue(ctx, resolvedTeamIDKey, teamID)
}

func withResolvedChannelID(ctx context.Context, channelID string) context.Context {
	return context.WithValue(ctx, resolvedChannelIDKey, channelID)
}
