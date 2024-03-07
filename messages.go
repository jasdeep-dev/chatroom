package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func sendMessage(message string, sessionID string) {
	if message == "" {
		return
	}

	if sessionID == "" {
		return
	}

	session := UserSessions[sessionID]
	messageChannel <- Message{
		Text:      message,
		Name:      session.Name,
		Email:     session.KeyCloakUser.Email,
		TimeStamp: time.Now(),
	}
}
func getMessages(ctx context.Context) {
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

	for _, message := range MessagesArray {
		fmt.Printf("Timestamp: %s, Text: %s, Name: %s, Email: %s\n", message.TimeStamp.Format(time.RFC3339), message.Text, message.Name, message.Email)
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
