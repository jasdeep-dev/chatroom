package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

var users map[string]User

type User struct {
	Conn  net.Conn
	Name  string
	Color string
}

var messages []Message

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
	go receiver()

	go startHTTP()
	startTerminal()
}

func receiver() {
	for message := range messageChannel {
		fmt.Println("Number of users: ", len(users))
		messages = append(messages, message)

		for _, user := range users {
			if user.Name != message.Name && user.Conn != nil {
				user.Conn.Write(
					[]byte(users[message.Name].Color + message.Name + "> \x1b[0m" + message.Text + "\n"),
				)
			}
		}
	}
}
