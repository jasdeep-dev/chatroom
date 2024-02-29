package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
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

func (m Users) Restore(row []byte) {
	var usr User
	err := json.Unmarshal(row, &usr)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	users[usr.Name] = usr
}

func createTCPUser(conn net.Conn, name string, passwordHash string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("name is empty")
	}
	//limit the users
	if len(users) >= Settings.MaxUsers {
		conn.Write([]byte("Sorry, too many users. Please try again later!\n"))
		conn.Close()
		return errors.New("too many users")
	}

	// ANSI codes for foreground colors (30-37)
	colorCode := rand.Intn(8) + 90
	color := fmt.Sprintf("\x1b[%dm", colorCode)

	users[name] = User{
		Name:         name,
		Color:        color,
		PasswordHash: passwordHash,
	}
	connections[name] = conn
	BackupData(users[name], "./users.db")
	return nil
}

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

	_, ok := users[name]

	if ok {
		// Authenticate user by comparing password with the hashed password
		err := bcrypt.CompareHashAndPassword([]byte(users[name].PasswordHash), []byte(password))
		if err != nil {
			r.Header.Set("ERROR", "Invalid password")
			loginHandler(w, r)
			return
		}
	} else {
		users[name] = User{
			Name:         name,
			PasswordHash: string(passwordHash),
		}
		BackupData(users[name], "./users.db")
	}

	sessionId, err := generateSessionID(name)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",         // Name of the cookie
		Value:    sessionId,            // Session ID generated previously
		Path:     "/",                  // Path attribute (scope of the cookie)
		SameSite: http.SameSiteLaxMode, // SameSite attribute
	})

	// fmt.Println("socketConnections[sessionId]=> ", socketConnections)

	// sendMessageToSocket(socketConnections[sessionId], "Message")
	messageChannel <- Message{
		TimeStamp: time.Now(),
		Text:      genericMessage["joined"],
		Name:      name,
	}
}

func generateSessionID(name string) (string, error) {
	b := make([]byte, 16) // 16 bytes for 128 bits of entropy
	_, err := crypt.Read(b)
	if err != nil {
		return "", err
	}

	sessionId := hex.EncodeToString(b)
	UserSessions[sessionId] = UserSession{
		Name:       name,
		LoggedInAt: time.Now(),
	}

	return sessionId, nil
}

func readCookie(w http.ResponseWriter, r *http.Request, key string) string {
	value, err := r.Cookie(key)
	if err != nil {
		log.Println(err)
	}
	return value.Value
}

func isLoggedIn() {

}
