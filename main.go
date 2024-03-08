package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type User struct {
	ID                int    `sql:"id"`
	Name              string `sql:"name"`
	IsOnline          bool   `sql:"is_online"`
	Theme             string `sql:"theme"`
	PreferredUsername string `sql:"preferred_username"`
	GivenName         string `sql:"given_name"`
	FamilyName        string `sql:"family_name"`
	Email             string `sql:"email"`
}

type user map[int]User

var Users = make(user)

type Messages []Message

var MessagesArray Messages

var messages Messages

type Message struct {
	TimeStamp time.Time
	Text      string
	Name      string
	Email     string
}

var genericMessage map[string]string
var messageChannel = make(chan Message, 100)

var DBConn *pgxpool.Pool

var userChannel = make(chan User, 100)

func init() {
	// Initialize the map inside an init function
	genericMessage = make(map[string]string)
	genericMessage["joined"] = "I have joined the chat."
	genericMessage["welcome"] = "Welcome to chatroom."
	genericMessage["welcomeBack"] = "Welcome back!"
}

func main() {
	// Read config file
	err := readConfigFromFile("./config.json")
	if err != nil {
		log.Fatal("Could not read the config file: ", err)
	}

	// Connect to database
	err = godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", err)
	}

	ctx := context.Background()
	DBConn = establishConnection(ctx)
	defer DBConn.Close()

	// migrateDatabase(ctx)
	getUsers(ctx)
	getMessages(ctx)

	go messageReceiver()
	go userReciver()
	go startWebSocket()
	startHTTP()
}

func messageReceiver() {
	for message := range messageChannel {
		InsertMessage(message)
		MessagesArray = append(MessagesArray, message)

		messages = append(messages, message)

		deliverMessageToWebSocketConnections(message)
	}
}

func userReciver() {
	for user := range userChannel {
		if user.Name != "" {
			deliverUsersToWebSocketConnections(user)
		}
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
