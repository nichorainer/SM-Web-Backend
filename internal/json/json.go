package json

import (
	"encoding/json"
	"net/http"
)

func Write(w http.ResponseWriter, status int, data any, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(data)
}

func Read(r *http.Request, data any, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	return decoder.Decode(data)
}
