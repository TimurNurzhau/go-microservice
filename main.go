package main

import (
	"fmt"
	"go-microservice/handlers"
	"go-microservice/metrics"
	"go-microservice/services"
	"go-microservice/storage"
	"go-microservice/utils"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	_, err := storage.InitMinIO()
	if err != nil {
		log.Printf("Предупреждение: MinIO не инициализирован: %v\n", err)
	} else {
		log.Println("MinIO успешно инициализирован")
	}

	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)
	integrationHandler := handlers.NewIntegrationHandler()

	metrics.Init()

	r := mux.NewRouter()

	// /metrics и /health — вне rate limiter
	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	r.HandleFunc("/health", integrationHandler.HealthCheck).Methods("GET")

	// API: rate limiter + метрики
	api := r.PathPrefix("/api").Subrouter()
	api.Use(utils.RateLimitMiddleware)
	api.Use(metrics.MetricsMiddleware)
	api.HandleFunc("/users", userHandler.GetAll).Methods("GET")
	api.HandleFunc("/users", userHandler.Create).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetByID).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.Update).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}