package lib

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSession struct {
	ID               int
	Name             string
	AccessToken      string
	LoggedInAt       time.Time
	KeyCloakUser     KeyCloakUserInfo
	SocketConnection *websocket.Conn
}

var UserSessions = make(map[string]UserSession)

type User struct {
	ID                int    `sql:"id"`
	Name              string `sql:"name"`
	IsOnline          bool   `sql:"is_online"`
	Theme             string `sql:"theme"`
	PreferredUsername string `sql:"preferred_username"`
	GivenName         string `sql:"given_name"`
	FamilyName        string `sql:"family_name"`
	Email             string `sql:"email"`
}

type user map[int]User

var Users = make(user)

type messages []Message

var MessagesArray messages

var Messages messages

type Message struct {
	TimeStamp time.Time
	Text      string
	Name      string
	Email     string
}

var genericMessage map[string]string
var MessageChannel = make(chan Message, 100)

var DBConn *pgxpool.Pool

var UserChannel = make(chan User, 100)
