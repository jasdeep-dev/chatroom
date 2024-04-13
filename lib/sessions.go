package lib

import (
	"chatroom/app"
	"chatroom/lib/keycloak"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

var Cache *memcache.Client

func InitCache() error {
	Cache = memcache.New("localhost:11211")
	return Cache.Ping()
}

func GetSession(sessionID string, r *http.Request) (app.UserSession, error) {
	var userSession app.UserSession

	val, err := Cache.Get(sessionID)
	if err != nil {
		return userSession, err
	}

	err = json.Unmarshal(val.Value, &userSession)
	if err != nil {
		return userSession, err
	}

	keycloak, err := r.Cookie("KEYCLOAK_SESSION")
	if err != nil {
		log.Print("Error")
	}
	fmt.Println("===> keyclock", keycloak)
	return userSession, err
}

func GetUserFromSession(r *http.Request) (app.KeyCloakUser, error) {
	var user app.KeyCloakUser

	keyCloakSessionID, err := r.Cookie("user_id")
	if err != nil {
		log.Println("keycloak user does not exist in the session")
	}

	userID := extractUserID(keyCloakSessionID.Value)

	user, err = keycloak.FindUserByID(r.Context(), userID)
	if err != nil {
		return user, err
	}
	return user, err
}

func extractUserID(cookieID string) string {
	parts := strings.Split(cookieID, "/")
	if len(parts) >= 3 {
		return parts[1] // Assuming user ID is always the second part
	}
	return "" // Return empty string if the cookie ID format is invalid
}

func SetSession(session app.UserSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	err = Cache.Set(&memcache.Item{Key: session.ID, Value: data})
	if err != nil {
		return err
	}

	return nil
}
