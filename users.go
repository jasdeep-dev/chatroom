package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	crypt "crypto/rand"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

type UserSession struct {
	Name             string
	LoggedInAt       time.Time
	SocketConnection *websocket.Conn
}

var UserSessions = make(map[string]UserSession)

func createHTTPUser(w http.ResponseWriter, r *http.Request) {
	name := r.Form.Get("name")
	password := r.Form.Get("password")

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		r.Header.Set("ERROR", "Unable to create user")
		loginHandler(w, r)
		return
	}

	user, ok := users[name]

	if ok {
		// Authenticate user by comparing password with the hashed password
		err := bcrypt.CompareHashAndPassword([]byte(users[name].PasswordHash), []byte(password))
		if err != nil {
			r.Header.Set("ERROR", "Invalid password")
			loginHandler(w, r)
			return
		}

		user.IsOnline = true
		users[name] = user
	} else {

		users[name] = User{
			Name:         name,
			IsOnline:     true,
			PasswordHash: string(passwordHash),
		}

		insertUser(users[name])
	}

	sessionId, err := generateSessionID(name)
	if err != nil {
		fmt.Println("Error creating Session ID", err)
	}

	UserSessions[sessionId] = UserSession{
		Name:       name,
		LoggedInAt: time.Now(),
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",         // Name of the cookie
		Value:    sessionId,            // Session ID generated previously
		Path:     "/",                  // Path attribute (scope of the cookie)
		SameSite: http.SameSiteLaxMode, // SameSite attribute
	})

	messageChannel <- Message{
		TimeStamp: time.Now(),
		Text:      genericMessage["joined"],
		Name:      name,
	}
}

func generateSessionID(name string) (string, error) {
	b := make([]byte, 128) // 16 bytes for 128 bits of entropy
	_, err := crypt.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func insertUser(userData User) {
	ctx := context.Background()
	query := `INSERT INTO users (name, is_online, theme, password_hash) VALUES ($1, $2, $3, $4)`
	_, err := DBConn.Exec(ctx, query, userData.Name, userData.IsOnline, userData.Theme, userData.PasswordHash)
	if err != nil {
		fmt.Println("Error inserting data into users table:", err)
		return
	}
}
