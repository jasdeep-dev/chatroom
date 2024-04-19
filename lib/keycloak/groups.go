package keycloak

import (
	"bytes"
	"chatroom/app"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

func GetGroups(ctx context.Context) ([]app.KeycloakGroup, error) {
	var err error
	var users []app.KeycloakGroup

	query := `SELECT * FROM "public"."keycloak_group" ORDER BY "id" LIMIT 100 OFFSET 0;`
	// GetGroupsViaAPI()
	rows, err := app.KeycloackDBConn.Query(ctx, query)
	if err != nil {
		log.Println("Error GetUsers from Keycloak", err)
		return users, err
	}

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.KeycloakGroup])
	defer rows.Close()

	return users, err
}

func GetGroupsById(ctx context.Context, groupID string) (app.KeycloakGroup, error) {
	var err error
	var group app.KeycloakGroup

	query := `SELECT * FROM public.keycloak_group WHERE id = $1 ORDER BY "id" LIMIT 100 OFFSET 0;`
	// GetGroupsViaAPI()
	err = app.KeycloackDBConn.QueryRow(ctx, query, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.ParentGroup,
		&group.RealmID,
	)
	if err != nil {
		return group, fmt.Errorf("error scanning row for Groups query: %w", err)
	}
	return group, nil
}

func GetGroupsByUserID(ctx context.Context, groupID string) (app.KeycloakGroup, error) {
	var err error
	var group app.KeycloakGroup

	query := `SELECT * FROM public.keycloak_group WHERE id = $1 ORDER BY "id" LIMIT 100 OFFSET 0;`
	// GetGroupsViaAPI()
	err = app.KeycloackDBConn.QueryRow(ctx, query, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.ParentGroup,
		&group.RealmID,
	)
	if err != nil {
		return group, fmt.Errorf("error scanning row for Groups query: %w", err)
	}
	return group, nil
}

func GetGroupsByUserIDViaAPI(groupID string) (app.Group, error) {
	var group app.Group
	var err error
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/groups/%s",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
		groupID,
	)

	client := &http.Client{}

	groupsReq, err := http.NewRequest("GET", groupsURL, nil)
	if err != nil {
		log.Println("Error creating groups request:", err)
	}

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	groupsReq.Header.Set("Authorization", "Bearer "+access_token)

	groupsResp, err := client.Do(groupsReq)
	if err != nil {
		log.Println("Error making groups request:", err)
	}
	defer groupsResp.Body.Close()

	groupsBody, err := io.ReadAll(groupsResp.Body)
	if err != nil {
		log.Println("Error reading groups response body:", err)
	}

	err = json.Unmarshal([]byte(groupsBody), &group)
	if err != nil {
		log.Println("Error parsing JSON:", err)
	}

	return group, err
}

func GetGroupsViaAPI() ([]app.Group, error) {
	var groups []app.Group
	var err error
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/groups",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
	)

	client := &http.Client{}

	groupsReq, err := http.NewRequest("GET", groupsURL, nil)
	if err != nil {
		log.Println("Error creating groups request:", err)
	}

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	groupsReq.Header.Set("Authorization", "Bearer "+access_token)

	groupsResp, err := client.Do(groupsReq)
	if err != nil {
		log.Println("Error making groups request:", err)
	}
	defer groupsResp.Body.Close()

	groupsBody, err := ioutil.ReadAll(groupsResp.Body)
	if err != nil {
		log.Println("Error reading groups response body:", err)
	}

	err = json.Unmarshal([]byte(groupsBody), &groups)
	if err != nil {
		log.Println("Error parsing JSON:", err)
	}

	return groups, err
}

func CreateGroup(name string, userID string) error {
	var err error
	// Define the URL for creating a group

	url := fmt.Sprintf("%s/admin/realms/%s/groups",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
	)

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	group := struct {
		Name       string                 `json:"name"`
		Path       string                 `json:"path"`
		Attributes map[string]interface{} `json:"attributes"`
	}{
		Name: app.Titleize(name),
		Path: fmt.Sprintf("/%s", name),
		Attributes: map[string]interface{}{
			"created_by": []string{userID},
		},
	}

	groupJSON, err := json.Marshal(group)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(groupJSON))
	if err != nil {
		log.Printf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 409 {
		log.Printf("%s group already exists!", name)
	} else if resp.StatusCode != http.StatusCreated {
		log.Printf("unexpected response status: %s", resp.Status)
	}

	fmt.Println("Group created successfully")
	return nil
}

