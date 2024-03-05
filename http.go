package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
	http.HandleFunc("/logout", logoutHandler)

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
		createNewProvider(w, r)
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
	// createNewProvider(w, r)
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Define the URL for logout endpoint
	logoutURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
	)

	refreshToken, error := r.Cookie("refresh_token")
	if error != nil {
		fmt.Println("Error in cookies", error)
	}

	accessToken, error := r.Cookie("access_token")
	if error != nil {
		fmt.Println("Error in cookies", error)
	}

	// Create a map to hold the form data
	formData := url.Values{}
	formData.Set("client_id", os.Getenv("CLIENT_ID"))
	formData.Set("client_secret", os.Getenv("CLIENT_SECRET"))
	formData.Set("refresh_token", refreshToken.Value)

	// Create a new HTTP POST request with the form data
	req, err := http.NewRequest("POST", logoutURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	// Set the Content-Type header to application/x-www-form-urlencoded
	req.Header.Set("Authorization", "Bearer "+accessToken.Value)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Println("err", resp.Status, body)
	} else {
		DeleteCookie("session_id", w)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func DeleteCookie(name string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		MaxAge: -1,
		Path:   "/",
	})
}
