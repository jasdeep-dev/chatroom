package lib

import (
	"context"
	"fmt"
	"log"
)

func GetUsers(ctx context.Context) {
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

	log.Println("users", Users)
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func FindUserByEmail(email string) (User, error) {
	var user User
	ctx := context.Background()

	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users WHERE email = $1"
	err := DBConn.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.IsOnline,
		&user.Theme,
		&user.PreferredUsername,
		&user.GivenName,
		&user.FamilyName,
		&user.Email,
	)
	if err != nil {
		return user, fmt.Errorf("error scanning row: %w", err)
	}
	return user, nil
}

func FindUserByID(id int) (User, error) {
	var user User
	ctx := context.Background()

	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users WHERE id = $1"
	err := DBConn.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.IsOnline,
		&user.Theme,
		&user.PreferredUsername,
		&user.GivenName,
		&user.FamilyName,
		&user.Email,
	)
	if err != nil {
		return user, fmt.Errorf("error scanning row: %w", err)
	}
	return user, nil
}

func insertUser(user User) User {
	ctx := context.Background()
	var id int
	query := `INSERT INTO users (name, is_online, theme, preferred_username, given_name, family_name, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`

	log.Println("inser user =>", user)
	err := DBConn.QueryRow(ctx, query, user.Name, user.IsOnline, user.Theme, user.PreferredUsername, user.GivenName, user.FamilyName, user.Email).Scan(&id)
	if err != nil {
		log.Fatal("Error executing INSERT statement:", err)
	}

	user.ID = id
	Users[id] = user
	log.Println("inserted User=>", user)
	return user
}

func UpdateUser(user User) {
	ctx := context.Background()
	// Update statement
	query := `UPDATE users 
              SET name = $1, 
                  is_online = $2, 
                  theme = $3, 
                  preferred_username = $4, 
                  given_name = $5, 
                  family_name = $6, 
                  email = $7 
              WHERE id = $8;`

	// Execute the update statement
	_ = DBConn.QueryRow(ctx, query, user.Name, user.IsOnline, user.Theme, user.PreferredUsername, user.GivenName, user.FamilyName, user.Email, user.ID)
	log.Println("Field updated successfully", user.ID)
}
