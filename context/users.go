package context

import (
	"Gallery/models"
	"context"
)

type key string

const (
	userKey key = "user"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	user, ok := val.(*models.User)
	// fmt.Println(val)
	if !ok {
		return nil
	}
	return user
}
