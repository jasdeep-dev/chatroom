package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func migrateDatabase(ctx context.Context) {
	migrationsDir := "db/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		fmt.Println("Error reading migrations directory:", err)
		return
	}

	// Iterate over each .sql file
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			// Read SQL queries from file
			queries, err := readSQLFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				fmt.Printf("Error reading SQL file %s: %v\n", file.Name(), err)
				continue
			}

			// Execute each query
			for _, query := range queries {
				_, err := DBConn.Exec(ctx, query)
				if err != nil {
					fmt.Printf("Error executing query from %s: %v\n", file.Name(), err)
					continue
				}
			}

			fmt.Printf("Queries from %s executed successfully.\n", file.Name())
		}
	}

	fmt.Println("All migrations executed successfully.")
}

func databaseUp(ctx context.Context) {

}

// Function to read SQL queries from file
func readSQLFile(filename string) ([]string, error) {
	// Open the SQL file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Read the file contents
	content := make([]byte, stat.Size())
	_, err = file.Read(content)
	if err != nil {
		return nil, err
	}

	// Split the file contents by semicolon to separate queries
	queries := strings.Split(string(content), ";")

	// Remove any empty strings
	var cleanedQueries []string
	for _, query := range queries {
		if strings.TrimSpace(query) != "" {
			cleanedQueries = append(cleanedQueries, query)
		}
	}

	return cleanedQueries, nil
}
