package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func establishConnection(ctx context.Context) *pgx.Conn {
	// Define the connection parameters
	config, err := pgx.ParseConfig("")
	if err != nil {
		log.Fatalf("Failed to parse config: %v\n", err)
	}
	config.User = "chat"
	config.Password = "Chat123#"
	config.Host = "localhost"
	config.Port = 5432
	config.Database = "chatroom"

	// Use config to establish the connection
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v\n", err)
	}

	err = conn.Ping(ctx)
	if err == nil {
		fmt.Println("Connected to Database")
	} else {
		log.Fatalf("Unable to ping the database: %v\n", err)
	}

	return conn
}
