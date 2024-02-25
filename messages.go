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

func sendMessage(conn net.Conn, message string, name string) {
	name = strings.TrimSpace(name)
	messageChannel <- Message{
		Text:      message,
		Name:      name,
		TimeStamp: time.Now(),
	}
}
