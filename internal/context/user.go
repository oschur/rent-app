package context

import (
	"context"
)

type UserInfo struct {
	UserID     int
	IsLandlord bool
	IsAdmin    bool
}

type contextKey string

const userContextKey contextKey = "user"

func SetUserInfo(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, userContextKey, userInfo)
}

func GetUserInfo(ctx context.Context) *UserInfo {
	userInfo, ok := ctx.Value(userContextKey).(*UserInfo)
	if !ok {
		return nil
	}
	return userInfo
}
