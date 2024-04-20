package keycloak

import (
	"chatroom/app"
	"fmt"
)

func (kc *KeycloakService) GetUsersViaAPI() (users []app.KeyCloakUser, err error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users",
		kc.URL,
		kc.Realm,
	)

	req, err := kc.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if err := kc.doRequest(req, &users); err != nil {
		return nil, err
	}

	app.Users = users
	app.AllUsers = users

	return users, nil
}

func (kc *KeycloakService) FindUserByID(id string) (user app.KeyCloakUser, err error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s",
		kc.URL,
		kc.Realm,
		id,
	)

	req, err := kc.newRequest("GET", url, nil)
	if err != nil {
		return user, err
	}

	if err := kc.doRequest(req, &user); err != nil {
		return user, err
	}

	return user, err
}

func (kc *KeycloakService) RemoveUserFromGroup(userID string, groupID string) error {
	var err error

	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups/%s",
		kc.URL,
		kc.Realm,
		userID,
		groupID,
	)

	req, err := kc.newRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return kc.doRequestWithoutResponse(req)
}
