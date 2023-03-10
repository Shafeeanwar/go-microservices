package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	//specify who is allowed to connect
	//Middleware for handling cors
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"POST", "PUT", "GET", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//Can be used to check health status of service
	mux.Use(middleware.Heartbeat("/ping"))

	mux.Post("/api", app.Broker)
	mux.Post("/api/handle", app.HandleSubmission)
	mux.Post("/api/log-grpc", app.LogViaGRPC)

	return mux
}
