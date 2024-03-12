package lib

import (
	"chatroom/app"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func GetUsers(ctx context.Context) ([]app.User, error) {
	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users ORDER BY name ASC"
	var users []app.User

	rows, err := app.DBConn.Query(ctx, query)
	if err != nil {
		log.Println("Error GetUsers", err)
		return users, err
	}

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.User])
	defer rows.Close()

	return users, err
}

func FindUserByEmail(email string) (app.User, error) {
	var user app.User
	ctx := context.Background()

	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users WHERE email = $1"
	err := app.DBConn.QueryRow(ctx, query, email).Scan(
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

func FindUserByID(id int) (app.User, error) {
	var user app.User
	ctx := context.Background()

	query := "SELECT id, name, is_online, theme, preferred_username, given_name, family_name, email FROM users WHERE id = $1"
	err := app.DBConn.QueryRow(ctx, query, id).Scan(
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

func insertUser(user app.User) (app.User, error) {
	ctx := context.Background()
	var id int
	query := `INSERT INTO users (name, is_online, theme, preferred_username, given_name, family_name, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`

	log.Println("inser user =>", user)
	err := app.DBConn.QueryRow(ctx, query, user.Name, user.IsOnline, user.Theme, user.PreferredUsername, user.GivenName, user.FamilyName, user.Email).Scan(&id)
	if err != nil {
		log.Fatal("Error executing INSERT statement:", err)
		return user, err

	}

	user.ID = id
	return user, nil
}

func UpdateUser(user app.User) {
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
	app.DBConn.QueryRow(ctx, query, user.Name, user.IsOnline, user.Theme, user.PreferredUsername, user.GivenName, user.FamilyName, user.Email, user.ID)
	log.Println("Field updated successfully", user.ID)
}
