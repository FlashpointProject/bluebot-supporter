package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextString string

type ctxKeys struct {
	UserID    contextString
	Log       contextString
	RequestID contextString
}

// CtxKeys is context value keys
var CtxKeys = ctxKeys{
	UserID:    "userID",
	Log:       "Log",
	RequestID: "requestID",
}

func userIDFromContext(ctx context.Context) string {
	v := ctx.Value(CtxKeys.UserID)
	if v == nil {
		return ""
	}
	return v.(string)
}

func requestIDFromContext(ctx context.Context) string {
	v := ctx.Value(CtxKeys.RequestID)
	if v == nil {
		return ""
	}
	return v.(string)
}

func lc(ctx context.Context) *logrus.Entry {
	l := ctx.Value(CtxKeys.Log).(*logrus.Logger)
	entry := logrus.NewEntry(l)

	userID := userIDFromContext(ctx)
	if len(userID) > 0 {
		entry = entry.WithField(string(CtxKeys.UserID), userID)
	}
	RequestID := requestIDFromContext(ctx)
	if len(RequestID) > 0 {
		entry = entry.WithField(string(CtxKeys.RequestID), RequestID)
	}

	return entry
}
