package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
)

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