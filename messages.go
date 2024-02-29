package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

func (m Messages) Restore(row []byte) {
	var msg Message
	err := json.Unmarshal(row, &msg)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	messages = append(messages, msg)
}

func sendMessage(message string, sessionID string) {
	if message == "" {
		return
	}

	if sessionID == "" {
		return
	}

	session := UserSessions[sessionID]
	messageChannel <- Message{
		Text:      message,
		Name:      session.Name,
		TimeStamp: time.Now(),
	}
}

func sendMessageTCP(conn net.Conn, message string, name string) {
	fmt.Println("message", message)
	if message == "" {
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	messageChannel <- Message{
		Text:      message,
		Name:      name,
		TimeStamp: time.Now(),
	}
}
