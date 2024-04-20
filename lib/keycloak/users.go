package keycloak

import (
	"bytes"
	"chatroom/app"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

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
	app.AllUsers = users
	return users, err
}

func FindUserByID(id string) (app.KeyCloakUser, error) {
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

func RemoveUserFromGroup(userID string, groupID string) error {
	var err error

	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups/%s",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
		userID,
		groupID,
	)

	access_token := os.Getenv("ADMIN_ACCESS_TOKEN")

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte{}))
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
		log.Printf("User not present in the group!")
	} else if resp.StatusCode != http.StatusCreated {
		log.Printf("unexpected response status: %s", resp.Status)
	}

	fmt.Printf("User has been removed from the group!")
	return nil
}
