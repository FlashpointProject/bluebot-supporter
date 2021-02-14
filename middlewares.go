package main

import (
	"context"
	"net/http"
)

func isAuthorized(keys []AccessKey, next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		reqKey := r.Header.Get("access-key")

		if len(reqKey) == 0 {
			http.Error(w, "missing access key", http.StatusUnauthorized)
			return
		}

		for _, ak := range keys {
			if reqKey == ak.UserKey {
				r = r.WithContext(context.WithValue(r.Context(), CtxKeys.UserID, ak.UserID))
				next(w, r)
				return
			}
		}

		http.Error(w, "invalid access key", http.StatusUnauthorized)
	}
}
