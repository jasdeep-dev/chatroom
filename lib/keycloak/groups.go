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
)

type GroupService interface {
	GetGroups() ([]app.KeycloakGroup, error)
}

type KeycloakService struct {
	AccessToken string
	URL         string
	Realm       string
}

func NewKeycloakService(accessToken, url, realm string) *KeycloakService {
	return &KeycloakService{
		AccessToken: os.Getenv("ADMIN_ACCESS_TOKEN"),
		URL:         url,
		Realm:       realm,
	}
}

func (kc *KeycloakService) newRequest(method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+kc.AccessToken)

	return req, nil
}

func (kc *KeycloakService) doRequest(req *http.Request, v interface{}) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// Example usage
// kc := NewKeycloakService(
// 	os.Getenv("ADMIN_ACCESS_TOKEN"),
// 	os.Getenv("KEYCLOAK_URL"),
// 	os.Getenv("REALM_NAME"),
// )

// groups, err := kc.GetUsersGroupsViaAPI(userID)
//
//	if err != nil {
//		log.Printf("Error getting groups: %v", err)
//	}

func (kc *KeycloakService) GetUsersGroupsViaAPI(userID string) ([]app.Group, error) {
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups", kc.URL, kc.Realm, userID)
	req, err := kc.newRequest("GET", groupsURL, nil)
	if err != nil {
		return nil, err
	}

	var groups []app.Group
	if err := kc.doRequest(req, &groups); err != nil {
		return nil, err
	}

	return groups, nil
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

	SetAdminToken()
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

	err = json.Unmarshal([]byte(groupsBody), &users)
	if err != nil {
		log.Println("Error parsing JSON:", err)
	}
	return users, err
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
