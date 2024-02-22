package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

var users []net.Conn

type User struct {
	Addr net.Addr
	Name string
	Msg  string
}

func main() {
	fmt.Println("Starting server")
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("error 1")
	}
	defer ln.Close()

	ch := make(chan User)
	go receiver(ch)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error 2")
		}

		go newUser(conn, ch)
		users = append(users, conn)
	}
}

func receiver(ch chan User) {
	for user := range ch {
		for _, conn := range users {
			if user.Addr != conn.RemoteAddr() {
				conn.Write([]byte(user.Msg))
			}
		}
	}
}

func newUser(conn net.Conn, ch chan User) {
	conn.Write([]byte("What is your name?\n"))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("error 3")
	}

	name = strings.Trim(name, "\n")
	conn.Write([]byte("Hi " + name + ", welcome to chatroom\n"))

	for {
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Connection terminated by user", name)
			break
		}
		response = strings.Trim(response, "\n")

		ch <- User{
			Msg:  name + "> " + response + "\n",
			Addr: conn.RemoteAddr(),
			Name: name,
		}
	}
}
