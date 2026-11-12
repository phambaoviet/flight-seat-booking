package main

import (
	"context"
	"errors"
	"flight-booking/internal/duffel"
	"flight-booking/internal/handler"
	"flight-booking/internal/middleware"
	"flight-booking/internal/router"
	"flight-booking/internal/service"
	"flight-booking/internal/store"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func loadAPITokenFromEnv() (string, error) {
	apiToken := strings.TrimSpace(os.Getenv("DUFFEL_API_TOKEN"))

	if apiToken == "" {
		return "", errors.New("DUFFEL_API_TOKEN is not set")
	}

	return apiToken, nil
}

func StartServer() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()
	// Initialize PostgreSQL connection
	dbSource := os.Getenv("DB_SOURCE")
	if dbSource == "" {
		log.Fatal("DB_SOURCE environment variable is not set")
	}

	db, err := pgxpool.New(ctx, dbSource)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	log.Println("Connected to the database successfully")

	// Create UserStore
	userStore := store.NewStore(db)

	// Initialize JWT TokenService
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	expiresInStr := os.Getenv("JWT_EXPIRES_IN")
	if expiresInStr == "" {
		log.Fatal("JWT_EXPIRES_IN environment variable is not set")
	}
	expiresIn, err := time.ParseDuration(expiresInStr)
	if err != nil || expiresIn <= 0 {
		log.Fatal("JWT_EXPIRES_IN must be a valid positive duration")
	}
	tokenService := service.NewTokenService([]byte(jwtSecret), expiresIn)

	// Initialize AuthService
	authService := service.NewAuthService(userStore, tokenService)

	// Initialize AuthHandler
	authHandler := handler.NewAuthHandler(authService)

	// Initialize AuthMiddleware
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// Initialize HoldService and HoldHandler
	holdService := service.NewHoldService(userStore)
	holdHandler := handler.NewHoldHandler(holdService)

	// Load API token from environment
	apiToken, err := loadAPITokenFromEnv()
	if err != nil {
		log.Fatalf("Error loading API token: %v", err)
	}
	log.Printf("Duffel token loaded: %t", apiToken != "")

	// Create Duffel client
	duffelClient, err := duffel.NewClient(apiToken)
	if err != nil {
		log.Fatalf("Failed to create Duffel client: %v", err)
	}

	// Create FlightService
	flightService := service.NewFlightService(duffelClient)

	// Create FlightHandler
	flightHandler := handler.NewFlightHandler(flightService)

	// Create router
	appRouter := router.NewRouter(flightHandler, holdHandler, authHandler, authMiddleware)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: appRouter,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
func main() {
	log.Println("Starting Flight Booking API...")
	StartServer()
}
