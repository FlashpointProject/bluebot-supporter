package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

type contextString string

type ctxKeys struct {
	Log       contextString
	RequestID contextString
}

// CtxKeys is context value keys
var CtxKeys = ctxKeys{
	Log:       "Log",
	RequestID: "requestID",
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

	RequestID := requestIDFromContext(ctx)
	if len(RequestID) > 0 {
		entry = entry.WithField(string(CtxKeys.RequestID), RequestID)
	}

	return entry
}
