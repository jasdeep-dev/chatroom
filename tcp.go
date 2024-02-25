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
		fmt.Println("error listening to TCP:", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting to TCP connection:", err)
		}
		go newUser(conn)
	}
}

func newUser(conn net.Conn) {
	conn.Write([]byte("What is your name?\n"))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("error reading message over TCP", err)
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
		conn, ok := connections[user.Name]
		if ok {
			if conn != nil {
				conn.Close()
			}
		}

		updateConnection(name, conn)
		sendMessage(conn, Settings.WelcomeBackMessage, name)
	}

	readMessages(conn, name)
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
		Name:  name,
		Color: color,
	}
	connections[name] = conn
	BackupData(users[name], "./users.db")
	return nil
}

func updateConnection(name string, conn net.Conn) {
	name = strings.TrimSpace(name)
	users[name] = User{
		Name:  name,
		Color: users[name].Color,
	}
	connections[name] = conn
}

func readMessages(conn net.Conn, name string) {
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
