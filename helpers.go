package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
)

func (app *Application) serverError(w http.ResponseWriter, err error) {
	// logs after recovering from panic
	trace := fmt.Sprintf("%s /n %s", err.Error(), debug.Stack())

	app.errorLog.Output(2, trace)
	// sending http response
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) getUserFromContext(r *http.Request) *User {
	user, ok := r.Context().Value(contextUserKey).(*User)
	if !ok {
		panic("User not found in context!")
	}
	return user
}

func (app *Application) getFilterFromUrl(r *http.Request, key string, dValue int) int {
	val := r.URL.Query().Get(key)
	convVal, err := strconv.Atoi(val)
	if err != nil {
		return dValue
	} else {
		return convVal
	}
}
