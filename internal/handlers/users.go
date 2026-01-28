package handlers

import (
    "encoding/json"
	"strconv"
	"log"
    "net/http"
	"net/mail"
	"strings"

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
	r.Post("/login", s.LoginUser)
}

// CreateUserRequest is the expected JSON body for creating a user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"` // e.g., "admin" or "staff"
}

// Standard response API
type APIResponse struct {
    Status  string      `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// ListUsers returns all users.
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
    params := repo.ListUsersParams{
        Limit:  100,
        Offset: 0,
    }

    users, err := s.Repo.ListUsers(r.Context(), params)
    if err != nil {
		// Error log
        log.Println("failed to list users:", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "failed to list users",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data:   users,
    })
}

// GetUserByID returns a user by ID.
func (s *Server) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// get param as string
    userIDStr := chi.URLParam(r, "user_id")
    if userIDStr == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "missing user_id",
        })
        return
    }
	
    // convert to int (match generated type)
    id64, err := strconv.ParseInt(userIDStr, 10, 32)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid user_id",
        })
        return
    }
    userID := int32(id64) // cast to int32 (change if generated type differs)

    user, err := s.Repo.UserByID(r.Context(), userID)
    if err != nil {
        log.Println("user not found:", err)
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "user not found",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data:   user,
    })
}

// CreateUser creates a new user using sqlc-generated params.
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
        log.Println("invalid request body:", err) // ✅ logging error
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid request body",
        })
        return
    }
	// Input sanitation to trim spaces in email and username
    req.Username = strings.TrimSpace(req.Username)
    req.Email = strings.TrimSpace(req.Email)
    req.FullName = strings.TrimSpace(req.FullName)
    req.Role = strings.TrimSpace(req.Role)

	// Basic validation
    if req.Username == "" || req.Email == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "username, email and password are required",
        })
        return
    }

	// Email validation format
	if _, err := mail.ParseAddress(req.Email); err != nil {
        log.Println("invalid email:", err)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid email format",
        })
        return
    }

	// Check email and username duplicate
    existing, err := s.Repo.GetUserByUsernameOrEmail(r.Context(), repo.GetUserByUsernameOrEmailParams{
		Username: req.Username,
		Email:    req.Email,
	})

	if err == nil && existing.ID != 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "error",
			Message: "username or email already registered",
		})
		return
	}

	// generate user id
	uid := uuid.NewString()

	// hash password
	hashed, err := hashPassword(req.Password)
    if err != nil {
        log.Println("failed to hash password:", err) // ✅ logging error
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "failed to hash password",
        })
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
        log.Println("failed to create user:", err) // ✅ logging error
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "failed to create user",
        })
        return
    }

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data:   u,
    })
}

// hashPassword hashes a plaintext password using bcrypt.
func hashPassword(pw string) (string, error) {
    b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

// LoginRequest is format JSON from frontend
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// LoginUser verifies email and password
func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        log.Println("invalid login request:", err)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid request body",
        })
        return
    }

    req.Email = strings.TrimSpace(req.Email)
    req.Password = strings.TrimSpace(req.Password)

    if req.Email == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "email and password are required",
        })
        return
    }

    // Get user by email
    user, err := s.Repo.GetUserByUsernameOrEmail(r.Context(), repo.GetUserByUsernameOrEmailParams{
        Username: "",
        Email:    req.Email,
    })
    if err != nil || user.ID == 0 {
        log.Println("user not found:", err)
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid email or password",
        })
        return
    }

    // Password verification
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        log.Println("invalid password:", err)
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid email or password",
        })
        return
    }

    // Succesfull login
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data: map[string]interface{}{
            "user_id":   user.UserID,
            "username":  user.Username,
            "email":     user.Email,
            "full_name": user.FullName,
            "role":      user.Role,
        },
    })
}