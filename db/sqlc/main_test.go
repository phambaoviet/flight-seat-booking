package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbSource := os.Getenv("DB_SOURCE")
	pool, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("Cannot connect to db:", err)
	}
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Cannot ping db:", err)
	}

	testDB = pool
	code := m.Run()
	pool.Close()
	os.Exit(code)

}
