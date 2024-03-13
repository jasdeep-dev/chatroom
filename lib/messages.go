package lib

import (
	"chatroom/app"
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

func MessageReceiver() {
	for msg := range app.MessageChannel {
		err := InsertMessage(msg.Context, msg.Message)
		if err != nil {
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

func sendMessage(ctx context.Context, message string, sessionID string, sockConn *websocket.Conn) {
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
		SockConn:  sockConn,
		Context:   ctx,
		Message: app.Message{
			Text:      message,
			UserID:    session.UserID,
			Name:      session.UserInfo.Name,
			Email:     session.UserInfo.Email,
			TimeStamp: time.Now(),
		},
	}
}

func GetMessages(ctx context.Context) ([]app.Message, error) {
	var err error
	var messages []app.Message

	query := `
		SELECT m.id, m.timestamp, m.text, u.id AS user_id, u.name AS name, u.email AS email
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
	`
	rows, err := app.DBConn.Query(ctx, query)
	if err != nil {
		log.Println("Error GetMessages", err)
		return messages, err
	}

	messages, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.Message])
	defer rows.Close()

	return messages, err
}

func InsertMessage(ctx context.Context, message app.Message) error {
	query := "INSERT INTO messages (timestamp, text, user_id) VALUES ($1, $2, $3)"

	newUser, err := FindUserByEmail(ctx, message.Email)
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
