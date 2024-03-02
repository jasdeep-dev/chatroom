package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/cap/oidc"
)

var oidcRequest *oidc.Req

func createNewProvider() string {
	// Create a new provider config
	pc, err := oidc.NewConfig(
		"http://localhost:9000/realms/chatroom",
		"chatroom",
		"WU3E47c1C2263VqdB4q9QOZIzXFTg9DC",
		[]oidc.Alg{oidc.RS256},
		[]string{"http://localhost:8080/oauth2"},
	)
	if err != nil {
		log.Fatal("error creating config", err)
	}

	// Create a provider
	p, err := oidc.NewProvider(pc)
	if err != nil {
		log.Fatal("error creating provider", err)
	}
	defer p.Done()

	// Create a Request for a user's authentication attempt that will use the
	// authorization code flow.  (See NewRequest(...) using the WithPKCE and
	// WithImplicit options for creating a Request that uses those flows.)
	oidcRequest, err = oidc.NewRequest(2*time.Minute, "http://localhost:8080/oauth2")
	if err != nil {
		log.Fatal("error requesting oidc", err)
	}

	// Create an auth URL
	authURL, err := p.AuthURL(context.Background(), oidcRequest)
	if err != nil {
		log.Fatal("error auth-url", err)
	}
	fmt.Println("open url to kick-off authentication: ", authURL)
	return authURL
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	http.Redirect(w, r, "/user", http.StatusSeeOther)
	pc, err := oidc.NewConfig(
		"http://localhost:9000/realms/chatroom",
		"chatroom",
		"WU3E47c1C2263VqdB4q9QOZIzXFTg9DC",
		[]oidc.Alg{oidc.RS256},
		[]string{"http://localhost:8080/oauth2"},
	)

	if err != nil {
		log.Fatal("error creating config", err)
	}

	p, err := oidc.NewProvider(pc)
	if err != nil {
		log.Fatal("error creating provider", err)
	}
	defer p.Done()

	// Exchange a successful authentication's authorization code and
	// authorization state (received in a callback) for a verified Token.
	t, err := p.Exchange(ctx, oidcRequest, r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Fatal("1 error creating provider", err)
	}
	var claims map[string]interface{}
	if err := t.IDToken().Claims(&claims); err != nil {
		log.Fatal("4 error creating provider", err)
	}

	// Get the user's claims via the provider's UserInfo endpoint
	var infoClaims map[string]interface{}
	err = p.UserInfo(ctx, t.StaticTokenSource(), claims["sub"].(string), &infoClaims)
	if err != nil {
		log.Fatal("2 error creating provider", err)
	}

	resp := struct {
		IDTokenClaims  map[string]interface{}
		UserInfoClaims map[string]interface{}
	}{claims, infoClaims}

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		log.Fatal("3 error creating provider", err)
	}

	// Login:
	// {
	// 	"IDTokenClaims": {
	// 		"acr": "1",
	// 		"at_hash": "qi9S4V8Mnih-hCNsKYTqeg",
	// 		"aud": "chatroom",
	// 		"auth_time": 1709323331,
	// 		"azp": "chatroom",
	// 		"email_verified": true,
	// 		"exp": 1709323631,
	// 		"iat": 1709323331,
	// 		"iss": "http://localhost:9000/realms/chatroom",
	// 		"jti": "28307d3f-2839-4219-a3c5-705533973f9d",
	// 		"nonce": "n_fscahv1t8uzIOA919CJQ",
	// 		"preferred_username": "jasdeep",
	// 		"session_state": "ee71d785-ee31-46cb-9d63-b80f35b76635",
	// 		"sid": "ee71d785-ee31-46cb-9d63-b80f35b76635",
	// 		"sub": "9573ec5f-73e1-4ed4-93d1-baf08fc5c414",
	// 		"typ": "ID"
	// 	},
	// 	"UserInfoClaims": {
	// 		"email_verified": true,
	// 		"preferred_username": "jasdeep",
	// 		"sub": "9573ec5f-73e1-4ed4-93d1-baf08fc5c414"
	// 	}
	// }

	// Cookie("session_id")

	//Register
	// {
	// 	"IDTokenClaims": {
	// 		"acr": "1",
	// 		"at_hash": "AT9HHXlPfj-mFVmH6U-cqw",
	// 		"aud": "chatroom",
	// 		"auth_time": 1709332530,
	// 		"azp": "chatroom",
	// 		"email": "jasdeep@yopmail.com",
	// 		"email_verified": false,
	// 		"exp": 1709332830,
	// 		"family_name": "Kaur",
	// 		"given_name": "Smile",
	// 		"iat": 1709332530,
	// 		"iss": "http://localhost:9000/realms/chatroom",
	// 		"jti": "62b85654-af6a-4eb9-98e9-08aaa599372e",
	// 		"name": "Smile Kaur",
	// 		"nonce": "n_Od8bsyCV2Z6zA48Tfkg7",
	// 		"preferred_username": "jasdeep@yopmail.com",
	// 		"session_state": "68b2a086-482f-4e6f-92b5-83460fa1d2d8",
	// 		"sid": "68b2a086-482f-4e6f-92b5-83460fa1d2d8",
	// 		"sub": "620a4ae4-65c1-402e-8be7-defab24954a1",
	// 		"typ": "ID"
	// 	},
	// 	"UserInfoClaims": {
	// 		"email": "jasdeep@yopmail.com",
	// 		"email_verified": false,
	// 		"family_name": "Kaur",
	// 		"given_name": "Smile",
	// 		"name": "Smile Kaur",
	// 		"preferred_username": "jasdeep@yopmail.com",
	// 		"sub": "620a4ae4-65c1-402e-8be7-defab24954a1"
	// 	}
	// }

}
