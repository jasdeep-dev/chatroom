package app

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSession struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	AccessToken  string       `json:"access_token"`
	LoggedInAt   time.Time    `json:"logged_in_at"`
	KeyCloakUser KeyCloakUser `json:"keycloak_user"`
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

type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var GroupIds = []string{}
var Users = []KeyCloakUser{}

// type KeyCloakUser struct {
// 	ID                       string      `json:"id"`
// 	Email                    *string     `json:"email"`
// 	EmailConstraint          *string     `json:"email_constraint"`
// 	EmailVerified            bool        `json:"email_verified"`
// 	Enabled                  bool        `json:"enabled"`
// 	FederationLink           *string     `json:"federation_link"`
// 	FirstName                *string     `json:"first_name"`
// 	LastName                 *string     `json:"last_name"`
// 	RealmID                  *string     `json:"realm_id"`
// 	Username                 *string     `json:"username"`
// 	CreatedTimestamp         int64       `json:"created_timestamp"`
// 	ServiceAccountClientLink *string     `json:"service_account_client_link"`
// 	NotBefore                int32       `json:"not_before"`
// 	Attributes               []Attribute `json:"attributes"`
// }

type KeyCloakUser struct {
	ID                         string              `json:"id"`
	Username                   string              `json:"username"`
	FirstName                  string              `json:"firstName"`
	LastName                   string              `json:"lastName"`
	Email                      string              `json:"email"`
	EmailVerified              bool                `json:"emailVerified"`
	Attributes                 map[string][]string `json:"attributes"`
	CreatedTimestamp           int64               `json:"createdTimestamp"`
	Enabled                    bool                `json:"enabled"`
	Totp                       bool                `json:"totp"`
	DisableableCredentialTypes []string            `json:"disableableCredentialTypes"`
	RequiredActions            []string            `json:"requiredActions"`
	NotBefore                  int64               `json:"notBefore"`
	Access                     Access              `json:"access"`
}

type Access struct {
	ManageGroupMembership bool `json:"manageGroupMembership"`
	View                  bool `json:"view"`
	MapRoles              bool `json:"mapRoles"`
	Impersonate           bool `json:"impersonate"`
	Manage                bool `json:"manage"`
}

type Message struct {
	ID        int       `db:"id"`
	TimeStamp time.Time `db:"timestamp"`
	Text      string    `db:"text"`
	UserID    string    `db:"user_id"`
	GroupID   string    `db:"group_id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
}

type MessageReceived struct {
	SessionID string
	SockConn  *websocket.Conn
	Context   context.Context
	Message   Message
}

var MessageChannel = make(chan MessageReceived, 100)

var DBConn *pgxpool.Pool
var KeycloackDBConn *pgxpool.Pool

// var UserChannel = make(chan User, 100)
var KUserChannel = make(chan KeyCloakUser, 100)

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

type KeycloakGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ParentGroup string `json:"parent_group"`
	RealmID     string `json:"realm_id"`
}

type UserGroupMembership struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

type Group struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Path          string              `json:"path"`
	SubGroupCount int                 `json:"subGroupCount"`
	SubGroups     []string            `json:"subGroups"`
	ParentGroup   string              `json:"parentId"`
	Attributes    map[string][]string `json:"attributes"`
}

type MessageData struct {
	Message string `json:"message"`
	GroupID string `json:"group_id"`
}

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

var GroupUsers []KeyCloakUser
var AllUsers []KeyCloakUser
var RestUsers []KeyCloakUser
var Groups []Group

var Session UserSession
var GroupAdmin KeyCloakUser

var PublicGroupID string
var PersonalGroupID string
var PublicGroup []Group
var PersonalGroup []Group
