package main

import (
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"sync"
)

type TemplateRenderer struct {
	cache       map[string]*template.Template
	isDev       bool
	mutex       sync.RWMutex
	templateDir string
}

type TemplateData struct {
	Form            *Form
	IsAuthenticated bool
	Flash           string
	Posts           []Post
	Post            *Post
	Metadata        Metadata
	Comments        []Comment
	NextPage        int
	PrevPage        int
}

func NewTemplateRenderer(tempDir string, isDev bool) *TemplateRenderer {
	return &TemplateRenderer{
		cache:       make(map[string]*template.Template),
		isDev:       isDev,
		templateDir: tempDir,
	}
}

func (t *TemplateRenderer) render(w http.ResponseWriter, fileName string, data *TemplateData) {
	tmpl, err := t.getTemplate(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (t *TemplateRenderer) getTemplate(templateName string) (*template.Template, error) {
	if !t.isDev {
		// RLock for reading from cache (concurrent reads allowed)
		t.mutex.RLock()
		cachedTemplate, ok := t.cache[templateName]
		t.mutex.RUnlock()
		if ok {
			return cachedTemplate, nil
		}
	}

	tmpl, err := t.parseTemplate(templateName)
	if err != nil {
		return nil, err
	}

	if !t.isDev {
		// Lock for writing to cache (exclusive access)
		t.mutex.Lock()
		t.cache[templateName] = tmpl
		t.mutex.Unlock()
	}

	return tmpl, err
}

func (t *TemplateRenderer) parseTemplate(templateName string) (*template.Template, error) {
	templatePath := path.Join(t.templateDir, templateName)
	files := []string{templatePath}
	layoutPath := path.Join(t.templateDir, "layouts/*.html")
	layoutFiles, err := filepath.Glob(layoutPath)
	// no need to check error, because without layout also page should load
	if err == nil {
		files = append(files, layoutFiles...)
	}

	partialPath := path.Join(t.templateDir, "partials/*.html")
	partialFiles, err := filepath.Glob(partialPath)
	if err == nil {
		files = append(files, partialFiles...)
	}

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
