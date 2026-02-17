package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/go-chi/chi/v5"

	repo "github.com/nichorainer/backend-go/internal/adapters/postgresql/sqlc"
)

// CreateOrderParams represents the JSON payload from the frontend
type CreateOrderParams struct {
	OrderNumber   string `json:"order_number"`
	IdFromProduct int32  `json:"id_from_product"` // FK ke products.id
	ProductID     string `json:"product_id"`      // kode produk
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

	createdAt, err := time.Parse(time.RFC3339, req.CreatedAt)
	if err != nil {
		http.Error(w, "invalid created_at format, must be RFC3339", http.StatusBadRequest)
		return
	}

	normalizedStatus := strings.ToLower(req.Status)
	if normalizedStatus == "shipped" {
		normalizedStatus = "shipping"
	}

	arg := repo.CreateOrderParams{
		OrderNumber:   req.OrderNumber,
		IDFromProduct: pgtype.Int4{Int32: req.IdFromProduct, Valid: true},
		ProductID:     pgtype.Text{String: req.ProductID, Valid: true},
		CustomerName:  req.CustomerName,
		Platform:      req.Platform,
		Destination:   req.Destination,
		TotalAmount:   pgtype.Int4{Int32: req.TotalAmount, Valid: true},
		Status:        normalizedStatus,
		PriceIdr:      pgtype.Int4{Int32: req.PriceIDR, Valid: true},
		CreatedAt:     pgtype.Timestamptz{Time: createdAt, Valid: true},
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

// generateOrderNumber generates the next order number in the format #000001
func generateOrderNumber(ctx context.Context, repo repo.Querier) (string, error) {
	last, err := repo.GetLastOrderNumber(ctx)
	if err != nil {
		// if there aren't any orders yet, start with #000001
		if errors.Is(err, pgx.ErrNoRows) {
			return "#000001", nil
		}
		return "", err
	}

	// strip '#'
	numStr := strings.TrimPrefix(last, "#")

	// parse to int
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("invalid order_number format: %s", last)
	}

	// increment
	next := num + 1

	// format to 6 digit with leading zero
	return fmt.Sprintf("#%06d", next), nil
}

// GetNextOrderNumber handles the HTTP request to get the next order number
func (s *Server) GetNextOrderNumber(w http.ResponseWriter, r *http.Request) {
    next, err := generateOrderNumber(r.Context(), s.Repo)
    if err != nil {
        http.Error(w, "failed to generate order number: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "order_number": next,
    })
}

// ListOrdersWithProduct lists all orders along with their associated product details
func (s *Server) ListOrdersWithProduct(w http.ResponseWriter, r *http.Request) {
	params := repo.ListOrdersWithProductParams{
		Limit:  100,
		Offset: 0,
	}

	orders, err := s.Repo.ListOrdersWithProduct(r.Context(), params)
	if err != nil {
		http.Error(w, "failed to list orders with product", http.StatusInternalServerError)
		return
	}

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
	idStr := chi.URLParam(r, "id")
    if idStr == "" {
        http.Error(w, "missing order id", http.StatusBadRequest)
        return
    }

	// get order_id from query param
    orderIDInt, err := strconv.Atoi(idStr)
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

	// get query sqlc for update
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
    json.NewEncoder(w).Encode(order)
}

// DeleteOrder deletes an existing order by ID
func (s *Server) DeleteOrder(w http.ResponseWriter, r *http.Request) {
    // get order_id from param
    idStr := chi.URLParam(r, "id")
    if idStr == "" {
        http.Error(w, "missing order id", http.StatusBadRequest)
        return
    }

    orderIDInt, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "invalid order id", http.StatusBadRequest)
        return
    }

	// get order before delete to check if it exists
	order, err := s.Repo.GetOrderByID(r.Context(), int32(orderIDInt))
    if err != nil {
        http.Error(w, "order not found: "+err.Error(), http.StatusNotFound)
        return
    }

    // delete order
    err = s.Repo.DeleteOrder(r.Context(), int32(orderIDInt))
    if err != nil {
        http.Error(w, "failed to delete order: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(order)
}

// GetTopProductsFromOrders returns top 5 products based on total_amount
func (s *Server) GetTopProductsFromOrders(w http.ResponseWriter, r *http.Request) {
    products, err := s.Repo.GetTopProductsFromOrders(r.Context())
    if err != nil {
        http.Error(w, "failed to get top products: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}