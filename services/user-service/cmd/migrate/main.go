package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
)

func main() {
	migrationsPath := flag.String("path", "internal/migrations", "path to migrations directory")
	direction := flag.String("direction", "up", "migration direction (up or down)")
	steps := flag.Int("steps", 0, "number of migrations to apply (0 for all)")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	if direction == nil || (*direction != "up" && *direction != "down") {
		log.Fatal("direction must be 'up' or 'down'")
	}

	source := "file://" + *migrationsPath
	m, err := migrate.New(source, databaseURL)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	fmt.Printf("Running migrations (direction=%s, steps=%d)...\n", *direction, *steps)

	if *direction == "up" {
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	} else {
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
	}

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Migration completed successfully")
}
