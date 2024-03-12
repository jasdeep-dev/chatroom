package lib

import (
	"chatroom/app"
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
)

var Cache *memcache.Client

func InitCache() error {
	Cache = memcache.New("localhost:11211")
	return Cache.Ping()
}

func GetSession(sessionID string) (app.UserSession, error) {
	var userSession app.UserSession
	val, err := Cache.Get(sessionID)
	if err != nil {
		return userSession, err
	}

	err = json.Unmarshal(val.Value, &userSession)
	if err != nil {
		return userSession, err
	}

	return userSession, err
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
