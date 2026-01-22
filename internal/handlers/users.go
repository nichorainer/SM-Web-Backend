package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
    "github.com/go-chi/chi/v5"
    repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
)

// RegisterUserRoutes registers user-related routes on the provided router.
func (s *Server) RegisterUserRoutes(r chi.Router) {
	r.Get("/users", s.ListUsers)
	r.Get("/users/{user_id}", s.GetUserByID)
	r.Post("/users", s.CreateUser)
}

// CreateUserRequest is the expected JSON body for creating a user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"` // e.g., "admin" or "staff"
}

// ListUsers returns all users.
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
    params := repo.ListUsersParams{
        Limit:  100, // or parse from query string
        Offset: 0,   // or parse from query string
    }

    users, err := s.Repo.ListUsers(r.Context(), params)
    if err != nil {
        http.Error(w, "failed to list users: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(users); err != nil {
        http.Error(w, "failed to encode response", http.StatusInternalServerError)
    }
}

// GetUserByID returns a user by ID.
func (s *Server) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// get param as string
    userIDStr := chi.URLParam(r, "user_id")
    if userIDStr == "" {
        http.Error(w, "missing user_id", http.StatusBadRequest)
        return
    }

    // convert to int (match generated type)
    id64, err := strconv.ParseInt(userIDStr, 10, 32)
    if err != nil {
        http.Error(w, "invalid user_id", http.StatusBadRequest)
        return
    }
    userID := int32(id64) // cast to int32 (change if generated type differs)

    user, err := s.Repo.GetUserByID(r.Context(), userID)
    if err != nil {
        http.Error(w, "user not found: "+err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(user); err != nil {
        http.Error(w, "failed to encode response", http.StatusInternalServerError)
    }
}

// CreateUser creates a new user using sqlc-generated params.
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// basic validation
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "username, email and password are required", http.StatusBadRequest)
		return
	}

	// generate user id
	uid := uuid.NewString()

	// hash password
	hashed, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	arg := repo.CreateUserParams{
		UserID:       uid,
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: hashed,
		Role:         req.Role,
	}

	u, err := s.Repo.CreateUser(r.Context(), arg)
	if err != nil {
		http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(u); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// hashPassword hashes a plaintext password using bcrypt.
func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
