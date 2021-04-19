package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

func LandingViewHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}
}
