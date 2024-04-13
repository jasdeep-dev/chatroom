package lib

import (
	"chatroom/app"
	"context"
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

	user, err := FindUserByID(r.Context(), session.UserID)
	if err != nil {
		log.Println("User does not exist", err)
	}

	user.IsOnline = true
	UpdateUser(r.Context(), user)

	// app.UserChannel <- user

	listenForMessages(r.Context(), sessionID, conn, r)
}

func listenForMessages(ctx context.Context, sessionID string, conn *websocket.Conn, r *http.Request) {
	log.Println("User connected and listening for messages over Socket:", sessionID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "close 1001") {
				log.Println("Session terminated by client: ", sessionID)

				session, err := GetSession(sessionID, r)
				if err != nil {
					log.Println("Session does not exist", err)
					return
				}

				user, err := FindUserByID(ctx, session.UserID)
				if err != nil {
					log.Println("User does not exist", err)
					return
				}

				user.IsOnline = false
				UpdateUser(ctx, user)

				// app.UserChannel <- user
			} else {
				log.Println("Error reading message:", err)
			}
			break
		}

		log.Printf("%s sent from browser: %s\n", conn.RemoteAddr(), message)

		if sessionID == "" {
			log.Println("sendMessage: Session id is blank")
			return
		}

		session, err := GetSession(sessionID, r)
		if err != nil {
			log.Println("sendMessage: Session not found", sessionID)
		}

		sendMessage(ctx, string(message), session, conn)
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
