package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"
  "time"

  "github.com/go-chi/jwtauth/v5"
  "github.com/jackc/pgx/v5/pgtype"
  "github.com/go-chi/chi/v5"

  repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

// Request payload for creating an order (no user_id needed, taken from JWT)
type CreateOrderRequest struct {
	Items []OrderItemPayload `json:"items"`
}

type OrderItemPayload struct {
	ProductID   int32  `json:"product_id"`
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	UnitPrice   int64  `json:"unit_price"`
	Quantity    int32  `json:"quantity"`
}

// CreateOrder inserts a new order and its items
func (s *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// 🔑 Extract user_id from JWT claims
	_, claims, _ := jwtauth.FromContext(r.Context())
	userIDClaim, ok := claims["user_id"].(string)
	if !ok || userIDClaim == "" {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}
	userID := userIDClaim

	// Generate order number (simple timestamp-based)
	orderNumber := "ORD-" + strconv.FormatInt(time.Now().Unix(), 10)

	// Insert order
	order, err := s.Repo.CreateOrder(r.Context(), repo.CreateOrderParams{
		OrderNumber: orderNumber,
		CustomerID:  pgtype.Text{String: userID, Valid: true},
		CreatedBy:   pgtype.Text{String: userID, Valid: true},
		TotalAmount: pgtype.Int8{Int64: 0, Valid: true},
		Status: pgtype.Text{String: "pending", Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var total int64 = 0

	// Insert order items
	for _, item := range req.Items {
		lineTotal := int64(item.Quantity) * item.UnitPrice
		total += lineTotal

		_, err := s.Repo.CreateOrderItem(r.Context(), repo.CreateOrderItemParams{
      OrderID:     pgtype.Int4{Int32: order.ID, Valid: true},
      ProductID:   pgtype.Int4{Int32: item.ProductID, Valid: true},
      ProductCode: pgtype.Text{String: item.ProductCode, Valid: true},
      ProductName: pgtype.Text{String: item.ProductName, Valid: true},
      UnitPrice:   pgtype.Int8{Int64: item.UnitPrice, Valid: true},
      Quantity:    pgtype.Int4{Int32: item.Quantity, Valid: true},
      LineTotal:   pgtype.Int8{Int64: lineTotal, Valid: true},
    })

		if err != nil {
			http.Error(w, "failed to add order item: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}


func (s *Server) GetOrderByID(w http.ResponseWriter, r *http.Request) {
  idStr := chi.URLParam(r, "id")
  id, _ := strconv.Atoi(idStr)
  row, err := s.Repo.GetOrderByID(r.Context(), int32(id))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  json.NewEncoder(w).Encode(row)
}
