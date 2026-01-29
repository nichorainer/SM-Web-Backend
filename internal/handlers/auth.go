package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/go-chi/chi/v5"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
	"github.com/yourorg/backend-go/internal/middleware"
	"github.com/yourorg/backend-go/internal/config"
	"github.com/yourorg/backend-go/internal/models"
)

// Request payloads
type AuthRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
}

// Login verifies credentials and returns a signed JWT string
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "invalid request body",
		})
		return
	}

	// Lookup user by username or email
	user, err := s.Repo.GetUserByUsernameOrEmail(r.Context(), repo.GetUserByUsernameOrEmailParams{
		Username: req.UsernameOrEmail,
		Email:    req.UsernameOrEmail,
	})
	if err != nil || user.ID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "invalid email, username or password",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "invalid email, username or password",
		})
		return
	}

	// Build claims (iat, exp)
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"user_id": user.UserID,
		"role":    user.Role, 
		"iat":     now.Unix(),
		"exp":     now.Add(24 * time.Hour).Unix(),
	}

	// Create token and sign it
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middleware.GetSecret())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "failed to generate token",
		})
		return
	}

	// Return token string + user info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"token":     tokenString,
			"user_id":   user.UserID,
			"username":  user.Username,
			"email":     user.Email,
			"full_name": user.FullName,
			"role":      user.Role,
		},
	})
}

// UpdateUser handles updating user profile
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    userId := chi.URLParam(r, "id")

    var input models.User
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    claims, err := middleware.ExtractClaims(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if userId == "me" {
        userId = claims["user_id"].(string)
    }

    db := config.GetDB()
    // Jalankan query update ke PostgreSQL
    _, err = db.Exec(
        r.Context(),
        `UPDATE users 
        SET full_name=$1, username=$2, email=$3, password=$4, role=$5, avatar_url=$6 
        WHERE id=$7`,
        input.FullName,
        input.Username,
        input.Email,
        input.Password, // pastikan sudah di-hash sebelumnya
        input.Role,
        input.AvatarUrl,
        userId,
    )

    if err != nil {
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }

    // Ambil kembali user yang sudah diupdate
    row := db.QueryRow(
        r.Context(),
        `SELECT id, full_name, username, email, role, avatar_url 
        FROM users WHERE id=$1`,
        userId,
    )

    var updated models.User
    if err := row.Scan(&updated.ID, &updated.FullName, &updated.Username, &updated.Email, &updated.Role, &updated.AvatarUrl); err != nil {
        http.Error(w, "Failed to fetch updated user", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data":   updated,
    })
}
