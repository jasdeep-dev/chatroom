package lib

import (
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

	session, ok := UserSessions[sessionID]
	if !ok {
		return
	}

	session.SocketConnection = conn
	UserSessions[sessionID] = session

	if user, exists := Users[session.ID]; exists {
		user.IsOnline = true
		Users[session.ID] = user
		UpdateUser(user)
	} else {
		user, err := FindUserByID(session.ID)
		if err != nil {
			log.Fatal("User does not exist", err)
		}

		Users[session.ID] = user
	}

	UserChannel <- Users[session.ID]

	listenForMessages(r)
}

func listenForMessages(r *http.Request) {
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println(err)
		return
	}
	sessionID := sessionCookie.Value

	conn := UserSessions[sessionID].SocketConnection
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				log.Println("Session terminated by client: ", sessionID)

				user := Users[UserSessions[sessionID].ID]
				user.IsOnline = false
				Users[UserSessions[sessionID].ID] = user
				UserChannel <- Users[UserSessions[sessionID].ID]

				UpdateUser(user)

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
