package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

func startTerminal() {
	fmt.Println("Starting server")
	ln, err := net.Listen("tcp", "0.0.0.0:8000")

	if err != nil {
		fmt.Println("error 1")
	}
	defer ln.Close()

	users = make(map[string]User)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error 2")
		}
		go newUser(conn)
	}
}

func newUser(conn net.Conn) {
	conn.Write([]byte("What is your name?\n"))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("error 3")
	}
	name = strings.Trim(name, "\n")

	if users[name].Name == "" {
		colorCode := rand.Intn(8) + 90 // ANSI codes for foreground colors (30-37)
		color := fmt.Sprintf("\x1b[%dm", colorCode)

		users[name] = User{
			Conn:  conn,
			Name:  name,
			Color: color,
		}

		messageChannel <- Message{
			Text:      genericMessage["joined"],
			Name:      name,
			TimeStamp: time.Now(),
		}

		conn.Write([]byte("Hi " + name + ", " + genericMessage["welcome"] + "\n"))
	} else {
		if users[name].Conn != nil {
			users[name].Conn.Close()
		}

		users[name] = User{
			Conn:  conn,
			Name:  name,
			Color: users[name].Color,
		}
		messageChannel <- Message{
			Text:      "I am Back!",
			Name:      name,
			TimeStamp: time.Now(),
		}
		conn.Write([]byte("Hi " + name + ", " + genericMessage["welcomeBack"] + "\n"))
	}

	for {
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Connection terminated by user", name)
			break
		}
		response = strings.Trim(response, "\n")

		messageChannel <- Message{
			Text:      response,
			Name:      name,
			TimeStamp: time.Now(),
		}
	}
}
