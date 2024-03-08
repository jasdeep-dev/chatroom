package lib

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

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

var KeyCloakUser KeyCloakUserInfo

func Oauth2Config(ctx context.Context) (oauth2.Config, *oidc.IDTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx,
		fmt.Sprintf("%s/realms/%s", os.Getenv("KEYCLOAK_URL"), os.Getenv("REALM_NAME")))

	if err != nil {
		return oauth2.Config{}, nil, err
	}

	config := oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Endpoint:     provider.Endpoint(),
		RedirectURL:  os.Getenv("APP_URL") + "/oauth2",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	oidcConfig := &oidc.Config{
		ClientID: config.ClientID,
	}

	verifier := provider.Verifier(oidcConfig)
	return config, verifier, nil
}

func createNewProvider(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	config, _, err := Oauth2Config(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	config, verifier, err := Oauth2Config(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	state, err := r.Cookie("state")
	if err != nil {
		log.Println("state not found", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		log.Println("state did not match")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		log.Println("Failed to exchange token:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Println("No id_token field in oauth2 token.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Println("Failed to verify ID Token:", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		log.Println("nonce not found", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if idToken.Nonce != nonce.Value {
		log.Println("nonce did not match")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	idTokenClaims := IDTokenClaims{}

	if err := idToken.Claims(&idTokenClaims); err != nil {
		log.Println("Error in idtoken claims", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    idTokenClaims.Sid,
		Path:     "/",
		HttpOnly: true,
	})

	user := getKeyCloakUserInfo(oauth2Token.AccessToken)

	newUser := User{
		Name:              idTokenClaims.Name,
		IsOnline:          true,
		Theme:             Settings.DefaultTheme,
		PreferredUsername: idTokenClaims.PreferredUsername,
		GivenName:         idTokenClaims.GivenName,
		FamilyName:        idTokenClaims.FamilyName,
		Email:             idTokenClaims.Email,
	}

	var currentUser User

	currentUser, err = FindUserByEmail(newUser.Email)
	if err != nil {
		currentUser = insertUser(newUser)
	}

	UserSessions[idTokenClaims.Sid] = UserSession{
		ID:           currentUser.ID,
		Name:         idTokenClaims.Name,
		AccessToken:  oauth2Token.AccessToken,
		KeyCloakUser: user,
		LoggedInAt:   time.Now(),
	}

	DeleteCookie("state", w)
	DeleteCookie("nonce", w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

func getKeyCloakUserInfo(access_token string) KeyCloakUserInfo {

	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("REALM_NAME"),
	)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}

	req.Header.Set("Authorization", "Bearer "+access_token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request: ", err)
	}
	defer resp.Body.Close()

	var keyCloakUser KeyCloakUserInfo

	err = json.NewDecoder(resp.Body).Decode(&keyCloakUser)
	if err != nil {
		log.Println("Error decoding JSON: ", err)
	}
	return keyCloakUser
}
