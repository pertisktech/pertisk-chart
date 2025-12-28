package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pertisk-tech/pertisk-chart/pkg/api"
	"github.com/pertisk-tech/pertisk-chart/pkg/storage"
)

func main() {
	var (
		port            = flag.String("port", "8080", "Port to listen on")
		storageBackend  = flag.String("storage", "local", "Storage backend (local)")
		storageRootDir  = flag.String("storage-local-rootdir", "./chartstorage", "Local storage root directory")
		enableMetrics   = flag.Bool("enable-metrics", false, "Enable Prometheus metrics")
		debug           = flag.Bool("debug", false, "Enable debug mode")
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

	// Create API server
	server := api.NewServer(store, &api.Config{
		Port:          *port,
		EnableMetrics: *enableMetrics,
		Debug:         *debug,
	})

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting Pertisk Chart Server on %s", addr)
	log.Printf("Storage backend: %s", *storageBackend)
	log.Printf("Storage root: %s", *storageRootDir)
	
	if err := server.Start(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

