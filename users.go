package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type UserSession struct {
	ID               int
	Name             string
	AccessToken      string
	LoggedInAt       time.Time
	KeyCloakUser     KeyCloakUserInfo
	SocketConnection *websocket.Conn
}

var UserSessions = make(map[string]UserSession)

// func createHTTPUser(w http.ResponseWriter, r *http.Request) {
// 	name := r.Form.Get("name")
// 	password := r.Form.Get("password")

// 	passwordHash, err := bcrypt.GenerateFromPassword(
// 		[]byte(password),
// 		bcrypt.DefaultCost,
// 	)
// 	if err != nil {
// 		r.Header.Set("ERROR", "Unable to create user")
// 		loginHandler(w, r)
// 		return
// 	}

// 	user, ok := users[name]

// 	if ok {
// 		// Authenticate user by comparing password with the hashed password
// 		err := bcrypt.CompareHashAndPassword([]byte(users[name].PasswordHash), []byte(password))
// 		if err != nil {
// 			r.Header.Set("ERROR", "Invalid password")
// 			loginHandler(w, r)
// 			return
// 		}

// 		user.IsOnline = true
// 		users[name] = user
// 	} else {

// 		users[name] = User{
// 			Name:         name,
// 			IsOnline:     true,
// 			PasswordHash: string(passwordHash),
// 		}

// 		// insertUser(users[name])
// 	}

// 	sessionId, err := generateSessionID(name)
// 	if err != nil {
// 		fmt.Println("Error creating Session ID", err)
// 	}

// 	// UserSessions[sessionId] = UserSession{
// 	// 	Name:       name,
// 	// 	LoggedInAt: time.Now(),
// 	// }

// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "session_id",         // Name of the cookie
// 		Value:    sessionId,            // Session ID generated previously
// 		Path:     "/",                  // Path attribute (scope of the cookie)
// 		SameSite: http.SameSiteLaxMode, // SameSite attribute
// 	})

// 	messageChannel <- Message{
// 		TimeStamp: time.Now(),
// 		Text:      genericMessage["joined"],
// 		Name:      name,
// 	}
// }

// func generateSessionID(name string) (string, error) {
// 	b := make([]byte, 128) // 16 bytes for 128 bits of entropy
// 	_, err := crypt.Read(b)
// 	if err != nil {
// 		return "", err
// 	}

// 	return hex.EncodeToString(b), nil
// }

func insertUser(user User) *User {
	ctx := context.Background()
	var id int
	query := `INSERT INTO users (name, is_online, theme, preferred_username, given_name, family_name, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`

	err := DBConn.QueryRow(ctx, query, user.Name, user.IsOnline, user.Theme, user.PreferredUsername, user.GivenName, user.FamilyName, user.Email).Scan(&id)
	if err != nil {
		log.Fatal("Error executing INSERT statement:", err)
	}

	user.ID = id
	Users[id] = user
	return &user
}

func getUsers(ctx context.Context) {

	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users LIMIT 10"

	rows, err := DBConn.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.IsOnline,
			&user.Theme,
			&user.PreferredUsername,
			&user.GivenName,
			&user.FamilyName,
			&user.Email); err != nil {
			log.Fatal(err)
		}
		Users[user.ID] = user
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func FindUserByEmail(email string) (*User, error) {
	var user User

	// Prepare SQL query
	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users WHERE email = $1"

	// Execute query and scan result into user struct
	err := DBConn.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Name,
		&user.IsOnline,
		&user.Theme,
		&user.PreferredUsername,
		&user.GivenName,
		&user.FamilyName,
		&user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, err
	}

	return &user, nil
}
