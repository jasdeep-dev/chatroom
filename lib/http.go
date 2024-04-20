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
	"slices"
	"strings"
)

func StartHTTP() {

	mux := http.NewServeMux()

	//Serve static files from the public directory
	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("/public/", http.StripPrefix("/public/", fs))

	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/oauth2", callbackHandler)
	mux.HandleFunc("/logout", logoutHandler)

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/users", userHandler)
	mux.HandleFunc("/api/messages", MessageHandler)

	mux.HandleFunc("/api/groups", GroupsHandler)
	mux.HandleFunc("/api/search", searchHandler)
	mux.HandleFunc("/addUser", AddUserToGroupHandler)
	mux.HandleFunc("/removeUser", RemoveUserFromGroupHandler)

	// Start the server
	err := http.ListenAndServe(Settings.HttpServer, mux) // Start the server
	if err != nil {
		log.Fatal("ListenAndServe:", err)
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

	kc := keycloak.NewKeycloakService()

	err = kc.AddUserToGroup(userIds[0], groupIds[0])
	if err != nil {
		log.Println("Error Adding user to the group")
	}

	user, err := kc.FindUserByID(userIds[0])
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

	kc := keycloak.NewKeycloakService()

	err = kc.RemoveUserFromGroup(userIds[0], groupIds[0])
	if err != nil {
		log.Println("User removed from Group", err)
	}

	app.GroupUsers, err = kc.GetGroupMembersViaAPI(groupIds[0])
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

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve session ID from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("Error retrieving session ID from cookie:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check if session ID is empty
	if cookie.Value == "" {
		log.Println("sendMessage: Session id is blank")
		http.Error(w, "Session ID is blank", http.StatusBadRequest)
		return
	}

	// Get session using session ID
	session, err := GetSession(cookie.Value, r)
	if err != nil {
		log.Println("sendMessage: Session not found:", cookie.Value)
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	// Handle POST request
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println("Error parsing form:", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Extract message data from form
		message := app.MessageData{
			Message: r.Form.Get("message"),
			GroupID: r.Form.Get("groupId"),
		}

		// Send message
		sendMessage(context.Background(), message, session, nil, w, r)

	} else if r.Method == http.MethodGet { // Handle GET request
		// Extract group ID from URL query parameter
		groupID := r.URL.Query().Get("groupId")

		// Retrieve messages for the group ID
		messages := GetMessagesByGroupID(groupID)

		// Retrieve group information
		kc := keycloak.NewKeycloakService()
		group, err := kc.GetGroupByIDViaAPI(groupID)
		if err != nil {
			log.Fatal("Unable to find the group for messages:", err)
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		// Retrieve group members
		app.GroupUsers, err = kc.GetGroupMembersViaAPI(group.ID)
		if err != nil {
			log.Println("Unable to retrieve group members:", err)
			http.Error(w, "Unable to retrieve group members", http.StatusInternalServerError)
			return
		}

		// Remove group users from all users
		subtractSlices(app.AllUsers, app.GroupUsers)

		// Log the count of group users
		log.Printf("Set the Users hash for group %s with the count %v", group.Name, len(app.GroupUsers))

		// Render the messages view
		err = views.Messages(messages, session, group).Render(r.Context(), w)
		if err != nil {
			log.Println("Error rendering messages view:", err)
			http.Error(w, "Error rendering messages view", http.StatusInternalServerError)
			return
		}
	}
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
func GroupsHandler(w http.ResponseWriter, r *http.Request) {
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

		kc := keycloak.NewKeycloakService()

		err = kc.CreateGroup(name, session.KeyCloakUser.ID)
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
				err = kc.AddUserToGroup(session.KeyCloakUser.ID, groupID)
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

	kc := keycloak.NewKeycloakService()

	messages, err := GetMessages(r.Context())
	if err != nil {
		log.Println("Error GetMessages in homeHandler", err)
		return
	}

	keycloak.SetAdminToken()

	keycloak_users, err := kc.GetUsersViaAPI()
	if err != nil {
		log.Fatal("Unable to connect to Keycloak: ", err)
	}

	Groups, err := kc.GetUsersGroupsViaAPI(session.UserID)
	if err != nil {
		log.Printf("Error getting groups: %v", err)
	}

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
