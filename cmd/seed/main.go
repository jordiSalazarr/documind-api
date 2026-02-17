package main

import (
	"database/sql"
	"log"

	"documind.jordi.org/config"
	shared "documind.jordi.org/internal/shared/domain"
	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Starting database seed...")

	// TODO: Implement seeding logic
	// 1. Create test workspace
	// 2. Create test users/members
	// 3. Create test projects
	// 4. Create test item types
	// 5. Create test relation types
	// 6. Create test items with versions

	// Example seed data structure:
	_ = wsdomain.NewWorkspace(
		"Acme Corp",
		shared.Slug("acme-corp"),
	)

	_ = db // suppress unused

	log.Println("Database seed completed")
}
