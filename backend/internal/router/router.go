package router

import (
	"flight-booking/internal/handler"
	authmw "flight-booking/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
    flightHandler *handler.FlightHandler,
    holdHandler *handler.HoldHandler,
    authHandler *handler.AuthHandler,
    authMiddleware *authmw.AuthMiddleware,
    bookingHandler *handler.BookingHandler,
) http.Handler {
    r := chi.NewRouter()

    r.Use(chimiddleware.Logger)

    r.Route("/api/v1", func(r chi.Router) {
        // Public routes
        r.Route("/auth", func(r chi.Router) {
            r.Post("/register", authHandler.Register)
            r.Post("/login", authHandler.Login)
        })

        r.Route("/flights", func(r chi.Router) {
            // Public
            r.Post("/search", flightHandler.SearchFlights)

            // Protected
            r.With(authMiddleware.Authenticate).
                Post("/{flight_id}/hold", holdHandler.CreateHold)
        })

        // Protected routes
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware.Authenticate)

            r.Post("/holds/release", holdHandler.ReleaseHold)

            r.Route("/bookings", func(r chi.Router) {
                r.Post("/confirm", bookingHandler.ConfirmBooking)
                r.Post("/cancel", bookingHandler.CancelBooking)
            })
        })
    })

    return r
}