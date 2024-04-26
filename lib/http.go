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
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

func StartHTTP() {

	mux := mux.NewRouter()

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
	mux.HandleFunc("/api/users/search", searchUsersHandler)

	mux.HandleFunc("/api/users/{userID}", UserMessagesHandler)

	// Start the server
	err := http.ListenAndServe(Settings.HttpServer, mux)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func sessionData(w http.ResponseWriter, r *http.Request) {
	app.PublicGroupID = "ba018e56-9f30-4a62-bc48-7ba3a59ce485"
	app.PersonalGroupID = "42dd286d-1c72-4e22-b703-10b269c44bbd"
	handleError := func() {
		url := createNewProvider(w, r)

		http.Redirect(w, r, url, http.StatusFound)
	}

	var err error

	cookie, err := r.Cookie("session_id")
	if err != nil {
		handleError()
		return
	}

	app.Session, err = GetSession(cookie.Value, r)
	if err != nil {
		handleError()
		return
	}
}

func setBasicData() {
	var err error

	keycloak.SetAdminToken()

	_, err = keycloak.NewKeycloakService().GetUsersViaAPI()
	if err != nil {
		log.Println("Unable to connect to Keycloak: ", err)
		return
	}

	err = keycloak.NewKeycloakService().GetUsersGroupsViaAPI(app.Session.KeyCloakUser.ID)
	if err != nil {
		log.Println("Unable to connect to Keycloak: ", err)
		return
	}
}

func UserMessagesHandler(w http.ResponseWriter, r *http.Request) {
	sessionData(w, r)
	var user app.KeyCloakUser
	var err error
	kc := keycloak.NewKeycloakService()

	vars := mux.Vars(r)
	userID := vars["userID"]

	user, err = kc.FindUserByID(userID)
	if err != nil {
		log.Println("Unable to find the error")
	}

	group := checkIfGroupExist(user.Username)

	err = kc.GetGroupMembersViaAPI(group.ID)
	if err != nil {
		log.Println("Unable to fetch the group members")
	}

	for _, usr := range app.GroupUsers {
		if usr.ID == group.Attributes["created_by"][0] {
			app.GroupAdmin = usr
		}
	}

	err = kc.AddUserToGroup(user.ID, group.ID)
	if err != nil {
		log.Println("Unable to add the user to the group", user.Username)
	}

	err = kc.AddUserToGroup(app.Session.KeyCloakUser.ID, group.ID)
	if err != nil {
		log.Println("Unable to add the user to the group: ", app.Session.KeyCloakUser.Username)
	}

	err = kc.GetUsersGroupsViaAPI(app.Session.KeyCloakUser.ID)
	if err != nil {
		log.Println("Error in fetching the groups")
	}

	messages := GetMessagesByGroupID(context.Background(), group.ID)

	subtractSlices(app.AllUsers, app.GroupUsers)

	if r.Method == http.MethodGet {
		views.Home(messages, app.Session, app.AllUsers, app.Groups, group).Render(r.Context(), w)

	}
}

func userName(username string) string {
	stringsArray := []string{username, app.Session.KeyCloakUser.Username}

	sort.Strings(stringsArray)
	concatenatedString := strings.Join(stringsArray, "_")
	return concatenatedString
}

func checkIfGroupExist(username string) app.Group {
	concatenatedString := userName(username)
	kc := keycloak.NewKeycloakService()

	var err error
	var id string
	var group app.Group

	id, err = keycloak.GetGroupByName(context.Background(), concatenatedString)
	if err != nil {
		log.Println("Group does not exist")
		err = kc.CreateGroup(concatenatedString, app.PersonalGroupID)
		if err != nil {
			log.Println("Could not create the Group")
		}
		id, err = keycloak.GetGroupByName(context.Background(), concatenatedString)
		if err != nil {
			log.Println("Could find the Group")
		}
	}

	group, err = kc.GetGroupByIDViaAPI(id)
	if err != nil {
		log.Println("Could not find the Group")
	}

	return group
}
func GroupUsersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	groupID := vars["groupID"]
	var err error

	kc := keycloak.NewKeycloakService()

	setBasicData()
	if r.Method == http.MethodPost {

		var response string

		err = kc.AddUserToGroup(userID, groupID)
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

		err = kc.RemoveUserFromGroup(userID, groupID)
		if err != nil {
			log.Println("Unable to remove the user", err)
			views.SearchBar(groupID, "Unable to remove the user").Render(r.Context(), w)
			return
		}

		err = kc.GetGroupMembersViaAPI(groupID)
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

func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("search")

	var matches []app.KeyCloakUser
	users, err := keycloak.NewKeycloakService().GetUsersViaAPI()
	if err != nil {
		log.Println("Error in fetching the users")
	}
	for _, user := range users {
		if strings.Contains(strings.ToLower(user.FirstName), strings.ToLower(query)) {
			matches = append(matches, user)
		}
	}
	views.SearchUsers(matches).Render(r.Context(), w)
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	sessionData(w, r)

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println("Error parsing form:", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		message := app.MessageData{
			Message: r.Form.Get("message"),
			GroupID: r.Form.Get("groupId"),
		}

		sendMessage(context.Background(), message, app.Session, nil, r)

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

		err = kc.GetGroupMembersViaAPI(group.ID)
		if err != nil {
			log.Println("Unable to retrieve group members:", err)
			http.Error(w, "Unable to retrieve group members", http.StatusInternalServerError)
			return
		}

		subtractSlices(app.AllUsers, app.GroupUsers)

		log.Printf("Set the Users hash for group %s with the count %v", group.Name, len(app.GroupUsers))

		views.Home(messages, app.Session, app.AllUsers, app.Groups, group).Render(r.Context(), w)
	}
}

func subtractSlices(allUsers []app.KeyCloakUser, groupUsers []app.KeyCloakUser) {

	present := make(map[string]bool)
	for _, person := range groupUsers {
		present[person.ID] = true
	}

	app.RestUsers = []app.KeyCloakUser{}
	for _, person := range allUsers {
		if _, found := present[person.ID]; !found {
			app.RestUsers = append(app.RestUsers, person)
		}
	}
	log.Println("Avaiable users to add to the group!")
}

func GroupsHandler(w http.ResponseWriter, r *http.Request) {

	sessionData(w, r)

	kc := keycloak.NewKeycloakService()

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusInternalServerError)
			return
		}

		name := r.Form.Get("name")

		err = kc.CreateGroup(name, app.PublicGroupID)
		if err != nil {
			http.Error(w, "Unable to create the groups", http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		groupsCreatedByUser, err := keycloak.GroupsCreatedByUser(ctx, app.Session.KeyCloakUser.ID)
		if err != nil {
			log.Fatal("Unable to find the group created by this user: ", err)
			return
		}

		var createdgroupID string
		for _, groupID := range groupsCreatedByUser {
			if !slices.Contains(app.GroupIds, groupID) {
				createdgroupID = groupID
				err = kc.AddUserToGroup(app.Session.KeyCloakUser.ID, groupID)
				if err != nil {
					log.Println("Unable to find the group: ", err)
				}

			}
		}
		responseUrl := fmt.Sprint("/api/groups/", createdgroupID)
		w.Write([]byte(responseUrl))
	} else if r.Method == http.MethodGet {
		setBasicData()
		vars := mux.Vars(r)
		groupID := vars["groupID"]

		var group app.Group
		var err error
		var messages []app.Message

		var groupIds []string
		for _, group = range app.Groups {
			groupIds = append(groupIds, group.ID)
		}

		if !slices.Contains(groupIds, groupID) {
			fmt.Println("User does not have the acess to this group")
			views.Home(messages, app.Session, app.AllUsers, app.Groups, app.Groups[0]).Render(r.Context(), w)
			return
		}

		messages = GetMessagesByGroupID(context.Background(), groupID)

		group, err = kc.GetGroupByIDViaAPI(groupID)
		if err != nil {
			return
		}

		err = kc.GetGroupMembersViaAPI(group.ID)
		if err != nil {
			log.Printf("Error getting groups: %v", err)
		}

		for _, usr := range app.GroupUsers {
			if usr.ID == group.Attributes["created_by"][0] {
				app.GroupAdmin = usr
			}
		}

		subtractSlices(app.AllUsers, app.GroupUsers)

		views.Home(messages, app.Session, app.AllUsers, app.Groups, group).Render(r.Context(), w)
	} else if r.Method == http.MethodDelete {
		var err error
		vars := mux.Vars(r)
		groupID := vars["groupID"]
		err = kc.DeleteGroupViaAPI(groupID)
		if err != nil {
			return
		}

		setBasicData()

		err = kc.GetUsersGroupsViaAPI(app.Session.UserID)
		if err != nil {
			log.Printf("Error getting groups: %v", err)
		}
		messages := GetMessagesByGroupID(context.Background(), groupID)

		views.Home(messages, app.Session, app.AllUsers, app.Groups, app.Groups[0]).Render(r.Context(), w)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	app.PublicGroupID = "ba018e56-9f30-4a62-bc48-7ba3a59ce485"
	app.PersonalGroupID = "42dd286d-1c72-4e22-b703-10b269c44bbd"
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

	err = kc.GetUsersGroupsViaAPI(session.UserID)
	if err != nil {
		log.Printf("Error getting groups: %v", err)
	}

	var group app.Group

	if len(app.Groups) != 0 {
		group, err := kc.GetGroupByIDViaAPI(app.Groups[0].ID)
		if err != nil {
			log.Printf("Error getting group by ID: %v", err)
		}

		app.GroupAdmin, err = kc.FindUserByID(group.Attributes["created_by"][0])
		if err != nil {
			log.Printf("Error getting user by ID: %v", err)
		}
	}

	views.Home(messages, session, keycloak_users, app.Groups, group).Render(r.Context(), w)
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
