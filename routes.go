package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *Application) routes() http.Handler {
	defaultMiddlewares := alice.New(app.recover, app.logger)
	dynamic := alice.New(app.session.Enable, app.Authenticator)
	protected := dynamic.Append(app.requireAuth)

	mux := http.NewServeMux()
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(app.publicDir))))
	mux.Handle("/", dynamic.ThenFunc(app.Home))
	mux.Handle("/login", dynamic.ThenFunc(app.Login))
	mux.Handle("/register", dynamic.ThenFunc(app.Register))
	mux.Handle("/contact", dynamic.ThenFunc(app.Contact))
	mux.Handle("/about", protected.ThenFunc(app.About))
	mux.Handle("/submit", protected.ThenFunc(app.Submit))
	mux.Handle("/vote", protected.ThenFunc(app.Vote))
	mux.Handle("/comments", protected.ThenFunc(app.Comments))
	mux.Handle("/logout", protected.ThenFunc(app.Logout))

	return defaultMiddlewares.Then(mux)
}
