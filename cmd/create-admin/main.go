package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pertisk-tech/pertisk-chart/pkg/auth"
)

func main() {
	var (
		username = flag.String("username", "", "Admin username (required)")
		email    = flag.String("email", "", "Admin email (required)")
		password = flag.String("password", "", "Admin password (required)")
		dataDir  = flag.String("data-dir", "./data", "Data directory for user storage")
		dbType   = flag.String("db-type", "sqlite", "Database type: sqlite, postgres, or file")
		dbDSN    = flag.String("db-dsn", "", "Database connection string (DSN). For SQLite: file path (default: ./data/users.db). For PostgreSQL: connection string")
	)
	flag.Parse()

	if *username == "" || *email == "" || *password == "" {
		log.Fatal("Error: username, email, and password are required")
	}

	// Validate user input
	if err := auth.ValidateUser(*username, *email, *password); err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Initialize user store
	dbTypeValue := auth.DatabaseType(*dbType)
	if dbTypeValue != auth.DatabaseTypeSQLite && dbTypeValue != auth.DatabaseTypePostgres && dbTypeValue != auth.DatabaseTypeFile {
		log.Fatalf("Invalid database type: %s. Must be sqlite, postgres, or file", *dbType)
	}

	// Set default DSN if not provided
	dsn := *dbDSN
	if dsn == "" {
		switch dbTypeValue {
		case auth.DatabaseTypeSQLite:
			dsn = filepath.Join(*dataDir, "users.db")
		case auth.DatabaseTypePostgres:
			dsn = os.Getenv("DATABASE_URL")
			if dsn == "" {
				log.Fatal("PostgreSQL requires --db-dsn flag or DATABASE_URL environment variable")
			}
		}
	}

	userStore, err := auth.NewUserStore(dbTypeValue, dsn, *dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize user store: %v", err)
	}

	// Check if user already exists
	existingUser, err := userStore.GetUserByUsername(*username)
	if err == nil {
		// User exists, check if already admin
		if existingUser.IsAdmin {
			log.Printf("User '%s' is already an admin", *username)
			return
		}
		// Promote existing user to admin
		existingUser.IsAdmin = true
		if err := userStore.UpdateUser(existingUser); err != nil {
			log.Fatalf("Failed to promote user to admin: %v", err)
		}
		log.Printf("Successfully promoted user '%s' to admin", *username)
		return
	}

	// User doesn't exist, create new admin user
	passwordHash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	userID, err := auth.GenerateUserID()
	if err != nil {
		log.Fatalf("Failed to generate user ID: %v", err)
	}

	user := &auth.User{
		ID:           userID,
		Username:     *username,
		Email:        *email,
		PasswordHash: passwordHash,
		IsAdmin:      true, // Set as admin
	}

	if err := userStore.CreateUser(user); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Successfully created admin user:\n")
	fmt.Printf("  Username: %s\n", *username)
	fmt.Printf("  Email: %s\n", *email)
	fmt.Printf("  Admin: true\n")
}

