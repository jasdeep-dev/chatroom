package lib

import (
	"chatroom/app"
	"chatroom/lib/keycloak"
	"context"
	"encoding/json"
	"fmt"
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

	session, err := GetSession(sessionID, r)
	if err != nil {
		log.Println("User session not found for Socket connection:", sessionID)
		return
	}

	app.SocketConnections = append(app.SocketConnections, conn)
	// TODO: Remove the socket connection from app.SocketConnections when connection terminated
	log.Println("New Socket Connection:", conn.RemoteAddr())

	user, err := keycloak.NewKeycloakService().FindUserByID(session.UserID)
	if err != nil {
		log.Println("User does not exist", err)
	}

	app.KUserChannel <- user

	listenForMessages(r.Context(), sessionID, conn, w, r)
}

func listenForMessages(ctx context.Context, sessionID string, conn *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	log.Println("User connected and listening for messages over Socket:", sessionID)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				log.Println("Session terminated by client: ", sessionID)

				session, err := GetSession(sessionID, r)
				if err != nil {
					log.Println("Session does not exist", err)
					return
				}

				user, err := keycloak.NewKeycloakService().FindUserByID(session.UserID)
				if err != nil {
					log.Println("User does not exist", err)
					return
				}

				app.KUserChannel <- user
			} else {
				log.Println("Error reading message:", err)
			}
			break
		}

		var messageData app.MessageData

		err = json.Unmarshal([]byte(data), &messageData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		log.Printf("%s sent from browser: %s\n", conn.RemoteAddr(), messageData)

		if sessionID == "" {
			log.Println("sendMessage: Session id is blank")
			return
		}

		session, err := GetSession(sessionID, r)
		if err != nil {
			log.Println("sendMessage: Session not found", sessionID)
		}

		sendMessage(ctx, messageData, session, conn, w, r)
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
