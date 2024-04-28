package main

import (
	"context"
	"log"

	"chatroom/app"
	"chatroom/lib"
	"chatroom/lib/keycloak"

	"github.com/joho/godotenv"
)

func main() {
	// Read config file
	err := lib.ReadConfigFromFile("./config.json")
	if err != nil {
		log.Fatal("Could not read the config file: ", err)
	}

	// Connect to database
	err = godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}

	ctx := context.Background()
	app.DBConn = lib.EstablishConnection(ctx)
	defer app.DBConn.Close()

	app.KeycloackDBConn = keycloak.EstablishKeyCloakConnection(ctx)
	defer app.KeycloackDBConn.Close()

	keycloak.SetAdminToken()
	//migrate database
	// lib.MigrateDatabase((ctx))

	// Connect to memcahed
	err = lib.InitCache()
	if err != nil {
		log.Println("Unable to connect to Memcached: ", err)
	}

	go lib.MessageReceiver()
	go lib.UserReciver()
	go lib.StartWebSocket()
	lib.StartHTTP()
}
