package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  sqlc "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

type Server struct {
  Repo sqlc.Querier
}

type CreateProductRequest struct {
  Name       string `json:"name"`
  PriceInIDR int    `json:"price_in_idr"`
  Quantity   int    `json:"quantity"`
}

func (s *Server) ListProducts(w http.ResponseWriter, r *http.Request) {
  products, err := s.Repo.ListProducts(r.Context())
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  json.NewEncoder(w).Encode(products)
}

func (s *Server) GetProductByID(w http.ResponseWriter, r *http.Request) {
  idStr := chi.URLParam(r, "id")
  id, _ := strconv.Atoi(idStr)
  product, err := s.Repo.FindProductByID(r.Context(), int32(id))
  if err != nil {
    http.Error(w, err.Error(), http.StatusNotFound)
    return
  }
  json.NewEncoder(w).Encode(product)
}

func (s *Server) CreateProduct(w http.ResponseWriter, r *http.Request) {
  var req CreateProductRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "invalid json", http.StatusBadRequest)
    return
  }

  p, err := s.Repo.CreateProduct(r.Context(), sqlc.CreateProductParams{
    Name:       req.Name,
    PriceInIdr: int32(req.PriceInIDR),
    Quantity:   int32(req.Quantity),
  })

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(p)  
}