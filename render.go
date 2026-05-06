package main

import (
	"net/http"
)

func (app *Application) render(w http.ResponseWriter, r *http.Request, fileName string, data *TemplateData) {
	if app.tempRender == nil {
		http.Error(w, "Template rendering engine not set!", http.StatusInternalServerError)
		return
	}
	tempData := app.defaultTemplateData(r, data)
	app.tempRender.render(w, fileName, tempData)
}

func (app *Application) defaultTemplateData(r *http.Request, data *TemplateData) *TemplateData {
	if data == nil {
		data = &TemplateData{}
	}
	data.Flash = app.session.PopString(r, "flash")
	data.IsAuthenticated = app.isAuthenticated(r)
	return data
}
