package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"go-microservice/handlers"
	"go-microservice/metrics"
	"go-microservice/services"
	"go-microservice/storage"
	"go-microservice/utils"

	"github.com/gorilla/mux"
)

func main() {
	if _, err := storage.InitMinIO(); err != nil {
		log.Printf("Предупреждение: S3-хранилище не инициализировано: %v", err)
	} else {
		log.Println("S3-хранилище успешно инициализировано")
	}

	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)
	integrationHandler := handlers.NewIntegrationHandler()

	metrics.Init()

	router := mux.NewRouter()
	router.Handle("/metrics", metrics.Handler()).Methods(http.MethodGet)
	router.HandleFunc("/health", integrationHandler.HealthCheck).Methods(http.MethodGet)

	api := router.PathPrefix("/api").Subrouter()

	// Rate limiter находится внутри instrumentation, поэтому ответы 429
	// также учитываются в счётчиках ошибок и histogram.
	api.Handle(
		"/users",
		metrics.Instrument(
			http.MethodGet,
			"/api/users",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.GetAll)),
		),
	).Methods(http.MethodGet)

	api.Handle(
		"/users",
		metrics.Instrument(
			http.MethodPost,
			"/api/users",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.Create)),
		),
	).Methods(http.MethodPost)

	api.Handle(
		"/users/{id}",
		metrics.Instrument(
			http.MethodGet,
			"/api/users/{id}",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.GetByID)),
		),
	).Methods(http.MethodGet)

	api.Handle(
		"/users/{id}",
		metrics.Instrument(
			http.MethodPut,
			"/api/users/{id}",
			utils.RateLimitMiddleware(http.HandlerFunc(userHandler.Update)),
		),
	).Methods(http.MethodPut)

	api.Handle(
		"/users/{id}",
		metrics.Instrument(
			http.MethodDelete,
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
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("HTTP server listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}
