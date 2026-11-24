package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-microservice/handlers"
	"go-microservice/metrics"
	"go-microservice/services"
	"go-microservice/storage"
	"go-microservice/utils"

	"github.com/gorilla/mux"
)

func main() {
	_, err := storage.InitMinIO()
	if err != nil {
		log.Printf("Предупреждение: S3-хранилище не инициализировано: %v", err)
	} else {
		log.Println("S3-хранилище успешно инициализировано")
	}

	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)
	integrationHandler := handlers.NewIntegrationHandler()

	metrics.Init()

	r := mux.NewRouter()

	r.Handle("/metrics", metrics.Handler()).Methods(http.MethodGet)
	r.HandleFunc("/health", integrationHandler.HealthCheck).Methods(http.MethodGet)

	api := r.PathPrefix("/api").Subrouter()

	// Метрики оборачивают rate limiter, чтобы учитывать также ответы 429.
	api.Handle(
		"/users",
		metrics.Instrument(
			"/api/users",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.GetAll)),
		),
	).Methods(http.MethodGet)

	api.Handle(
		"/users",
		metrics.Instrument(
			"/api/users",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.Create)),
		),
	).Methods(http.MethodPost)

	api.Handle(
		"/users/{id}",
		metrics.Instrument(
			"/api/users/{id}",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.GetByID)),
		),
	).Methods(http.MethodGet)

	api.Handle(
		"/users/{id}",
		metrics.Instrument(
			"/api/users/{id}",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.Update)),
		),
	).Methods(http.MethodPut)

	api.Handle(
		"/users/{id}",
		metrics.Instrument(
			"/api/users/{id}",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.Delete)),
		),
	).Methods(http.MethodDelete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5_000_000_000, // 5 секунд
	}

	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(server.ListenAndServe())
}
