package lib

import (
	"context"
	"log"
	"time"
)

func MessageReceiver() {
	for message := range MessageChannel {
		InsertMessage(message)
		MessagesArray = append(MessagesArray, message)

		Messages = append(Messages, message)

		deliverMessageToWebSocketConnections(message)
	}
}

func UserReciver() {
	for user := range UserChannel {
		if user.Name != "" {
			deliverUsersToWebSocketConnections(user)
		}
	}
}

func deliverMessageToWebSocketConnections(message Message) {
	for _, userSession := range UserSessions {
		if userSession.SocketConnection == nil {
			continue
		}
		msg := struct {
			Message Message
		}{
			Message: message,
		}
		userSession.SocketConnection.WriteJSON(msg)
	}
}

func deliverUsersToWebSocketConnections(user User) {
	for _, userSession := range UserSessions {
		if userSession.SocketConnection == nil {
			continue
		}

		usr := struct {
			User User
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

	session := UserSessions[sessionID]
	MessageChannel <- Message{
		Text:      message,
		Name:      session.Name,
		Email:     session.KeyCloakUser.Email,
		TimeStamp: time.Now(),
	}
}
func GetMessages(ctx context.Context) {
	query := `
		SELECT m.timestamp, m.text, u.name, u.email
		FROM messages m
		INNER JOIN users u ON m.user_id = u.id
	`

	rows, err := DBConn.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.TimeStamp, &message.Text, &message.Name, &message.Email); err != nil {
			log.Fatal(err)
		}
		MessagesArray = append(MessagesArray, message)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func InsertMessage(message Message) error {
	query := "INSERT INTO messages (timestamp, text, user_id) VALUES ($1, $2, $3)"

	newUser, err := FindUserByEmail(message.Email)
	if err != nil {
		log.Fatal("user with email does not exist in our database")
	}
	// Execute the SQL statement
	_, err = DBConn.Exec(context.Background(), query, message.TimeStamp, message.Text, newUser.ID)
	if err != nil {
		return err
	}

	return nil
}
