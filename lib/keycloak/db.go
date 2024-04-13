package keycloak

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

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
