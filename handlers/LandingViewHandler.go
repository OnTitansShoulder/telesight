package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

func LandingViewHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	}
}
