package keycloak

import (
	"bytes"
	"chatroom/app"
	"context"
	"encoding/json"
	"fmt"
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
	SetAdminToken()
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

func (kc *KeycloakService) doRequestWithoutResponse(req *http.Request) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		log.Printf("group already exists!")
	} else if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	fmt.Println("Group created successfully")
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

func (kc *KeycloakService) GetUsersGroupsViaAPI(userID string) (groups []app.Group, err error) {
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups", kc.URL, kc.Realm, userID)
	req, err := kc.newRequest("GET", groupsURL, nil)
	if err != nil {
		return nil, err
	}

	if err := kc.doRequest(req, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (kc *KeycloakService) GetGroupByIDViaAPI(groupID string) (group app.Group, err error) {
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/groups/%s", kc.URL, kc.Realm, groupID)
	req, err := kc.newRequest("GET", groupsURL, nil)
	if err != nil {
		return group, err
	}

	if err := kc.doRequest(req, &group); err != nil {
		return group, err
	}

	return group, nil
}

func (kc *KeycloakService) GetGroupsViaAPI() (groups []app.Group, err error) {
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/groups", kc.URL, kc.Realm)
	req, err := kc.newRequest("GET", groupsURL, nil)
	if err != nil {
		return nil, err
	}

	if err := kc.doRequest(req, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (kc *KeycloakService) CreateGroup(name string, userID string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/groups", kc.URL, kc.Realm)

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
		return err
	}

	req, err := kc.newRequest("POST", url, groupJSON)
	if err != nil {
		return err
	}

	return kc.doRequestWithoutResponse(req)
}
func (kc *KeycloakService) GetGroupMembersViaAPI(groupID string) (users []app.KeyCloakUser, err error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s/members",
		kc.URL,
		kc.Realm,
		groupID,
	)

	req, err := kc.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if err := kc.doRequest(req, &users); err != nil {
		return nil, err
	}
	return users, nil
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
