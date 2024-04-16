package keycloak

import (
	"chatroom/app"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	usersBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading groups response body:", err)
		return users, err
	}

	fmt.Println(string(usersBody))

	err = json.Unmarshal([]byte(usersBody), &users)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return users, err
	}
	return users, err
}

func FindUserByID(ctx context.Context, id string) (app.KeyCloakUser, error) {
	var user app.KeyCloakUser

	query := `
	SELECT
		usr.id,
		usr.email,
		usr.first_name,
		usr.last_name,
		usr.username,
		usr.created_timestamp
	FROM
		public.user_entity usr
	JOIN
		public.user_attribute attr ON usr.id = attr.user_id
	WHERE
		usr.id = $1
	GROUP BY
		usr.id, usr.email, usr.first_name, usr.last_name, usr.username, usr.created_timestamp;

	`

	err := app.KeycloackDBConn.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.CreatedTimestamp,
	)

	if err != nil {
		return user, fmt.Errorf("error scanning row: %w", err)
	}
	return user, nil
}
