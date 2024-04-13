package lib

import (
	"chatroom/app"
	"context"
	"errors"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

func MessageReceiver() {
	for msg := range app.MessageChannel {
		err := InsertMessage(msg.Context, msg.Message)
		if err != nil && msg.SockConn != nil {
			err = deliverErrorToSocket(msg.SockConn, err)
			if err != nil {
				log.Println("Error from deliverErrorToSocket:", err)
			}
			continue
		}

		deliverMessageToWebSocketConnections(msg.Message)
	}
}

func deliverErrorToSocket(conn *websocket.Conn, err error) error {
	msg := struct {
		Error string
	}{
		Error: err.Error(),
	}

	if conn != nil {
		return conn.WriteJSON(msg)
	} else {
		return errors.New("websocket connection is nil")
	}
}

func UserReciver() {
	for user := range app.KUserChannel {
		if user.ID != "" {
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

func deliverUsersToWebSocketConnections(user app.KeyCloakUser) {
	for _, conn := range app.SocketConnections {
		if conn == nil {
			continue
		}

		usr := struct {
			User app.KeyCloakUser
		}{
			User: user,
		}
		conn.WriteJSON(usr)
	}
}

func sendMessage(ctx context.Context, message string, session app.UserSession, sockConn *websocket.Conn) {
	if message == "" {
		log.Println("sendMessage: Message is blank")
		return
	}

	app.MessageChannel <- app.MessageReceived{
		SessionID: session.ID,
		SockConn:  sockConn,
		Context:   ctx,
		Message: app.Message{
			Text:      message,
			UserID:    session.UserID,
			Name:      *session.KeyCloakUser.FirstName,
			Email:     *session.KeyCloakUser.Email,
			TimeStamp: time.Now(),
		},
	}
}

func GetMessages(ctx context.Context) ([]app.Message, error) {
	var err error
	var messages []app.Message

	query := `SELECT id, timestamp, text, id AS user_id, first_name AS name, email AS email FROM messages`
	rows, err := app.DBConn.Query(ctx, query)
	if err != nil {
		log.Println("Error GetMessages: ", err)
		return messages, err
	}

	messages, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.Message])
	defer rows.Close()

	return messages, err
}

func InsertMessage(ctx context.Context, message app.Message) error {
	query := "INSERT INTO messages (timestamp, text, user_id, first_name, email) VALUES ($1, $2, $3, $4, $5)"

	_, err := app.DBConn.Exec(ctx, query, message.TimeStamp, message.Text, message.UserID, message.Name, message.Email)
	if err != nil {
		return err
	}

	return nil
}
