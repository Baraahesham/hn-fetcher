package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DBConnection *pgxpool.Pool

func InitDBConnection(dbUrl string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	DBConnection = pool
	fmt.Println("Connected to PostgreSQL")
}
