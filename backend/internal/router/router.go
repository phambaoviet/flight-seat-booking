package router

import (
	"flight-booking/internal/handler"
	authmw "flight-booking/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(flightHandler *handler.FlightHandler, holdHandler *handler.HoldHandler, authHandler *handler.AuthHandler, authMiddleware *authmw.AuthMiddleware, bookingHandler *handler.BookingHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/flights", func(r chi.Router) {
			r.Post("/search", flightHandler.SearchFlights)
		})
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Route("/flights", func(r chi.Router) {
				r.Post("/{flight_id}/hold", holdHandler.CreateHold)
			})
			r.Route("/bookings", func(r chi.Router) {
				r.Post("/confirm", bookingHandler.ConfirmBooking)
			})
			r.Route("/holds", func(r chi.Router) {
				r.Post("/release", holdHandler.ReleaseHold)
			})
		})
	})
	return r
}
