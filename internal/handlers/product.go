package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"
  "math"
  "errors"
  "database/sql"
  
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

// ListProducts returns either full products or simplified options
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

	// cek query param ?mode=options
	mode := r.URL.Query().Get("mode")
	if mode == "options" {
		// FE hanya butuh product_id + product_name
		type ProductOption struct {
			ProductID   string `json:"product_id"`
			ProductName string `json:"product_name"`
		}

		options := make([]ProductOption, len(products))
		for i, p := range products {
			options[i] = ProductOption{
				ProductID:   p.ProductID,
				ProductName: p.ProductName,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(options); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// default: kirim full products untuk productspage
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// GetProductByID returns a product by id
func (s *Server) GetProductByID(w http.ResponseWriter, r *http.Request) {
	// parse numeric id from route and convert to int32 (sqlc expects int32)
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	id64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if id64 < math.MinInt32 || id64 > math.MaxInt32 {
		http.Error(w, "id out of range", http.StatusBadRequest)
		return
	}
	id := int32(id64)

	// call repo
	product, err := s.Repo.GetProductByID(r.Context(), id)
	if err != nil {
		// distinguish not found vs server error
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
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

// UpdateStockRequest accepts either a delta (relative change) or an absolute stock value.
type UpdateStockRequest struct {
    Delta *int32 `json:"delta,omitempty"`
    Stock *int32 `json:"stock,omitempty"`
}

// UpdateProductStock handles PATCH /products/{product_id}/stock
func (s *Server) UpdateProductStock(w http.ResponseWriter, r *http.Request) {
	// parse numeric id from route and convert to int32 (sqlc expects int32)
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	id64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if id64 < math.MinInt32 || id64 > math.MaxInt32 {
		http.Error(w, "id out of range", http.StatusBadRequest)
		return
	}
	id := int32(id64)

	// decode request
	var req UpdateStockRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Delta == nil && req.Stock == nil {
		http.Error(w, "either 'delta' or 'stock' must be provided", http.StatusBadRequest)
		return
	}

	// If delta provided, prefer atomic delta update in DB (prevents race)
	if req.Delta != nil {
		// Try repo method UpdateProductStockByDelta (generated by sqlc if you added it)
		updated, err := s.Repo.UpdateProductStockByDelta(r.Context(), repo.UpdateProductStockByDeltaParams{
			ID:    id,
			Stock: *req.Delta,
		})
		if err != nil {
			// if no rows returned, either product not found or condition failed (e.g. would go negative)
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "product not found or insufficient stock", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to update stock: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
		return
	}

	// If absolute stock provided, use UpdateProductStock (set stock)
	if req.Stock != nil {
		updated, err := s.Repo.UpdateProductStock(r.Context(), repo.UpdateProductStockParams{
			ID:    id,
			Stock: *req.Stock,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "product not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to update stock: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
		return
	}
}
