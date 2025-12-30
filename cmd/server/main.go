package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pertisk-tech/pertisk-chart/pkg/api"
	"github.com/pertisk-tech/pertisk-chart/pkg/auth"
	"github.com/pertisk-tech/pertisk-chart/pkg/storage"
)

func main() {
	var (
		port           = flag.String("port", "7080", "Port to listen on")
		storageBackend = flag.String("storage", "local", "Storage backend (local)")
		storageRootDir = flag.String("storage-local-rootdir", "./chartstorage", "Local storage root directory")
		dataDir        = flag.String("data-dir", "./data", "Data directory for user storage")
		dbType         = flag.String("db-type", "sqlite", "Database type: sqlite, postgres, or file")
		dbDSN          = flag.String("db-dsn", "", "Database connection string (DSN). For SQLite: file path (default: ./data/users.db). For PostgreSQL: connection string")
		jwtSecret      = flag.String("jwt-secret", "", "JWT secret key (default: auto-generated, set JWT_SECRET env var)")
		enableMetrics  = flag.Bool("enable-metrics", false, "Enable Prometheus metrics")
		debug          = flag.Bool("debug", false, "Enable debug mode")
		enableHTTP3    = flag.Bool("enable-http3", false, "Enable HTTP/3 support (requires TLS certificates)")
		tlsCertFile    = flag.String("tls-cert", "", "Path to TLS certificate file (required for HTTP/3)")
		tlsKeyFile     = flag.String("tls-key", "", "Path to TLS private key file (required for HTTP/3)")
		enableZstd     = flag.Bool("enable-zstd", true, "Enable zstd compression")
		webDir         = flag.String("web-dir", "", "Web directory path (default: ./web for development, /usr/share/pertisk-chart/web for packages)")
	)
	flag.Parse()

	if *debug {
		log.Println("Debug mode enabled")
	}

	// Initialize storage
	store, err := storage.NewLocalStorage(*storageRootDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
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

	log.Printf("Using %s database for user storage", *dbType)
	if dbTypeValue == auth.DatabaseTypeSQLite || dbTypeValue == auth.DatabaseTypePostgres {
		log.Printf("Database DSN: %s", dsn)
	}

	// Initialize config store (only for database types)
	var configStore auth.ConfigStore
	if dbTypeValue == auth.DatabaseTypeSQLite || dbTypeValue == auth.DatabaseTypePostgres {
		db, err := auth.OpenDatabase(dbTypeValue, dsn)
		if err != nil {
			log.Fatalf("Failed to open database for config store: %v", err)
		}
		configStore, err = auth.NewDBConfigStore(db)
		if err != nil {
			log.Fatalf("Failed to initialize config store: %v", err)
		}
	} else {
		// For file-based storage, create a simple in-memory config store
		// In production, you might want to use a file-based config store
		log.Println("Warning: Config store not available with file-based user storage. Using default domain.")
		configStore = nil
	}

	// Set JWT secret
	if *jwtSecret != "" {
		auth.SetJWTSecret(*jwtSecret)
	} else if secret := os.Getenv("JWT_SECRET"); secret != "" {
		auth.SetJWTSecret(secret)
	} else {
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable or --jwt-secret flag for production.")
	}

	// Set default web directory if not provided
	webDirPath := *webDir
	if webDirPath == "" {
		// Check if /usr/share/pertisk-chart/web exists (package installation)
		if _, err := os.Stat("/usr/share/pertisk-chart/web"); err == nil {
			webDirPath = "/usr/share/pertisk-chart/web"
		} else {
			// Default to ./web for development
			webDirPath = "./web"
		}
	}

	// Create API server
	server := api.NewServer(store, userStore, configStore, &api.Config{
		Port:          *port,
		EnableMetrics: *enableMetrics,
		Debug:         *debug,
		EnableHTTP3:   *enableHTTP3,
		TLSCertFile:   *tlsCertFile,
		TLSKeyFile:    *tlsKeyFile,
		EnableZstd:    *enableZstd,
		WebDir:        webDirPath,
	})

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting Pertisk Chart Server on %s", addr)
	log.Printf("Storage backend: %s", *storageBackend)
	log.Printf("Storage root: %s", *storageRootDir)
	if *enableHTTP3 {
		log.Printf("HTTP/3 enabled (TLS required)")
	}
	if *enableZstd {
		log.Printf("zstd compression enabled")
	}

	if err := server.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
