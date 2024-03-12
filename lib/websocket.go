package lib

import (
	"chatroom/app"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Adjust the origin checking to suit your needs
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
	}()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println(err)
		return
	}
	sessionID := sessionCookie.Value

	session, err := GetSession(sessionID)
	if err != nil {
		log.Println("User session not found for Socket connection:", sessionID)
		return
	}

	// TODO: this is a bug coz with the same session id a user can have any number of socket connections
	app.SocketConnections[sessionID] = conn

	user, err := FindUserByID(session.UserID)
	if err != nil {
		log.Fatal("User does not exist", err)
	}

	user.IsOnline = true
	UpdateUser(user)

	app.UserChannel <- user

	listenForMessages(r)
}

func listenForMessages(r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println(err)
		return
	}
	sessionID := sessionCookie.Value
	log.Println("User connected and listening for messages over Socket:", sessionID)

	conn := app.SocketConnections[sessionID]
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				log.Println("Session terminated by client: ", sessionID)

				session, err := GetSession(sessionID)
				if err != nil {
					log.Println("Session does not exist", err)
					return
				}

				user, err := FindUserByID(session.UserID)
				if err != nil {
					log.Println("User does not exist", err)
					return
				}

				user.IsOnline = false
				UpdateUser(user)

				app.UserChannel <- user
			} else {
				log.Println("Error reading message:", err)
			}
			break
		}

		log.Printf("%s sent from browser: %s\n", conn.RemoteAddr(), message)

		sendMessage(string(message), sessionID)
	}
}

func StartWebSocket() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket) // Register the WebSocket handler with the ServeMux

	log.Println("Starting server on :3000")
	err := http.ListenAndServe(":3000", mux) // Start the server
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
