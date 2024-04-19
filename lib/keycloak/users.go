package keycloak

import (
	"chatroom/app"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// func GetUsers(ctx context.Context) ([]app.KeyCloakUser, error) {
// 	var err error
// 	var users []app.KeyCloakUser

// 	query := `
// 		SELECT
// 			usr.*,
// 			json_agg(json_build_object('name', attr.name, 'value', attr.value)) AS attributes
// 		FROM
// 			public.user_entity usr
// 		LEFT JOIN public.user_attribute attr ON usr.id = attr.user_id
// 		GROUP BY
// 			usr.id;
// 	`

// 	rows, err := app.KeycloackDBConn.Query(ctx, query)
// 	if err != nil {
// 		log.Println("Error GetUsers from Keycloak", err)
// 		return users, err
// 	}

// 	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.KeyCloakUser])
// 	defer rows.Close()

// 	return users, err
// }

func GetUsersViaAPI() ([]app.KeyCloakUser, error) {
	var users []app.KeyCloakUser
	var err error
	url := fmt.Sprintf("%s/admin/realms/%s/users",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
	)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating groups request:", err)
		return users, err
	}

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	req.Header.Set("Authorization", "Bearer "+access_token)

	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making groups request:", err)
		return users, err
	}
	defer response.Body.Close()

	usersBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading groups response body:", err)
		return users, err
	}

	err = json.Unmarshal([]byte(usersBody), &users)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return users, err
	}
	app.Users = users
	return users, err
}

func FindUserByID(ctx context.Context, id string) (app.KeyCloakUser, error) {
	var user app.KeyCloakUser
	var err error
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
		id,
	)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating groups request:", err)
		return user, err
	}

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")
	if access_token == "" {
		log.Println("Access token is not present")
		SetAdminToken()
	}

	req.Header.Set("Authorization", "Bearer "+access_token)

	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making groups request:", err)
		return user, err
	}
	defer response.Body.Close()

	usersBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading groups response body:", err)
		return user, err
	}

	err = json.Unmarshal([]byte(usersBody), &user)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return user, err
	}
	return user, err
}
