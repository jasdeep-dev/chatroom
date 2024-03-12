package app

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserInfo struct {
	Sub               string `json:"sub"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

type UserSession struct {
	ID          string    `json:"id"`
	UserID      int       `json:"user_id"`
	AccessToken string    `json:"access_token"`
	LoggedInAt  time.Time `json:"logged_in_at"`
	UserInfo    UserInfo  `json:"user_info"`
}

var SocketConnections []*websocket.Conn

type User struct {
	ID                int    `db:"id"`
	Name              string `db:"name"`
	IsOnline          bool   `db:"is_online"`
	Theme             string `db:"theme"`
	PreferredUsername string `db:"preferred_username"`
	GivenName         string `db:"given_name"`
	FamilyName        string `db:"family_name"`
	Email             string `db:"email"`
}

type Message struct {
	TimeStamp time.Time
	Text      string
	Name      string
	Email     string
}

type MessageReceived struct {
	SessionID string
	SockConn  *websocket.Conn
	Message   Message
}

var MessageChannel = make(chan MessageReceived, 100)

var DBConn *pgxpool.Pool

var UserChannel = make(chan User, 100)

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
