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

	go receiver()
	go startWebSocket()
	startHTTP()
}

func receiver() {
	for message := range messageChannel {
		fmt.Println("Number of users: ", len(users))
		messages = append(messages, message)
		BackupData(message, "./messages.db")

		deliverMessageToWebSocketConnections(message)
	}
}

func deliverMessageToWebSocketConnections(message Message) {
	for _, userSession := range UserSessions {
		// if userSession.Name == message.Name {
		// 	continue
		// }

		if userSession.SocketConnection == nil {
			continue
		}

		userSession.SocketConnection.WriteJSON(message)
	}
}
