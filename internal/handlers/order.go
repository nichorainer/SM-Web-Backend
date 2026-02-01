package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

// CreateOrderParams merepresentasikan payload JSON dari FE
type CreateOrderParams struct {
	OrderNumber   string `json:"order_number"`
	CustomerName  string `json:"customer_name"`
	Platform      string `json:"platform"`
	Destination   string `json:"destination"`
	TotalAmount   int32  `json:"total_amount"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderParams
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// parse CreatedAt string ke time.Time
	createdAt, err := time.Parse(time.RFC3339, req.CreatedAt)
	if err != nil {
		http.Error(w, "invalid created_at format, must be RFC3339", http.StatusBadRequest)
		return
	}

	arg := repo.CreateOrderParams{
		OrderNumber:  req.OrderNumber,
		CustomerName: req.CustomerName,
		Platform:     req.Platform,
		Destination:  req.Destination,
		TotalAmount:  pgtype.Int4{Int32: req.TotalAmount, Valid: true},
		Status:       req.Status,
		CreatedAt:    pgtype.Timestamptz{Time: createdAt, Valid: true},
	}

	order, err := s.Repo.CreateOrder(r.Context(), arg)
	if err != nil {
		http.Error(w, "failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// ListOrders returns a paginated list of orders
func (s *Server) ListOrders(w http.ResponseWriter, r *http.Request) {
	params := repo.ListOrdersParams{
		Limit:  100,
		Offset: 0,
	}

	orders, err := s.Repo.ListOrders(r.Context(), params)
	if err != nil {
		http.Error(w, "failed to list orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
