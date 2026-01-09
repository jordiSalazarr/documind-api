package main

import (
	"database/sql"
	"log"

	"documind.jordi.org/config"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
	"documind.jordi.org/internal/domain/workspace"
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
	// 7. Create test sources
	// 8. Link sources to items
	// 9. Create test relations

	// Example seed data structure:
	_ = workspace.NewWorkspace(
		"Acme Corp",
		common.Slug("acme-corp"),
	)

	_ = knowledge.NewItemType(
		common.NewID(),
		"Workflow",
		common.Slug("workflow"),
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"priority": map[string]interface{}{
					"type": "string",
					"enum": []string{"low", "medium", "high"},
				},
			},
		},
	)

	log.Println("Database seed completed")
}