func GetGroupMembersViaAPI(groupID string) ([]app.KeyCloakUser, error) {
	var users []app.KeyCloakUser
	var err error

	groupsURL := fmt.Sprintf("%s/admin/realms/%s/groups/%s/members",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
		groupID,
	)
	client := &http.Client{}

	groupsReq, err := http.NewRequest("GET", groupsURL, nil)
	if err != nil {
		log.Println("Error creating groups request:", err)
	}

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	groupsReq.Header.Set("Authorization", "Bearer "+access_token)

	groupsResp, err := client.Do(groupsReq)
	if err != nil {
		log.Println("Error making groups request:", err)
	}
	defer groupsResp.Body.Close()

	groupsBody, err := ioutil.ReadAll(groupsResp.Body)
	if err != nil {
		log.Println("Error reading groups response body:", err)
	}

	err = json.Unmarshal([]byte(groupsBody), &users)
	if err != nil {
		log.Println("Error parsing JSON:", err)
	}
	return users, err
}

func GetUsersGroupsViaAPI(userID string) ([]app.Group, error) {
	var groups []app.Group
	var err error

	groupsURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
		userID,
	)

	client := &http.Client{}

	groupsReq, err := http.NewRequest("GET", groupsURL, nil)
	if err != nil {
		log.Println("Error creating groups request:", err)
	}

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")
	groupsReq.Header.Set("Authorization", "Bearer "+access_token)

	groupsResp, err := client.Do(groupsReq)
	if err != nil {
		log.Println("Error making groups request:", err)
	}
	defer groupsResp.Body.Close()

	groupsBody, err := io.ReadAll(groupsResp.Body)
	if err != nil {
		log.Println("Error reading groups response body:", err)
	}

	err = json.Unmarshal([]byte(groupsBody), &groups)
	if err != nil {
		log.Println("Error parsing JSON:", err)
	}
	for _, group := range groups {
		app.GroupIds = append(app.GroupIds, group.ID)
	}
	return groups, err
}

func AddUserToGroup(userID string, groupID string) error {
	var err error
	// Define the URL for creating a group

	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups/%s",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
		userID,
		groupID,
	)

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 409 {
		log.Printf("User added to group already exists!")
	} else if resp.StatusCode != http.StatusCreated {
		log.Printf("unexpected response status: %s", resp.Status)
	}

	fmt.Printf("User has been added to the group!")
	return nil
}

func FindGroupByName(ctx context.Context, name string) (app.KeycloakGroup, error) {
	var err error
	var group app.KeycloakGroup

	query := `SELECT * FROM public.keycloak_group WHERE "name" = $1`

	rows, err := app.KeycloackDBConn.Query(ctx, query, name)
	if err != nil {
		log.Println("Error in FindGroupByName", name, err)
	}

	group, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[app.KeycloakGroup])
	defer rows.Close()
	if err != nil {
		return group, fmt.Errorf("error scanning row: %w - email %v", err, name)
	}
	return group, nil
}

func GroupsCreatedByUser(ctx context.Context, userID string) (groupIds []string, err error) {
	query := `
        SELECT kg.id
        FROM keycloak_group kg
        JOIN group_attribute ga ON kg.id = ga.group_id
        WHERE ga.name = 'created_by' AND ga.value = $1;
    `

	rows, err := app.KeycloackDBConn.Query(ctx, query, userID)
	if err != nil {
		log.Println("Error fetching groups from Keycloak:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			log.Println("Error scanning group row:", err)
			return nil, err
		}
		groupIds = append(groupIds, groupID)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over group rows:", err)
		return nil, err
	}

	return groupIds, nil
}

func BulkInsertUserGroupMembership(ctx context.Context, groupIds []string, userID string) error {
	query := "INSERT INTO user_group_membership (group_id, user_id) VALUES "

	values := ""

	for i, pair := range groupIds {

		values += fmt.Sprintf("('%s', '%s')", string(pair), userID)

		// If it's not the last pair, add a comma to separate the values
		if i != len(groupIds)-1 {
			values += ","
		}
	}

	// Complete the SQL statement
	query += values

	// Add the WHERE NOT EXISTS clause to ensure duplicates are not inserted
	query += `
		ON CONFLICT (group_id, user_id) DO NOTHING;
	`

	// Execute the SQL statement
	_, err := app.KeycloackDBConn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}

func GetMessagesByGroupID(groupID string) []app.Message {
	var err error
	var messages []app.Message

	query := `
		SELECT
			id,
			timestamp,
			text,
			user_id,
			group_id,
			first_name AS name,
			email
		FROM
			messages
		WHERE group_id=$1`

	rows, err := app.DBConn.Query(context.Background(), query, groupID)
	if err != nil {
		log.Println("Error GetUsers from Keycloak", err)
		return messages
	}

	messages, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.Message])
	if err != nil {
		log.Println("Error GetUsers from Keycloak", err)
		return messages
	}
	defer rows.Close()

	return messages
}
