package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"strings"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

// CreateOrderParams represents the JSON payload from the frontend
type CreateOrderParams struct {
	OrderNumber   string `json:"order_number"`
	ProductID     int32  `json:"product_id"`
	CustomerName  string `json:"customer_name"`
	Platform      string `json:"platform"`
	Destination   string `json:"destination"`
	TotalAmount   int32  `json:"total_amount"`
	Status        string `json:"status"`
	PriceIDR      int32  `json:"price_idr"`
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

    // parse CreatedAt string dari FE ke time.Time
    createdAt, err := time.Parse(time.RFC3339, req.CreatedAt)
    if err != nil {
        http.Error(w, "invalid created_at format, must be RFC3339", http.StatusBadRequest)
        return
    }

	// normalize status
	normalizedStatus := strings.ToLower(req.Status)
	if normalizedStatus == "shipped" {
		normalizedStatus = "shipping"
	}

	arg := repo.CreateOrderParams{
		OrderNumber:  req.OrderNumber,
		ProductID:    pgtype.Int4{Int32: req.ProductID, Valid: true},
		CustomerName: req.CustomerName,
		Platform:     req.Platform,
		Destination:  req.Destination,
		TotalAmount:  pgtype.Int4{Int32: req.TotalAmount, Valid: true},
		Status:       normalizedStatus,
		PriceIdr:     pgtype.Int4{Int32: req.PriceIDR, Valid: true},
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

	// normalize status values before sending to FE
	for i := range orders {
		orders[i].Status = strings.ToLower(orders[i].Status)
		if orders[i].Status == "shipped" {
			orders[i].Status = "shipping"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateOrderStatus updates the status of an existing order
func (s *Server) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
    // ambil order_id dari query param
    orderIDStr := r.URL.Query().Get("id")
    if orderIDStr == "" {
        http.Error(w, "missing order id", http.StatusBadRequest)
        return
    }

    // konversi string → int32
    orderIDInt, err := strconv.Atoi(orderIDStr)
    if err != nil {
        http.Error(w, "invalid order id", http.StatusBadRequest)
        return
    }

    // decode payload JSON { "status": "completed" }
    var payload struct {
        Status string `json:"status"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
        return
    }

    // normalize status
    normalizedStatus := strings.ToLower(payload.Status)
    if normalizedStatus == "shipped" {
        normalizedStatus = "shipping"
    }

    // panggil query sqlc untuk update
    arg := repo.UpdateOrderStatusParams{
        ID:     int32(orderIDInt),
        Status: normalizedStatus,
    }

    order, err := s.Repo.UpdateOrderStatus(r.Context(), arg)
    if err != nil {
        http.Error(w, "failed to update order status: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(order); err != nil {
        http.Error(w, "failed to encode response", http.StatusInternalServerError)
    }
}