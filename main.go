package main

import (
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

func main() {
	go startHTTP()
	startTerminal()
}
