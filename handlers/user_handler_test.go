package handlers_test

import (
	"bytes"
	"encoding/json"
	"go-microservice/handlers"
	"go-microservice/models"
	"go-microservice/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func setupRouter() (*mux.Router, *services.UserService) {
	svc := services.NewUserService()
	h := handlers.NewUserHandler(svc)

	r := mux.NewRouter()
	r.HandleFunc("/api/users", h.GetAll).Methods("GET")
	r.HandleFunc("/api/users", h.Create).Methods("POST")
	r.HandleFunc("/api/users/{id}", h.GetByID).Methods("GET")
	r.HandleFunc("/api/users/{id}", h.Update).Methods("PUT")
	r.HandleFunc("/api/users/{id}", h.Delete).Methods("DELETE")
	return r, svc
}

func createUser(t *testing.T, r *mux.Router, name, email string) models.User {
	t.Helper()
	body := []byte(`{"name":"` + name + `","email":"` + email + `"}`)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var u models.User
	if err := json.Unmarshal(rr.Body.Bytes(), &u); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return u
}

func TestCreateAndGet(t *testing.T) {
	r, _ := setupRouter()
	created := createUser(t, r, "Ivan", "ivan@example.com")
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	req := httptest.NewRequest("GET", "/api/users/1", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestInvalidEmail(t *testing.T) {
	r, _ := setupRouter()
	body := []byte(`{"name":"Ivan","email":"not-an-email"}`)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateMissingName(t *testing.T) {
	r, _ := setupRouter()
	body := []byte(`{"name":"","email":"ivan@example.com"}`)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateMalformedJSON(t *testing.T) {
	r, _ := setupRouter()
	body := []byte(`{bad json`)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	r, _ := setupRouter()
	req := httptest.NewRequest("GET", "/api/users/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	r, _ := setupRouter()
	created := createUser(t, r, "Ivan", "ivan@example.com")
	id := created.ID

	body := []byte(`{"name":"Petr","email":"petr@example.com"}`)
	req := httptest.NewRequest("PUT", "/api/users/"+itoa(id), bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d", rr.Code)
	}

	req = httptest.NewRequest("DELETE", "/api/users/"+itoa(id), nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rr.Code)
	}
}

func TestUpdateNotFound(t *testing.T) {
	r, _ := setupRouter()
	body := []byte(`{"name":"Petr","email":"petr@example.com"}`)
	req := httptest.NewRequest("PUT", "/api/users/999", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	r, _ := setupRouter()
	req := httptest.NewRequest("DELETE", "/api/users/999", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
