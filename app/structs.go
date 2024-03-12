package app

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

// type user map[int]User

// var Users = make(user)

// type messages []Message

// var MessagesArray messages

// var Messages messages

type Message struct {
	TimeStamp time.Time
	Text      string
	Name      string
	Email     string
}

type MessageReceived struct {
	SessionID string
	Message   Message
}

// var genericMessage map[string]string
var MessageChannel = make(chan MessageReceived, 100)

var DBConn *pgxpool.Pool

var UserChannel = make(chan User, 100)

// type TemplateData struct {
// 	Users       map[int]User
// 	Messages    []Message
// 	CurrentUser KeyCloakUserInfo
// 	LoggedIn    time.Time
// }

type IDTokenClaims struct {
	Exp               int64  `json:"exp"`
	Iat               int64  `json:"iat"`
	AuthTime          int64  `json:"auth_time"`
	Jti               string `json:"jti"`
	Iss               string `json:"iss"`
	Aud               string `json:"aud"`
	Sub               string `json:"sub"`
	Typ               string `json:"typ"`
	Azp               string `json:"azp"`
	Nonce             string `json:"nonce"`
	SessionState      string `json:"session_state"`
	AtHash            string `json:"at_hash"`
	Acr               string `json:"acr"`
	Sid               string `json:"sid"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

type KeyCloakUserInfo struct {
	Sub               string `json:"sub"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

// var KeyCloakUser KeyCloakUserInfo
