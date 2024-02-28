package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Adjust the origin checking to suit your needs
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	session, ok := UserSessions[sessionCookie.Value]
	if !ok {
		return
	}

	session.SocketConnection = conn

	conn.WriteJSON(session)

	go writeTestMessages(session)
	listenForMessages(sessionCookie.Value, session)
}

func writeTestMessages(session UserSession) {
	for {
		session.SocketConnection.WriteJSON("just a test")
		time.Sleep(2 * time.Second)
	}
}

func listenForMessages(sessionID string, session UserSession) {
	for {
		messageType, message, err := session.SocketConnection.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				log.Println("Session terminated by client: ", sessionID, session.Name)
				// delete(UserSessions, *sessionID)
			} else {
				log.Println("Error reading message:", err)
			}

			break
		}

		log.Printf("%s sent: %s\n", session.SocketConnection.RemoteAddr(), message)

		// Write message back to browser
		if err := session.SocketConnection.WriteMessage(messageType, message); err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}

func startWebSocket() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket) // Register the WebSocket handler with the ServeMux

	log.Println("Starting server on :3000")
	err := http.ListenAndServe(":3000", mux) // Start the server
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

func sendMessageToSocket(ws *websocket.Conn, message string) {
	if err := ws.WriteMessage(1, []byte(message)); err != nil {
		log.Println(err)
	}
}
