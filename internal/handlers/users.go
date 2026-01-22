package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

type CreateCustomerRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// List all customers
func (s *Server) ListCustomers(w http.ResponseWriter, r *http.Request) {
    customers, err := s.Repo.ListCustomers(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(customers)
}

// Get a single customer by ID
func (s *Server) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, _ := strconv.Atoi(idStr)
    customer, err := s.Repo.FindCustomerByID(r.Context(), int32(id))
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(customer)
}

// Create a new customer
func (s *Server) CreateCustomer(w http.ResponseWriter, r *http.Request) {
    var req CreateCustomerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }
    c, err := s.Repo.CreateCustomer(r.Context(), sqlc.CreateCustomerParams{
        Name:  req.Name,
        Email: req.Email,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(c)
}