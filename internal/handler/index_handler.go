package handler

import (
	"html/template"
	"net/http"
)

type IndexHandler struct {
	tmpl *template.Template
}

func NewIndexHandler() *IndexHandler {
	tmpl := template.Must(template.ParseFiles("../../internal/templates/index.html"))
	return &IndexHandler{tmpl: tmpl}
}

func (i *IndexHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	i.tmpl.Execute(w, nil)
}
