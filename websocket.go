package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Adjust the origin checking to suit your needs
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	// Write message back to browser
	if err := ws.WriteMessage(1, []byte("message from server")); err != nil {
		log.Println(err)
	}

	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Print the message to the console
		log.Printf("%s sent: %s\n", ws.RemoteAddr(), message)

		// Write message back to browser
		if err := ws.WriteMessage(messageType, message); err != nil {
			log.Println(err)
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
