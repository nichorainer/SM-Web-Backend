package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
    "errors"
    "database/sql"
    
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	repo "github.com/yourorg/backend-go/internal/adapters/postgresql/sqlc"
	"github.com/yourorg/backend-go/internal/config"
	"github.com/yourorg/backend-go/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Standard response API
type APIResponse struct {
    Status  string      `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// CreateUserRequest is the expected JSON body for creating a user.
type CreateUserRequest struct {
    FullName string `json:"full_name"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

// UserResponse is the response struct for user data.
type UserResponse struct {
    ID       int32  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

// CreateUser creates a new user (Register)
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid request body"})
        return
    }

    req.Username = strings.TrimSpace(req.Username)
    req.Email = strings.TrimSpace(req.Email)
    req.FullName = strings.TrimSpace(req.FullName)

    if req.Username == "" || req.Email == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "username, email and password are required"})
        return
    }

    if _, err := mail.ParseAddress(req.Email); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid email format"})
        return
    }

    existing, err := s.Repo.GetUserByUsernameOrEmail(r.Context(),
        repo.GetUserByUsernameOrEmailParams{
            Username: req.Username, 
            Email: req.Email,
        })
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
        // Tidak ada user → aman lanjut create
    } else {
        log.Println("error checking existing user:", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "failed to check existing user",
        })
        return
    }
    } else {
        log.Println("error checking existing user:", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to check existing user"})
        return
    }
    if existing.ID != 0 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusConflict)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "username or email already registered"})
        return
    }

    uid := uuid.NewString()
    hashed, err := HashPassword(req.Password)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to hash password"})
        return
    }

    arg := repo.CreateUserParams{
        UserID:       uid,
        FullName:     req.FullName,
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: hashed,
        Role:         "admin",
    }

    u, err := s.Repo.CreateUser(r.Context(), arg)
    if err != nil {
        log.Println("failed to create user:", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to create user"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(APIResponse{Status: "success", Data: u, Message: "user registered"})
}

// LoginUser verifies email (or username) and password
type LoginRequest struct {
    UsernameOrEmail string `json:"username_or_email"`
    Password        string `json:"password"`
}

func (s *Server) LoginUser(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        log.Printf("Login decode error: %v", err)
        json.NewEncoder(w).Encode(APIResponse{
            Status:  "error",
            Message: "invalid request body: " + err.Error(),
        })
        return
    }

    // pilih email atau username
    params := repo.GetUserByUsernameOrEmailParams{
        Username: strings.TrimSpace(req.UsernameOrEmail),
        Email:    strings.TrimSpace(req.UsernameOrEmail),
    }

    user, err := s.Repo.GetUserByUsernameOrEmail(r.Context(), params)
    if err != nil || user.ID == 0 {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid email or password"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid email or password"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data: UserResponse{
            ID:       user.ID,
            Username: user.Username,
            Email:    user.Email,
        },
    })
}

// ListUsers returns all users.
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
    params := repo.ListUsersParams{Limit: 100, Offset: 0}
    users, err := s.Repo.ListUsers(r.Context(), params)
    if err != nil {
        log.Println("failed to list users:", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to list users"})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(APIResponse{Status: "success", Data: users})
}

// GetUserByID returns a user by ID.
func (s *Server) GetUserByID(w http.ResponseWriter, r *http.Request) {
    userIDStr := chi.URLParam(r, "id")
    id64, err := strconv.ParseInt(userIDStr, 10, 32)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid id"})
        return
    }
    userID := int32(id64)
    user, err := s.Repo.UserByID(r.Context(), userID)
    if err != nil {
        log.Println("user not found:", err)
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "user not found"})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(APIResponse{Status: "success", Data: user})
}

// GetProfile handler for GET /users/me
func (s *Server) GetProfile(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "id is required"})
        return
    }

    id64, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid id"})
        return
    }
    userID := int32(id64)

    u, err := s.Repo.UserByID(r.Context(), userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "user not found"})
            return
        }
        log.Println("failed to get profile:", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to get profile"})
        return
    }

    resp := models.User{
        ID:       int(u.ID),
        UserID:   u.UserID,
        FullName: u.FullName,
        Username: u.Username,
        Email:    u.Email,
        Role:     u.Role,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data:   resp,
    })
}


// UpdateUser handler
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    log.Println("UpdateUser handler triggered") // checking if the handler has started

    // Get user id (not user_id, id from table) from URL
    idStr := chi.URLParam(r, "id")
    if idStr == "" {
        log.Println("Missing id in URL")
        http.Error(w, "Missing id", http.StatusBadRequest)
        return
    }
    log.Printf("URL param id: %s", idStr)

    // Decode body JSON
    var input models.User
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        log.Printf("Invalid request body: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    log.Printf("Decoded input: %+v", input)

    // Plain password
    var plainPassword string
    if input.Password != "" {
        plainPassword = input.Password
    }

    db := config.GetDB()
    if db == nil {
        log.Println("DB pool is nil, did you call InitDB() in main.go?")
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        return
    }

    // Jalankan query update berdasarkan id
    _, err := db.Exec(
        r.Context(),
        `UPDATE users
         SET full_name     = COALESCE(NULLIF($2, ''), full_name),
             username      = COALESCE(NULLIF($3, ''), username),
             email         = COALESCE(NULLIF($4, ''), email),
             password_hash = COALESCE(NULLIF($5, ''), password_hash),
             updated_at    = NOW()
         WHERE id = $1`,
        idStr,
        input.FullName,
        input.Username,
        input.Email,
        plainPassword,
    )
    if err != nil {
        log.Printf("UpdateUser Exec error: %v", err)
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }

    // Ambil kembali user yang sudah diupdate
    row := db.QueryRow(
        r.Context(),
        `SELECT id, full_name, username, email, role
         FROM users WHERE id=$1`,
        idStr,
    )

    var updated models.User
    if err := row.Scan(&updated.ID, &updated.FullName, &updated.Username, &updated.Email, &updated.Role); err != nil {
        log.Printf("UpdateUser Scan error: %v", err)
        http.Error(w, "Failed to fetch updated user", http.StatusInternalServerError)
        return
    }
    log.Printf("Updated user: %+v", updated)

    // Return JSON lengkap
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "id":        updated.ID,
            "full_name": updated.FullName,
            "username":  updated.Username,
            "email":     updated.Email,
        },
    }); err != nil {
        log.Printf("UpdateUser Encode error: %v", err)
    }
}

// Utility
func HashPassword(pw string) (string, error) {
    b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(b), nil
}