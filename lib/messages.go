package lib

import (
	"chatroom/app"
	"context"
	"fmt"
	"log"
	"time"
)

func MessageReceiver() {
	for messageReceieved := range app.MessageChannel {
		err := InsertMessage(messageReceieved.Message)
		if err != nil {
			err = deliverErrorToSocket(messageReceieved.SessionID, err)
			if err != nil {
				log.Println("Error from deliverErrorToSocket:", err)
			}
			continue
		}

		deliverMessageToWebSocketConnections(messageReceieved.Message)
	}
}

func deliverErrorToSocket(sessionID string, err error) error {
	conn, ok := app.SocketConnections[sessionID]
	if !ok {
		return fmt.Errorf("session not found in deliverErrorToSocket")
	}

	msg := struct {
		Error string
	}{
		Error: err.Error(),
	}
	return conn.WriteJSON(msg)
}

func UserReciver() {
	for user := range app.UserChannel {
		if user.Name != "" {
			deliverUsersToWebSocketConnections(user)
		}
	}
}

func deliverMessageToWebSocketConnections(message app.Message) {
	for _, conn := range app.SocketConnections {
		if conn == nil {
			continue
		}
		msg := struct {
			Message app.Message
		}{
			Message: message,
		}
		conn.WriteJSON(msg)
	}
}

func deliverUsersToWebSocketConnections(user app.User) {
	for _, conn := range app.SocketConnections {
		if conn == nil {
			continue
		}

		usr := struct {
			User app.User
		}{
			User: user,
		}
		conn.WriteJSON(usr)
	}
}

func sendMessage(message string, sessionID string) {
	if message == "" {
		log.Println("sendMessage: Message is blank")
		return
	}

	if sessionID == "" {
		log.Println("sendMessage: Session id is blank")
		return
	}

	session, err := GetSession(sessionID)
	if err != nil {
		log.Println("sendMessage: Session not found", sessionID)
	}

	app.MessageChannel <- app.MessageReceived{
		SessionID: sessionID,
		Message: app.Message{
			Text:      message,
			Name:      session.UserInfo.Name,
			Email:     session.UserInfo.Email,
			TimeStamp: time.Now(),
		},
	}
}

func GetMessages(ctx context.Context) ([]app.Message, error) {
	query := `
		SELECT m.timestamp, m.text, u.name, u.email
		FROM messages m
		INNER JOIN users u ON m.user_id = u.id
	`

	var messages []app.Message
	rows, err := app.DBConn.Query(ctx, query)
	if err != nil {
		return messages, err
	}
	defer rows.Close()

	for rows.Next() {
		var message app.Message
		if err := rows.Scan(&message.TimeStamp, &message.Text, &message.Name, &message.Email); err != nil {
			log.Fatal(err)
		}
		messages = append(messages, message)
	}
	err = rows.Err()

	return messages, err
}

func InsertMessage(message app.Message) error {
	query := "INSERT INTO messages (timestamp, text, user_id) VALUES ($1, $2, $3)"

	newUser, err := FindUserByEmail(message.Email)
	if err != nil {
		log.Println("user with email does not exist in our database")
		return err
	}
	// Execute the SQL statement
	_, err = app.DBConn.Exec(context.Background(), query, message.TimeStamp, message.Text, newUser.ID)
	if err != nil {
		return err
	}

	return nil
}
