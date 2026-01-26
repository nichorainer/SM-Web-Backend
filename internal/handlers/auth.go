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

type RegisterRequest struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}


// Register creates a new user
func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	arg := repo.CreateUserParams{
		UserID:       req.UserID,
		Email:        req.Email,
		PasswordHash: string(hashed),
		Role:         req.Role,
	}

	// sqlc CreateUser returns (UserRow, error) — capture both (discard row if not needed)
	_, err = s.Repo.CreateUser(r.Context(), arg)
	if err != nil {
		http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

// Login verifies credentials and returns a signed JWT string
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Lookup user by username or email
	user, err := s.Repo.GetUserByUsernameOrEmail(r.Context(), repo.GetUserByUsernameOrEmailParams{
		Username: req.UsernameOrEmail,
		Email:    req.UsernameOrEmail,
	})
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Verify password (single check)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Build claims (iat, exp)
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"iat":     now.Unix(),
		"exp":     now.Add(24 * time.Hour).Unix(),
	}

	// Create token and sign it to get a string (no redeclaration issues)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return token string
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// JWTMiddleware verifies token and attaches user info to context
func JWTMiddleware(next http.Handler) http.Handler {
	return jwtauth.Authenticator(tokenAuth)(jwtauth.Verifier(tokenAuth)(next))
}