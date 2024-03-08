package main

import (
	"context"
	"log"

	"chatroom/lib"

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
	lib.DBConn = lib.EstablishConnection(ctx)
	defer lib.DBConn.Close()

	// migrateDatabase(ctx)
	lib.GetUsers(ctx)
	lib.GetMessages(ctx)

	go lib.MessageReceiver()
	go lib.UserReciver()
	go lib.StartWebSocket()
	lib.StartHTTP()
}
