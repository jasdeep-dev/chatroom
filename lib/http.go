package lib

import (
	"chatroom/app"
	"chatroom/lib/keycloak"
	"chatroom/views"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
)

func StartHTTP() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/messages/create", messageHandler)
	http.HandleFunc("/messages", messagesHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/oauth2", callbackHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/groups", formHandler)
	http.HandleFunc("/api/search", searchHandler)
	http.HandleFunc("/addUser", AddUserToGroupHandler)
	http.HandleFunc("/removeUser", RemoveUserFromGroupHandler)

	log.Println("Starting HTTP Server on", Settings.HttpServer)

	err := http.ListenAndServe(Settings.HttpServer, nil)
	if err != nil {
		log.Fatal("error starting http server", err)
	}
}

func AddUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	userIds, ok := queryValues["userId"]
	if !ok || len(userIds[0]) < 1 {
		http.Error(w, "groupId parameter is required", http.StatusBadRequest)
		return
	}

	groupIds, ok := queryValues["groupID"]
	if !ok || len(groupIds[0]) < 1 {
		http.Error(w, "groupId parameter is required", http.StatusBadRequest)
		return
	}

	err = keycloak.AddUserToGroup(userIds[0], groupIds[0])
	if err != nil {
		log.Println("Error Adding user to the group")
	}

	user, err := keycloak.FindUserByID(userIds[0])
	if err != nil {
		log.Println("Unable to parse the user")
	}

	app.GroupUsers = append(app.GroupUsers, user)
	views.UsersList(app.GroupUsers, groupIds[0]).Render(r.Context(), w)
}

func RemoveUserFromGroupHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	// Extract the groupId parameter from the query
	userIds, ok := queryValues["userId"]
	if !ok || len(userIds[0]) < 1 {
		http.Error(w, "groupId parameter is required", http.StatusBadRequest)
		return
	}

	groupIds, ok := queryValues["groupID"]
	if !ok || len(groupIds[0]) < 1 {
		http.Error(w, "groupId parameter is required", http.StatusBadRequest)
		return
	}

	err = keycloak.RemoveUserFromGroup(userIds[0], groupIds[0])
	if err != nil {
		log.Println("User removed from Group", err)
	}

	app.GroupUsers, err = keycloak.GetGroupMembersViaAPI(groupIds[0])
	if err != nil {
		log.Println("Unable to parse the user", err)
	}

	subtractSlices(app.AllUsers, app.GroupUsers)

	views.SearchBar(groupIds[0]).Render(r.Context(), w)
	// views.UsersList(app.GroupUsers, groupIds[0]).Render(r.Context(), w)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("search")
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	// Extract the groupId parameter from the query
	groupIds, ok := queryValues["groupId"]
	if !ok || len(groupIds[0]) < 1 {
		http.Error(w, "groupId parameter is required", http.StatusBadRequest)
		return
	}

	var matches []app.KeyCloakUser
	for _, user := range app.RestUsers {
		if strings.Contains(strings.ToLower(user.FirstName), strings.ToLower(query)) {
			matches = append(matches, user)
		}
	}
	views.SearchedUsers(matches, groupIds[0]).Render(r.Context(), w)
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	groupId := r.URL.Query().Get("groupId")
	messages := GetMessagesByGroupID(groupId)

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

	app.GroupUsers, err = keycloak.GetGroupMembersViaAPI(group.ID)
	if err != nil {
		log.Printf("unable to find the users for %s group", group.Name)
	}

	subtractSlices(app.AllUsers, app.GroupUsers)

	log.Printf("Set the Users hash for group %s with the count %v", group.Name, len(app.GroupUsers))

	views.Messages(messages, session, group).Render(r.Context(), w)
}

func subtractSlices(slice1 []app.KeyCloakUser, slice2 []app.KeyCloakUser) {

	// Create a map to track names in slice2 for quick lookup
	present := make(map[string]bool)
	for _, person := range slice2 {
		present[person.ID] = true
	}
	app.RestUsers = []app.KeyCloakUser{}
	// Add to result only those from slice1 not in slice2
	for _, person := range slice1 {
		if _, found := present[person.ID]; !found {
			app.RestUsers = append(app.RestUsers, person)
		}
	}
	fmt.Println("Avaiable users to add to the group!")
}
func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusInternalServerError)
			return
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

		name := r.Form.Get("name")
		err = keycloak.CreateGroup(name, session.KeyCloakUser.ID)
		if err != nil {
			log.Fatal("Unable to create the keycloak groups: ", err)
		}

		ctx := context.Background()
		groupsCreatedByUser, err := keycloak.GroupsCreatedByUser(ctx, session.KeyCloakUser.ID)
		if err != nil {
			log.Fatal("Unable to find the group created by this user: ", err)
		}

		for _, groupID := range groupsCreatedByUser {
			if !slices.Contains(app.GroupIds, groupID) {
				err = keycloak.AddUserToGroup(session.KeyCloakUser.ID, groupID)
				if err != nil {
					log.Println("Unable to find the group: ", err)
				}
			}
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

	kc := keycloak.NewKeycloakService(
		os.Getenv("ADMIN_ACCESS_TOKEN"),
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
	)

	Groups, err := kc.GetUsersGroupsViaAPI(session.UserID)
	if err != nil {
		log.Printf("Error getting groups: %v", err)
	}

	// Groups, err := keycloak.GetUsersGroupsViaAPI(session.UserID)
	// if err != nil {
	// 	log.Fatal("Unable to find the keycloak groups: ", err)
	// }

	views.Home(messages, session, keycloak_users, Groups).Render(r.Context(), w)
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

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("Error in cookies", err)
		return
	}

	if cookie.Value == "" {
		log.Println("sendMessage: Session id is blank")
		return
	}

	session, err := GetSession(cookie.Value, r)
	if err != nil {
		log.Println("sendMessage: Session not found", cookie.Value)
		return
	}

	message := app.MessageData{
		Message: r.Form.Get("message"),
		GroupID: r.Form.Get("groupId"),
	}
	log.Println("Message received in messageHandler:", cookie.Value, message)

	// Send the to all users in the group via websockets
	sendMessage(context.Background(), message, session, nil, w, r)

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
