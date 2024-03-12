package lib

import (
	"chatroom/views"
	"context"
	"html/template"
	"log"
	"net/http"
)

func StartHTTP() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/message", messageHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/oauth2", callbackHandler)
	http.HandleFunc("/logout", logoutHandler)

	log.Println("Starting HTTP Server on", Settings.HttpServer)

	// Start the server on port 8080
	err := http.ListenAndServe(Settings.HttpServer, nil)
	if err != nil {
		log.Fatal("error starting http server", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		createNewProvider(w, r)
		return
	}

	if cookie == nil {
		createNewProvider(w, r)
		return
	}

	session, err := GetSession(cookie.Value)
	if err != nil {
		createNewProvider(w, r)
		return
	}

	ctx := context.Background()
	users, err := GetUsers(ctx)
	if err != nil {
		log.Fatal("Error GetUsers", err)
	}

	messages, err := GetMessages(ctx)
	if err != nil {
		log.Fatal("Error GetMessages", err)
	}

	views.Home(users, messages, session).Render(r.Context(), w)
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
		log.Println("Message received in messageHandler:", cookie.Value, inputMessage)
		sendMessage(inputMessage, cookie.Value, nil)
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
