package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type TemplateData struct {
	Users       map[string]User
	Messages    []Message
	CurrentUser string
}

func startHTTP() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/message", messageHandler)

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
		Users:       users,
		Messages:    messages,
		CurrentUser: currentUser,
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

func userHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	name := r.Form.Get("name")

	if users[name].Name != "" {
		http.SetCookie(w, &http.Cookie{
			Name:  "error",
			Value: "User already exists",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	users[name] = User{Name: name}

	currentUser = name
	http.SetCookie(w, &http.Cookie{
		Name:  "name",
		Value: name,
		Path:  "/",
	})
	messageChannel <- Message{
		TimeStamp: time.Now(),
		Text:      genericMessage["joined"],
		Name:      name,
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	inputMessage := r.Form.Get("message")

	cookie, err := r.Cookie("name")

	if err != nil {
		fmt.Println("Error in cookies", err)
	}

	messageChannel <- Message{
		TimeStamp: time.Now(),
		Text:      inputMessage,
		Name:      cookie.Value,
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
