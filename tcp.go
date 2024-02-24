package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

func startTerminal() {
	fmt.Println("TCP Server listening on", Settings.TcpServer)

	ln, err := net.Listen("tcp", Settings.TcpServer)

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
	name = strings.TrimSpace(name)

	user, ok := users[name]
	if !ok {
		err := createNewUser(conn, name)
		if err != nil {
			conn.Write([]byte("error in creating users" + err.Error()))
			conn.Close()
			return
		}

		sendMessage(conn, Settings.JoinedMessage, name)
	} else {
		if user.Conn != nil {
			user.Conn.Close()
		}

		updateConnection(name, conn)
		sendMessage(conn, Settings.WelcomeBackMessage, name)
	}

	for {
		ReadMessage(conn, name)
	}
}

func createNewUser(conn net.Conn, name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("name is empty")
	}
	//limit the users
	if len(users) >= Settings.MaxUsers {
		conn.Write([]byte("Sorry, too many users. Please try again later!\n"))
		conn.Close()
		return errors.New("too many users")
	}

	// ANSI codes for foreground colors (30-37)
	colorCode := rand.Intn(8) + 90
	color := fmt.Sprintf("\x1b[%dm", colorCode)

	users[name] = User{
		Conn:  conn,
		Name:  name,
		Color: color,
	}
	return nil
}

func updateConnection(name string, conn net.Conn) {
	name = strings.TrimSpace(name)
	users[name] = User{
		Conn:  conn,
		Name:  name,
		Color: users[name].Color,
	}
}

func sendMessage(conn net.Conn, message string, name string) {
	name = strings.TrimSpace(name)
	messageChannel <- Message{
		Text:      message,
		Name:      name,
		TimeStamp: time.Now(),
	}
}

func ReadMessage(conn net.Conn, name string) {

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		if len(line) > Settings.MessageSize {
			fmt.Fprintf(conn, "Maximum word limit (250 words) reached.\n")
			return
		}

		messageChannel <- Message{
			Text:      line,
			Name:      name,
			TimeStamp: time.Now(),
		}

	}
}
