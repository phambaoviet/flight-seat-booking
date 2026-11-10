package router

import (
	"flight-booking/internal/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(flightHandler *handler.FlightHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Route("/flights", func(r chi.Router) {
				r.Post("/search", flightHandler.SearchFlights)
			})
		})
	})

	return r

}
