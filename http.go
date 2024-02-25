package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	http.HandleFunc("/login", loginHandler)

	fmt.Println("HTTP Server listening on", Settings.HttpServer)
	http.ListenAndServe(Settings.HttpServer, nil) // Start the server on port 8080
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"formatTime": formatTime,
	}).ParseGlob("./views/app/*.html")

	if err != nil {
		fmt.Println("can't parse the files", err)
		w.Write([]byte(err.Error()))
		return
	}

	cookie, err := r.Cookie("name")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := TemplateData{
		Users:       users,
		Messages:    messages,
		CurrentUser: cookie.Value,
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
	password := r.Form.Get("password")

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		r.Header.Set("ERROR", "Unable to create user")
		loginHandler(w, r)
		return
	}

	_, ok := users[name]
	if ok {
		// Authenticate user by comparing password with the hashed password
		err := bcrypt.CompareHashAndPassword([]byte(users[name].PasswordHash), []byte(password))
		if err != nil {
			r.Header.Set("ERROR", "Invalid password")
			loginHandler(w, r)
			return
		}
	}

	users[name] = User{
		Name:         name,
		PasswordHash: string(passwordHash),
	}

	BackupData(users[name], "./users.db")

	http.SetCookie(w, &http.Cookie{
		Name:     "name",
		Value:    name,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	messageChannel <- Message{
		TimeStamp: time.Now(),
		Text:      genericMessage["joined"],
		Name:      name,
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./views/login/new.html")
	if err != nil {
		fmt.Println("can't parse the files", err)
		w.Write([]byte(err.Error()))
		return
	}

	err = tmpl.Execute(w, r.Header.Get("ERROR"))
	if err != nil {
		fmt.Println("Unable to render templates.", err)
		w.Write([]byte(err.Error()))
	}
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
	} else {
		sendMessage(nil, inputMessage, cookie.Value)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
