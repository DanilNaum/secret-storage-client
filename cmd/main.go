// Package main is the entry point for the Secret Storage Client application.
// This application provides a terminal-based user interface for managing
// secret records with support for local storage, server synchronization,
// and intelligent caching.

package main

import (
	"log"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/handlers"
	"github.com/DanilNaum/secret-storage-client/internal/models"
	"github.com/DanilNaum/secret-storage-client/internal/storage"
	"github.com/DanilNaum/secret-storage-client/internal/ui"
	"github.com/DanilNaum/secret-storage-client/pkg/cache"
)

// main is the application entry point.
// It sets up the dependency injection container, creates all necessary components,
// and starts the user interface. The function follows the dependency injection
// pattern to ensure loose coupling between components.
//
// The initialization process:
// 1. Create storage layer with test data
// 2. Create generic cache for server records
// 3. Create handlers with injected dependencies
// 4. Create UI with injected dependencies
// 5. Start the application event loop
//
// If any error occurs during startup or execution, the application will
// terminate with a fatal error message.
func main() {
	// Create storage layer
	// This provides both local record storage and server record metadata management
	stor := storage.NewStorage()

	// Create generic cache for server records
	// Uses the configurable default TTL from constants for automatic expiration
	serverCache := cache.NewCache[models.Record](constants.DefaultCacheTTL)

	// Create handlers with injected dependencies
	// The handlers implement all business logic and coordinate between storage and cache
	eventHandlers := handlers.NewDefaultHandlers(stor, serverCache)

	// Create UI with injected dependencies
	// The UI depends only on interfaces, making it testable and flexible
	app := ui.NewApp(stor, eventHandlers)

	// Start the application
	// This enters the main event loop and blocks until the application is terminated
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
