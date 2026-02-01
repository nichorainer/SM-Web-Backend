package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

type CreateOrderParams struct {
	OrderNumber  string      `json:"order_number"`
	CustomerName string      `json:"customer_name"`
	TotalAmount  pgtype.Int4 `json:"total_amount"`
	Status       string      `json:"status"`
	CreatedBy    string      `json:"created_by"`
}

// CreateOrder inserts a new order and its items
func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderParams
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	arg := repo.CreateOrderParams{
		OrderNumber:  req.OrderNumber,
		CustomerName: req.CustomerName,
		TotalAmount:  req.TotalAmount,
		Status:       req.Status,
		CreatedBy:    pgtype.Text{String: req.CreatedBy, Valid: req.CreatedBy != ""},
	}

	p, err := s.Repo.CreateOrder(r.Context(), arg)

	if err != nil {
		http.Error(w, "failed to create product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// ListProducts returns a paginated list of products.
func (s *Server) ListOrders(w http.ResponseWriter, r *http.Request) {
	params := repo.ListOrdersParams{
		Limit:  100,
		Offset: 0,
	}

	products, err := s.Repo.ListOrders(r.Context(), params)
	if err != nil {
		http.Error(w, "failed to list products", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}