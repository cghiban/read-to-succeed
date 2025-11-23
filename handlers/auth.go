package handlers

import (
	"context"
	"net/http"
	"read2succeed/data"
)

// Auth provides middleware functions for authorizing users and setting the user
// in the request context.
type Auth struct {
	Service *Service
}

// UserViaSession will retrieve the current user set by the session cookie
// and set it in the request context. UserViaSession will NOT redirect
// to the sign in page if the user is not found. That is left for the
// RequireUser method to handle so that some pages can optionally have
// access to the current user.
func (a *Auth) UserViaSession(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session, err := a.Service.session.Get(r, "session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		//a.Service.l.Printf("logged_in: %v\t%T", session.Values["logged_in"], session.Values["logged_in"])
		if session.Values["logged_in"] != true {
			next.ServeHTTP(w, r)
			return
		}

		user_id, _ := session.Values["user_id"].(int)
		user, err := a.Service.store.GetUserByID(user_id)
		if err != nil {
			// If you want you can retain the original functionality to call
			// http.Error if any error aside from app.ErrNotFound is returned,
			// but I find that most of the time we can continue on and let later
			// code error if it requires a user, otherwise it can continue without
			// the user.
			next.ServeHTTP(w, r)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", user))
		next.ServeHTTP(w, r)
	}
}

// RequireUser will verify that a user is set in the request context. It if is
// set correctly, the next handler will be called, otherwise it will redirect
// the user to the sign in page and the next handler will not be called.
func (a *Auth) RequireUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp := r.Context().Value("user")
		if tmp == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if _, ok := tmp.(*data.AuthUser); !ok {
			// Whatever was set in the user key isn't a user, so we probably need to
			// sign in.
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	}
}
