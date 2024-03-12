package lib

import (
	"chatroom/app"
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

	log.Println("HTTP Server listening on", Settings.HttpServer)
	err := http.ListenAndServe(Settings.HttpServer, nil) // Start the server on port 8080
	if err != nil {
		log.Fatal("error starting http server", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// tmpl, err := template.New("html").Funcs(template.FuncMap{
	// 	"formatTime": formatTime,
	// }).ParseGlob("./views/app/*.html")

	// if err != nil {
	// 	log.Println("can't parse the files", err)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("error in cookies", err)
	}

	if cookie == nil {
		createNewProvider(w, r)
		return
	}

	session := app.UserSessions[cookie.Value]

	ctx := context.Background()
	users, err := GetUsers(ctx)
	if err != nil {
		log.Println("Error GetUsers", err)
	}

	messages, err := GetMessages(ctx)
	if err != nil {
		log.Println("Error GetMessages", err)
	}
	// data := app.TemplateData{
	// 	Users:       users,
	// 	Messages:    app.MessagesArray,
	// 	CurrentUser: session.KeyCloakUser,
	// 	LoggedIn:    session.LoggedInAt,
	// }

	// err = tmpl.ExecuteTemplate(w, "index.html", data)
	// if err != nil {
	// 	log.Println("Unable to render templates.", err)
	// 	w.Write([]byte(err.Error()))
	// }

	views.Home(users, messages, session).Render(r.Context(), w)
}

// func formatTime(t time.Time) string {
// 	return t.Format(time.RFC3339)
// }

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
