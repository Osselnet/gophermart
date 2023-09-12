package auth

import (
	"context"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"net/http"
)

func AuthCheck(gm *gophermart.GopherMart) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("session_token")
			if err != nil {
				if err == http.ErrNoCookie {
					http.Error(w, gophermart.ErrUnauthorizedAccess.Error(), http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			sessionToken := c.Value

			session, err := gm.Sessions.Get(sessionToken)
			if err != nil {
				http.Error(w, "session token is not present", http.StatusUnauthorized)
				return
			}

			if session.IsExpired() {
				gm.Sessions.Delete(sessionToken)
				http.Error(w, "session has expired", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), "session", session))

			next.ServeHTTP(w, r)
		})
	}
}
