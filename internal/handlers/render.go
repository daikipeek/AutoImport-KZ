package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type TemplateRenderer struct {
	Templates *template.Template
}

func NewTemplateRenderer(dir string) (*TemplateRenderer, error) {
	pattern := filepath.Join(dir, "*.html")
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}
	return &TemplateRenderer{Templates: tmpl}, nil
}

func (t *TemplateRenderer) Render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template render error (%s): %v", name, err)
		http.Error(w, "template rendering error", http.StatusInternalServerError)
	}
}
