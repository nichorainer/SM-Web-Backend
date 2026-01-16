package handlers

// import (
//     "encoding/json"
//     "net/http"
//     "strconv"

//     "github.com/go-chi/chi/v5"
    
// )

// func (s *Server) GetOrderByID(w http.ResponseWriter, r *http.Request) {
//     id := chi.URLParam(r, "id")
//     orderID, _ := strconv.Atoi(id)

//     order, err := s.Repo.GetOrderByID(r.Context(), int32(orderID))
//     if err != nil {
//         http.Error(w, err.Error(), http.StatusInternalServerError)
//         return
//     }

//     json.NewEncoder(w).Encode(order)
// }