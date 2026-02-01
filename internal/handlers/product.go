package handlers

import (
  "encoding/json"
  "net/http"
  
  "github.com/go-chi/chi/v5"
  
  repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

type Server struct {
  Repo repo.Querier
}

// CreateProductRequest is the expected JSON body for creating a product
type CreateProductRequest struct {
    ProductID    string `json:"product_id"`
    ProductName  string `json:"product_name"`
    SupplierName string `json:"supplier_name"`
    Category     string `json:"category"`
    PriceIdr     int64  `json:"price_idr"`
    Stock        int32  `json:"stock"`
}

// ListProducts returns a paginated list of products.
func (s *Server) ListProducts(w http.ResponseWriter, r *http.Request) {
	params := repo.ListProductsParams{
		Limit:  100,
		Offset: 0,
	}

	products, err := s.Repo.ListProducts(r.Context(), params)
	if err != nil {
		http.Error(w, "failed to list products", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}

// GetProductByID returns a product by product_id (string).
func (s *Server) GetProductByID(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "product_id")
	if productID == "" {
		http.Error(w, "missing product_id", http.StatusBadRequest)
		return
	}

	product, err := s.Repo.GetProductByProductID(r.Context(), productID)
	if err != nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

  json.NewEncoder(w).Encode(product)
}

func (s *Server) CreateProduct(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
        return
    }

    arg := repo.CreateProductParams{
        ProductID:    req.ProductID,
        ProductName:  req.ProductName,
        SupplierName: req.SupplierName,
        Category:     req.Category,
        PriceIdr:     req.PriceIdr,
        Stock:        req.Stock,
    }

    p, err := s.Repo.CreateProduct(r.Context(), arg)
    
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