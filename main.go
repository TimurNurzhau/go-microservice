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

	r.Handle("/metrics", metrics.Handler()).Methods("GET")
	r.HandleFunc("/health", integrationHandler.HealthCheck).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(utils.RateLimitMiddleware)

	api.HandleFunc("/users", metrics.Instrument("/api/users", userHandler.GetAll)).Methods("GET")
	api.HandleFunc("/users", metrics.Instrument("/api/users", userHandler.Create)).Methods("POST")
	api.HandleFunc("/users/{id}", metrics.Instrument("/api/users/{id}", userHandler.GetByID)).Methods("GET")
	api.HandleFunc("/users/{id}", metrics.Instrument("/api/users/{id}", userHandler.Update)).Methods("PUT")
	api.HandleFunc("/users/{id}", metrics.Instrument("/api/users/{id}", userHandler.Delete)).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
