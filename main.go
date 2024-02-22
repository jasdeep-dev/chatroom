package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
)

var users map[string]User

type User struct {
	Conn  net.Conn
	Name  string
	Color string
}

type Message struct {
	Text string
	Name string
}

func main() {
	fmt.Println("Starting server")
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("error 1")
	}
	defer ln.Close()

	users = make(map[string]User)
	ch := make(chan Message)
	go receiver(ch)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error 2")
		}
		go newUser(conn, ch)
	}
}

func receiver(ch chan Message) {
	for message := range ch {
		fmt.Println("Number of users: ", len(users))
		for _, user := range users {
			if user.Name != message.Name {
				user.Conn.Write([]byte(users[message.Name].Color + message.Name + "> \x1b[0m" + message.Text + "\n"))
			}
		}
	}
}

func newUser(conn net.Conn, ch chan Message) {
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

		ch <- Message{
			Text: "I have joined the chat",
			Name: name,
		}
		conn.Write([]byte("Hi " + name + ", welcome to chatroom\n"))
	} else {
		users[name] = User{
			Conn:  conn,
			Name:  name,
			Color: users[name].Color,
		}
		ch <- Message{
			Text: "I am Back!",
			Name: name,
		}
		conn.Write([]byte("Hi " + name + ", welcome back\n"))
	}

	for {
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Connection terminated by user", name)
			break
		}
		response = strings.Trim(response, "\n")

		ch <- Message{
			Text: response,
			Name: name,
		}
	}
}
