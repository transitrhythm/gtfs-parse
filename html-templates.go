package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func hdl(t *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, r.URL.Query()); err != nil {
			http.Error(w, fmt.Sprintf("error executing template (%s)", err), http.StatusInternalServerError)
		}
	})
}

// HTMLstart -
func HTMLstart() {
	tpl := template.Must(template.New("site.html").ParseGlob("templates/*.html"))
	if err := http.ListenAndServe(":8080", hdl(tpl)); err != nil {
		log.Fatalf("error running server (%s)", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>BC Transit Timetables</h1>")
}

// HTTPTimetables -
func HTTPTimetables() {
	http.HandleFunc("/", indexHandler)
}
