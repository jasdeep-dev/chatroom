package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type TemplateData struct {
	Users    map[string]User
	Messages []Message
}

func startHTTP() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", indexHandler)
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil) // Start the server on port 8080
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"formatTime": formatTime,
	}).ParseGlob("./views/*.html")

	if err != nil {
		fmt.Println("can't parse the files", err)
		w.Write([]byte(err.Error()))
		return
	}

	data := TemplateData{
		Users:    users,
		Messages: messages,
	}

	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		fmt.Println("Unable to render templates.", err)
		w.Write([]byte(err.Error()))
	}
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
