package main

import (
	"errors"
	"flight-booking/internal/duffel"
	"flight-booking/internal/handler"
	"flight-booking/internal/router"
	"flight-booking/internal/service"
	"log"
	"net/http"
	"os"
	"strings"

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
	appRouter := router.NewRouter(flightHandler)

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
