package keycloak

import (
	"bytes"
	"chatroom/app"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
		return fmt.Errorf("%s group already exists", resp.Status)
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}

func (kc *KeycloakService) GetUsersGroupsViaAPI(userID string) (err error) {
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups", kc.URL, kc.Realm, userID)
	req, err := kc.newRequest("GET", groupsURL, nil)
	if err != nil {
		return err
	}

	if err := kc.doRequest(req, &app.Groups); err != nil {
		return err
	}

	app.PublicGroup = []app.Group{}
	app.PersonalGroup = []app.Group{}
	for _, group := range app.Groups {
		if group.ParentGroup == app.PublicGroupID {
			app.PublicGroup = append(app.PublicGroup, group)
		} else if group.ParentGroup == app.PersonalGroupID {
			app.PersonalGroup = append(app.PersonalGroup, group)
		}
	}
	return nil
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

func (kc *KeycloakService) GetGroupsViaAPI() (err error) {
	groupsURL := fmt.Sprintf("%s/admin/realms/%s/groups", kc.URL, kc.Realm)
	req, err := kc.newRequest("GET", groupsURL, nil)
	if err != nil {
		return err
	}

	if err := kc.doRequest(req, &app.Groups); err != nil {
		return err
	}

	return nil
}

func (kc *KeycloakService) CreateGroup(name string, parentGroup string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s/children", kc.URL, kc.Realm, parentGroup)

	group := struct {
		Name       string                 `json:"name"`
		Path       string                 `json:"path"`
		Attributes map[string]interface{} `json:"attributes"`
	}{
		Name: name,
		Attributes: map[string]interface{}{
			"created_by": []string{app.Session.KeyCloakUser.ID},
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

func (kc *KeycloakService) GetGroupMembersViaAPI(groupID string) (err error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s/members",
		kc.URL,
		kc.Realm,
		groupID,
	)

	req, err := kc.newRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if err := kc.doRequest(req, &app.GroupUsers); err != nil {
		return err
	}
	return nil
}

func (kc *KeycloakService) DeleteGroupViaAPI(groupID string) (err error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s/members",
		kc.URL,
		kc.Realm,
		groupID,
	)

	req, err := kc.newRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return kc.doRequestWithoutResponse(req)
}

func (kc *KeycloakService) AddUserToGroup(userID string, groupID string) error {
	var err error

	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups/%s",
		kc.URL, kc.Realm,
		userID,
		groupID,
	)

	req, err := kc.newRequest("PUT", url, nil)
	if err != nil {
		log.Printf("error creating request: %v", err)
	}

	return kc.doRequestWithoutResponse(req)
}

func GroupsCreatedByUser(ctx context.Context, userID string) (groupIds []string, err error) {
	query := `
        SELECT kg.id
        FROM keycloak_group kg
        JOIN group_attribute ga ON kg.id = ga.group_id
        WHERE ga.name = 'created_by' AND ga.value = $1 AND kg.parent_group = $2;
    `

	rows, err := app.KeycloackDBConn.Query(ctx, query, userID, app.PublicGroup)
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
