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
	userSession, ok := app.UserSessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found in deliverErrorToSocket")
	}

	msg := struct {
		Error string
	}{
		Error: err.Error(),
	}
	return userSession.SocketConnection.WriteJSON(msg)
}

func UserReciver() {
	for user := range app.UserChannel {
		if user.Name != "" {
			deliverUsersToWebSocketConnections(user)
		}
	}
}

func deliverMessageToWebSocketConnections(message app.Message) {
	for _, userSession := range app.UserSessions {
		if userSession.SocketConnection == nil {
			continue
		}
		msg := struct {
			Message app.Message
		}{
			Message: message,
		}
		userSession.SocketConnection.WriteJSON(msg)
	}
}

func deliverUsersToWebSocketConnections(user app.User) {
	for _, userSession := range app.UserSessions {
		if userSession.SocketConnection == nil {
			continue
		}

		usr := struct {
			User app.User
		}{
			User: user,
		}
		userSession.SocketConnection.WriteJSON(usr)
	}
}

func sendMessage(message string, sessionID string) {
	if message == "" {
		return
	}

	if sessionID == "" {
		return
	}

	session := app.UserSessions[sessionID]
	app.MessageChannel <- app.MessageReceived{
		SessionID: sessionID,
		Message: app.Message{
			Text:      message,
			Name:      session.Name,
			Email:     session.KeyCloakUser.Email,
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
