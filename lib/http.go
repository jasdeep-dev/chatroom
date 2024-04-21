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

	"github.com/gorilla/mux"
)

func StartHTTP() {

	mux := mux.NewRouter()

	keycloak.SetAdminToken()

	var err error
	_, err = keycloak.NewKeycloakService().GetUsersViaAPI()
	if err != nil {
		log.Fatal("Unable to connect to Keycloak: ", err)
	}

	_, err = keycloak.NewKeycloakService().GetGroupsViaAPI()
	if err != nil {
		log.Fatal("Unable to connect to Keycloak: ", err)
	}

	//Serve static files from the public directory
	fs := http.FileServer(http.Dir("./public"))
	mux.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/oauth2", callbackHandler)
	mux.HandleFunc("/logout", logoutHandler)

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/users", userHandler)
	mux.HandleFunc("/api/users/{userID}/groups/{groupID}", GroupUsersHandler)

	mux.HandleFunc("/api/messages", MessageHandler)

	mux.HandleFunc("/api/groups", GroupsHandler)
	mux.HandleFunc("/api/groups/{groupID}", GroupsHandler)

	mux.HandleFunc("/api/search", searchHandler)

	// Start the server
	err = http.ListenAndServe(Settings.HttpServer, mux)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func GroupUsersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	groupID := vars["groupID"]
	var err error

	kc := keycloak.NewKeycloakService()

	if r.Method == http.MethodPost {

		var response string
		err = kc.AddUserToGroup(userID, "groupID")
		if err != nil {
			log.Println(response, err)
			views.UsersList(app.GroupUsers, groupID, "Unable to add the user to group").Render(r.Context(), w)
			return
		}

		user, err := kc.FindUserByID(userID)
		if err != nil {
			log.Println("Unable to find the user:", err)
			views.UsersList(app.GroupUsers, groupID, "Unable to add the user to group").Render(r.Context(), w)
			return
		}

		app.GroupUsers = append(app.GroupUsers, user)

		views.UsersList(app.GroupUsers, groupID, response).Render(r.Context(), w)

	} else if r.Method == http.MethodDelete {

		err = kc.RemoveUserFromGroup(userID, "groupID")
		if err != nil {
			log.Println("Unable to remove the user", err)
			views.SearchBar(groupID, "Unable to remove the user").Render(r.Context(), w)
			return
		}

		app.GroupUsers, err = kc.GetGroupMembersViaAPI(groupID)
		if err != nil {
			log.Println("Unable to fetch the group members", err)
			views.SearchBar(groupID, "Unable to fetch the group members").Render(r.Context(), w)
			return
		}

		subtractSlices(app.AllUsers, app.GroupUsers)

		views.SearchBar(groupID, "").Render(r.Context(), w)
	}
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

	} else if r.Method == http.MethodGet {
		groupID := r.URL.Query().Get("groupId")

		messages := GetMessagesByGroupID(context.Background(), groupID)

		kc := keycloak.NewKeycloakService()
		group, err := kc.GetGroupByIDViaAPI(groupID)
		if err != nil {
			log.Fatal("Unable to find the group for messages:", err)
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		app.GroupUsers, err = kc.GetGroupMembersViaAPI(group.ID)
		if err != nil {
			log.Println("Unable to retrieve group members:", err)
			http.Error(w, "Unable to retrieve group members", http.StatusInternalServerError)
			return
		}

		subtractSlices(app.AllUsers, app.GroupUsers)

		log.Printf("Set the Users hash for group %s with the count %v", group.Name, len(app.GroupUsers))

		views.Home(messages, session, app.AllUsers, app.Groups, group).Render(r.Context(), w)
	}
}

func subtractSlices(slice1 []app.KeyCloakUser, slice2 []app.KeyCloakUser) {

	present := make(map[string]bool)
	for _, person := range slice2 {
		present[person.ID] = true
	}
	app.RestUsers = []app.KeyCloakUser{}
	for _, person := range slice1 {
		if _, found := present[person.ID]; !found {
			app.RestUsers = append(app.RestUsers, person)
		}
	}
	fmt.Println("Avaiable users to add to the group!")
}

func GroupsHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusInternalServerError)
			return
		}

		name := r.Form.Get("name")

		err = kc.CreateGroup(name, session.KeyCloakUser.ID)
		if err != nil {
			http.Error(w, "Unable to create the groups", http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		groupsCreatedByUser, err := keycloak.GroupsCreatedByUser(ctx, session.KeyCloakUser.ID)
		if err != nil {
			log.Fatal("Unable to find the group created by this user: ", err)
			return
		}

		for _, groupID := range groupsCreatedByUser {
			if !slices.Contains(app.GroupIds, groupID) {
				err = kc.AddUserToGroup(session.KeyCloakUser.ID, groupID)
				if err != nil {
					log.Println("Unable to find the group: ", err)
				}
			}
		}
	} else if r.Method == http.MethodGet {

		vars := mux.Vars(r)
		groupID := vars["groupID"]

		var group app.Group
		messages := GetMessagesByGroupID(context.Background(), groupID)
		group, err = kc.GetGroupByIDViaAPI(groupID)

		if err != nil {
			handleError()
			return
		}

		views.Home(messages, session, app.AllUsers, app.Groups, group).Render(r.Context(), w)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

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

	app.Groups, err = kc.GetUsersGroupsViaAPI(session.UserID)
	if err != nil {
		log.Printf("Error getting groups: %v", err)
	}

	views.Home(messages, session, app.AllUsers, app.Groups, app.Groups[0]).Render(r.Context(), w)
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
