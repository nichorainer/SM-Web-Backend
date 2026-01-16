package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "Back-End/internal/adapters/postgresql/sqlc"
)

type Server struct {
    Repo sqlc.Querier
}

type CreateProductRequest struct {
    Name        string `json:"name"`
    PriceInIDR  int    `json:"price_in_idr"`
    Quantity    int    `json:"quantity"`
}

func (s *Server) CreateProduct(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    json.NewDecoder(r.Body).Decode(&req)

    product, err := s.Repo.CreateProduct(r.Context(), sqlc.CreateProductParams{
        Name:       req.Name,
        PriceInIdr: int32(req.PriceInIDR),
        Quantity:   int32(req.Quantity),
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(product)
}

func (s *Server) GetProductByID(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    productID, _ := strconv.Atoi(id)

    product, err := s.Repo.GetProductByID(r.Context(), int32(productID))
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(product)
}