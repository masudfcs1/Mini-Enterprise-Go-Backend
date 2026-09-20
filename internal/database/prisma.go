package database

import (
	"fmt"
	"log"

	"go-mini-setup/internal/database/db"
)

// Database wraps the generated Prisma client.
type Database struct {
	Client *db.PrismaClient
}

// New initializes and connects the Prisma client to the database.
func New() (*Database, error) {
	client := db.NewClient()

	if err := client.Prisma.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database via Prisma: %w", err)
	}

	log.Println("[Database] Successfully connected to PostgreSQL via Prisma")
	return &Database{Client: client}, nil
}

// Close disconnects the Prisma client.
func (d *Database) Close() error {
	if d.Client == nil {
		return nil
	}
	log.Println("[Database] Disconnecting Prisma Client...")
	return d.Client.Prisma.Disconnect()
}
