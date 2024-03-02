package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type TemplateData struct {
	Users       map[string]User
	Messages    []Message
	CurrentUser string
	LoggedIn    time.Time
}

func startHTTP() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/users/update", usersUpdateHandler)
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/oauth2", callbackHandler)

	fmt.Println("HTTP Server listening on", Settings.HttpServer)
	err := http.ListenAndServe(Settings.HttpServer, nil) // Start the server on port 8080
	if err != nil {
		log.Fatal("error starting http server", err)
	}
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

	cookie, err := r.Cookie("session_id")
	if err != nil {
		fmt.Println("error in cookie", err)
	}

	if cookie != nil {
		fmt.Println("Cookie is present")
	} else {
		authurl := createNewProvider()
		http.Redirect(w, r, authurl, http.StatusSeeOther)
		return
	}
	session := UserSessions[cookie.Value]

	data := TemplateData{
		Users:       users,
		Messages:    messages,
		CurrentUser: session.Name,
		LoggedIn:    session.LoggedInAt,
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
	createHTTPUser(w, r)
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

	cookie, err := r.Cookie("session_id")

	if err != nil {
		fmt.Println("Error in cookies", err)
	} else {
		sendMessage(inputMessage, cookie.Value)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func usersUpdateHandler(w http.ResponseWriter, r *http.Request) {
	cookie, error := r.Cookie("session_id")
	if error != nil {
		fmt.Println("Error in cookies", error)
	}

	name := UserSessions[cookie.Value].Name
	current_user := users[name]

	err := json.NewDecoder(r.Body).Decode(&current_user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	users[name] = current_user

	// Respond back to client
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
