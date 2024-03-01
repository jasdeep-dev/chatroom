package main

import (
	"fmt"
	"log"
	"time"
)

type Users map[string]User

var users = make(Users)

type User struct {
	Name         string
	IsOnline     bool
	Color        string
	PasswordHash string
}

type Messages []Message

var messages Messages

type Message struct {
	TimeStamp time.Time
	Text      string
	Name      string
}

var genericMessage map[string]string
var messageChannel = make(chan Message, 100)

var userChannel = make(chan User, 100)

func init() {
	// Initialize the map inside an init function
	genericMessage = make(map[string]string)
	genericMessage["joined"] = "I have joined the chat."
	genericMessage["welcome"] = "Welcome to chatroom."
	genericMessage["welcomeBack"] = "Welcome back!"
}
func main() {
	err := readConfigFromFile("./config.json")
	if err != nil {
		log.Fatal("Could not read the config file: ", err)
	}

	RestoreData(users, "./users.db")
	RestoreData(messages, "./messages.db")

	go messageReceiver()
	go userReciver()
	go startWebSocket()
	startHTTP()
}

func messageReceiver() {
	for message := range messageChannel {
		fmt.Println("Number of users: ", len(users))
		messages = append(messages, message)
		BackupData(message, "./messages.db")

		deliverMessageToWebSocketConnections(message)
	}
}

func userReciver() {
	for user := range userChannel {
		deliverUsersToWebSocketConnections(user)
	}
}

func deliverMessageToWebSocketConnections(message Message) {
	for _, userSession := range UserSessions {
		if userSession.SocketConnection == nil {
			continue
		}
		msg := struct {
			Message Message
		}{
			Message: message,
		}
		userSession.SocketConnection.WriteJSON(msg)
	}
}

func deliverUsersToWebSocketConnections(user User) {
	for _, userSession := range UserSessions {
		if userSession.SocketConnection == nil {
			continue
		}

		usr := struct {
			User User
		}{
			User: user,
		}
		userSession.SocketConnection.WriteJSON(usr)
	}
}
