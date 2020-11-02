package types

import "context"

type modeContextKey struct{}

func ModeFromContext(ctx context.Context) Mode {
	return ctx.Value(modeContextKey{}).(Mode)
}

func ContextWithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, modeContextKey{}, mode)
}
