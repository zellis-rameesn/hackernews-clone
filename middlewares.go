package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

// defines a custom type based on string.
// Now contextKey is different from normal string.
// This prevents collisions.
type contextKey string
type userKey User

// This creates a typed key used to store authentication info inside the request context.
const (
	contextAuthKey contextKey = contextKey("isAuthKey")
	contextUserKey contextKey = contextKey("authUserKey")
)

func (app *Application) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// do something before request
		app.infoLog.Println("Started:", r.Method, r.URL.Path)
		// Continue to the actual handler.
		next.ServeHTTP(w, r)
		// do something after request
		app.infoLog.Println("Completed in:", time.Since(start))
	})
}

func (app *Application) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app Application) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedIn := app.session.Exists(r, loggedInUserEmail)
		app.infoLog.Println("AUTHENTICATOR MIDDLEWARE")
		if !loggedIn {
			app.infoLog.Println("user not logged in!!")
			next.ServeHTTP(w, r)
			return
		}

		user, err := app.userRepo.FetchUser(app.session.GetString(r, loggedInUserEmail))
		if err == sql.ErrNoRows {
			app.infoLog.Println("No users found!")
			app.session.Remove(r, loggedInUserEmail)
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}
		if r.URL.Path == "/login" || r.URL.Path == "/register" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		ctx := context.WithValue(r.Context(), contextAuthKey, true)
		ctx = context.WithValue(ctx, contextUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, fmt.Sprintf("/login?redirectTo=%s", r.URL.Path), http.StatusSeeOther)
			return
		}
		// To prevent browsers from caching protected pages. so that if I logout and go back I can't view that protected page
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)

	})
}

func (app *Application) isAuthenticated(r *http.Request) bool {
	isAuth, ok := r.Context().Value(contextAuthKey).(bool)
	app.infoLog.Println("isAuth:", isAuth)
	if !ok {
		return false
	}
	return isAuth
}
