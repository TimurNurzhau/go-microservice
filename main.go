package main

import (
	"fmt"
	"go-microservice/handlers"
	"go-microservice/metrics"
	"go-microservice/services"
	"go-microservice/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Инициализация сервиса (in-memory)
	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)
	integrationHandler := handlers.NewIntegrationHandler()

	// Инициализация метрик Prometheus
	metrics.Init()

	// Настройка роутера
	r := mux.NewRouter()

	// Маршрут для метрик (без rate limiting)
	r.Handle("/metrics", metrics.Handler()).Methods("GET")

	// Health Check
	r.HandleFunc("/health", integrationHandler.HealthCheck).Methods("GET")

	// API для пользователей
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users", userHandler.GetAll).Methods("GET")
	api.HandleFunc("/users", userHandler.Create).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetByID).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.Update).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")

	// Применяем middleware ко всем маршрутам
	// MetricsMiddleware должен быть первым, чтобы считать всё
	r.Use(metrics.MetricsMiddleware)
	r.Use(utils.RateLimitMiddleware)

	port := ":8080"
	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}