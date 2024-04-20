package keycloak

import (
	"chatroom/app"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func EstablishKeyCloakConnection(ctx context.Context) *pgxpool.Pool {
	// Define the connection parameters

	config, err := pgxpool.ParseConfig("")
	if err != nil {
		log.Fatalf("Failed to parse config: %v\n", err)
	}

	portStr := os.Getenv("KEYCLOAK_PORT")
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		log.Fatal("Error parsing port:", err)
	}

	config.ConnConfig.User = os.Getenv("KEYCLOAK_USER")
	config.ConnConfig.Password = os.Getenv("KEYCLOAK_PASSWORD")
	config.ConnConfig.Host = os.Getenv("KEYCLOAK_HOST")
	config.ConnConfig.Port = uint16(port)
	config.ConnConfig.Database = os.Getenv("KEYCLOAK_DB")
	config.MaxConns = 10

	// Use config to establish the connection
	KeycloackDBConn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v\n", err)
	}

	fmt.Println("Connection established")
	return KeycloackDBConn
}

type GroupService interface {
	GetGroups() ([]app.KeycloakGroup, error)
}

type KeycloakService struct {
	AccessToken string
	URL         string
	Realm       string
}

func NewKeycloakService() *KeycloakService {
	SetAdminToken()
	return &KeycloakService{
		AccessToken: os.Getenv("ADMIN_ACCESS_TOKEN"),
		URL:         os.Getenv("KEYCLOAK_URL"),
		Realm:       os.Getenv("REALM_NAME"),
	}
}

func SetAdminToken() {
	// Define the URL for fetching the access token
	tokenURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_URL"),
	)

	// Define the form data for the token request
	data := strings.NewReader("client_id=admin-cli&username=admin&password=admin&grant_type=password")

	// Create a new HTTP POST request to fetch the access token
	tokenReq, err := http.NewRequest("POST", tokenURL, data)
	if err != nil {
		fmt.Println("Error creating token request:", err)
		return
	}

	// Set the content type header
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the token request
	client := &http.Client{}
	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		fmt.Println("Error making token request:", err)
		return
	}
	defer tokenResp.Body.Close()

	// Read the token response body
	tokenBody, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		fmt.Println("Error reading token response body:", err)
		return
	}

	// Parse the token response JSON
	var tokenData map[string]interface{}
	if err := json.Unmarshal(tokenBody, &tokenData); err != nil {
		fmt.Println("Error parsing token response JSON:", err)
		return
	}

	// Extract the access token
	accessToken, ok := tokenData["access_token"].(string)
	if !ok {
		fmt.Println("Error accessing access token from response")
		return
	}

	// Set environment variable
	os.Setenv("ADMIN_ACCESS_TOKEN", accessToken)
}
