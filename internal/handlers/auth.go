package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"

	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

// secret (move to env/config in production)
var jwtSecret = []byte("your-secret-key")

// jwtauth instance for middleware verification (uses same secret)
var tokenAuth = jwtauth.New("HS256", jwtSecret, nil)

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
	tokenString, err := token.SignedString(jwtSecret)
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

// JWTMiddleware verifies token and attaches user info to context
func JWTMiddleware(next http.Handler) http.Handler {
	return jwtauth.Authenticator(tokenAuth)(jwtauth.Verifier(tokenAuth)(next))
}