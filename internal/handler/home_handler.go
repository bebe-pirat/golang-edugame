package handler

import (
	"html/template"
	"net/http"
)

type HomeHandler struct {
	tmpl *template.Template
}

func NewHomeHandler() *IndexHandler {
	tmpl := template.Must(template.ParseFiles("../../internal/templates/home.html"))
	return &IndexHandler{tmpl: tmpl}
}

func (i *IndexHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	i.tmpl.Execute(w, nil)
}
