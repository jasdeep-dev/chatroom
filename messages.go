package main

import (
	"time"
)

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
