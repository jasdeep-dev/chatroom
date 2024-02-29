package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	var name string
	var err error
	for {
		conn.Write([]byte("Enter username:\n"))
		name, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("error reading message over TCP", err)
		}
		name = strings.TrimSpace(name)

		if name == "" {
			conn.Write([]byte("Error! Name can't be blank\n"))
			continue
		} else {
			break
		}
	}

	var (
		password     string
		passwordHash string
	)

	for {
		conn.Write([]byte("Enter password:\n"))
		password, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("error reading message over TCP", err)
			conn.Write([]byte("Error! Password can't be blank\n"))
			continue
		}
		password = strings.TrimSpace(password)

		if password == "" {
			conn.Write([]byte("Error! Password can't be blank\n"))
			continue
		} else {
			hsh, err := bcrypt.GenerateFromPassword(
				[]byte(password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				conn.Write([]byte("Error! Unable to use this password\n"))
				continue
			}

			passwordHash = string(hsh)
			break
		}
	}

	user, ok := users[name]
	if ok {
		// Authenticate user by comparing password with the hashed password
		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			fmt.Println("Error CompareHashAndPassword:", err)
			conn.Write([]byte("invalid password\n"))
			return
		}

		conn, ok := connections[user.Name]
		if ok {
			if conn != nil {
				conn.Close()
			}
		}

		updateConnection(name, conn)
		sendMessageTCP(conn, Settings.WelcomeBackMessage, name)
	} else {
		err := createTCPUser(conn, name, passwordHash)
		if err != nil {
			conn.Write([]byte("error in creating users" + err.Error()))
			conn.Close()
			return
		}

		sendMessageTCP(conn, Settings.JoinedMessage, name)
	}

	readMessages(conn, name)
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
