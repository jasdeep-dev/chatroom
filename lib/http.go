package lib

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

type TemplateData struct {
	Users       map[int]User
	Messages    []Message
	CurrentUser KeyCloakUserInfo
	LoggedIn    time.Time
}

func StartHTTP() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/oauth2", callbackHandler)
	http.HandleFunc("/logout", logoutHandler)

	log.Println("HTTP Server listening on", Settings.HttpServer)
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
		log.Println("can't parse the files", err)
		w.Write([]byte(err.Error()))
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("error in cookies", err)
	}

	if cookie == nil {
		createNewProvider(w, r)
		return
	}

	session := UserSessions[cookie.Value]

	data := TemplateData{
		Users:       Users,
		Messages:    MessagesArray,
		CurrentUser: session.KeyCloakUser,
		LoggedIn:    session.LoggedInAt,
	}

	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Println("Unable to render templates.", err)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./views/login/new.html")
	if err != nil {
		log.Println("can't parse the files", err)
		w.Write([]byte(err.Error()))
		return
	}

	err = tmpl.Execute(w, r.Header.Get("ERROR"))
	if err != nil {
		log.Println("Unable to render templates.", err)
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
		log.Println("Error in cookies", err)
	} else {
		sendMessage(inputMessage, cookie.Value)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	DeleteCookie("session_id", w)
	DeleteCookie("state", w)
	DeleteCookie("nonce", w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func DeleteCookie(name string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}