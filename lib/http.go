package lib

import (
	"chatroom/lib/keycloak"
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
	http.HandleFunc("/messages", messagesHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/oauth2", callbackHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/groups", formHandler)

	log.Println("Starting HTTP Server on", Settings.HttpServer)

	err := http.ListenAndServe(Settings.HttpServer, nil)
	if err != nil {
		log.Fatal("error starting http server", err)
	}
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	groupId := r.URL.Query().Get("groupId")
	messages := keycloak.GetMessagesByGroupID(groupId)

	group, err := keycloak.GetGroupsByUserIDViaAPI(groupId)
	if err != nil {
		log.Fatal("unable to find the group for messages", err)
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Fatal("Unable to find the session cookie: ", err)
	}

	//ADD user to the Group
	session, err := GetSession(cookie.Value, r)
	if err != nil {
		log.Fatal("Unable to get the sessions struct: ", err)
	}

	views.Messages(messages, session, group).Render(r.Context(), w)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusInternalServerError)
			return
		}

		name := r.Form.Get("name")
		err = keycloak.CreateGroup(name)
		if err != nil {
			log.Fatal("Unable to create the keycloak groups: ", err)
		}

		cookie, err := r.Cookie("session_id")
		if err != nil {
			log.Fatal("Unable to find the session cookie: ", err)
		}

		//ADD user to the Group
		session, err := GetSession(cookie.Value, r)
		if err != nil {
			log.Fatal("Unable to get the sessions struct: ", err)
		}

		group, err := keycloak.FindGroupByName(context.Background(), name)
		if err != nil {
			log.Fatal("Unable to find the group: ", err)
		}

		err = keycloak.AddUserToGroup(session.KeyCloakUser, group)
		if err != nil {
			log.Printf("unable to add the users to %s group", group.Name)
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	handleError := func() {
		url := createNewProvider(w, r)

		http.Redirect(w, r, url, http.StatusFound)
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		handleError()
		return
	}

	session, err := GetSession(cookie.Value, r)
	if err != nil {
		handleError()
		return
	}

	messages, err := GetMessages(r.Context())
	if err != nil {
		log.Println("Error GetMessages in homeHandler", err)
		return
	}

	keycloak.SetAdminToken()

	keycloak_users, err := keycloak.GetUsersViaAPI()
	if err != nil {
		log.Fatal("Unable to connect to Keycloak: ", err)
	}

	groups, err := keycloak.GetUsersGroupsViaAPI(session.UserID)
	if err != nil {
		log.Fatal("Unable to find the keycloak groups: ", err)
	}

	views.Home(messages, session, keycloak_users, groups).Render(r.Context(), w)
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
		if cookie.Value == "" {
			log.Println("sendMessage: Session id is blank")
			return
		}

		// session, err := GetSession(cookie.Value, r)
		// if err != nil {
		// 	log.Println("sendMessage: Session not found", cookie.Value)
		// }

		// sendMessage(context.Background(), inputMessage, session, nil)
	}

	// http.Redirect(w, r, "/", http.StatusSeeOther)
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
