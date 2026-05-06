package main

import (
	"log"
	"net/http"
)

func (app *Application) Serve() {
	srv := http.Server{
		Addr:    ":8080",
		Handler: app.routes(),
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
