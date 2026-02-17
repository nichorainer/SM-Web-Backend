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
    "golang.org/x/crypto/bcrypt"
    
	repo "github.com/nichorainer/backend-go/internal/adapters/postgresql/sqlc"
	"github.com/nichorainer/backend-go/internal/config"
	"github.com/nichorainer/backend-go/internal/models"
)

// Standard response API
type APIResponse struct {
    Status  string      `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// CreateUserRequest is the expected JSON body for creating a user.
type CreateUserRequest struct {
    FullName    string         `json:"full_name"`
    Username    string         `json:"username"`
    Email       string         `json:"email"`
    Password    string         `json:"password"`
    Permissions map[string]bool `json:"permissions,omitempty"`
}

// UserResponse is the response struct for user data.
type UserResponse struct {
    ID          int32  `json:"id"`
    Username    string `json:"username"`
    Email       string `json:"email"`
    Permissions map[string]bool `json:"permissions"`
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

    if req.Permissions == nil {
        req.Permissions = map[string]bool{
            "orders":   true,
            "products": true,
            "users":    false,
        }
    }

    permBytes, err := json.Marshal(req.Permissions)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to process permissions"})
        return
    }

    arg := repo.CreateUserParams{
        UserID:       uid,
        FullName:     req.FullName,
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: hashed,
        Role:         "staff",
        Permissions: permBytes,
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

    var rawPerms []byte
    err = s.DB.QueryRow(r.Context(),
        "SELECT permissions FROM users WHERE id = $1", user.ID).Scan(&rawPerms)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to load permissions"})
        return
    }

    var perms map[string]bool
    if err := json.Unmarshal(rawPerms, &perms); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "invalid permissions format"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(APIResponse{
        Status: "success",
        Data: UserResponse{
            ID:          user.ID,
            Username:    user.Username,
            Email:       user.Email,
            Permissions: perms,
        },
    })
}


// List all users
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
    params := repo.ListUsersParams{Limit: 100, Offset: 0}
    rows, err := s.Repo.ListUsers(r.Context(), params)
    if err != nil {
        log.Println("failed to list users:", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "failed to list users"})
        return
    }

    var users []models.User
    for _, r := range rows {
        permMap := make(map[string]bool)
        if err := json.Unmarshal(r.Permissions, &permMap); err != nil {
            log.Printf("failed to unmarshal permissions for user %d: %v", r.ID, err)
            // default ke empty permissions kalau error
            permMap = make(map[string]bool)
        }

        users = append(users, models.User{
            ID:          int(r.ID),
            UserID:      r.UserID,
            Username:    r.Username,
            Email:       r.Email,
            FullName:    r.FullName,
            Role:        r.Role,
            CreatedAt:   r.CreatedAt.Time.String(),
            UpdatedAt:   r.UpdatedAt.Time.String(),
            Permissions: permMap,
        })
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

    // Again, hashed password
    var hashedPassword string
    if input.Password != "" {
        // hash password dengan bcrypt
        hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
        if err != nil {
            log.Println("failed to hash password:", err)
            http.Error(w, "Failed to process password", http.StatusInternalServerError)
            return
        }
        hashedPassword = string(hash)
    }

    db := config.GetDB()
    if db == nil {
        log.Println("DB pool is nil, did you call InitDB() in main.go?")
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        return
    }

    // Jalankan query update berdasarkan id
    if hashedPassword != "" {
        _, err := db.Exec(
            r.Context(),
            `UPDATE users
            SET full_name  = COALESCE(NULLIF($2, ''), full_name),
                username   = COALESCE(NULLIF($3, ''), username),
                email      = COALESCE(NULLIF($4, ''), email),
                password_hash = $5,
                updated_at = NOW()
            WHERE id = $1`,
            idStr,
            input.FullName,
            input.Username,
            input.Email,
            hashedPassword,
        )
        if err != nil {
            log.Printf("UpdateUser Exec error: %v", err)
            http.Error(w, "Failed to update user", http.StatusInternalServerError)
            return
        }

    } else {
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
            hashedPassword,
        )
        
        if err != nil {
            log.Printf("UpdateUser Exec error: %v", err)
            http.Error(w, "Failed to update user", http.StatusInternalServerError)
            return
        }
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
            "role":      updated.Role,
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

// UpdatePermissions handler for PUT /users/permissions
func (s *Server) UpdatePermissions(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req struct {
        UserID      int32              `json:"user_id"`
        Permissions map[string]bool    `json:"permissions"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    db := config.GetDB()
    if db == nil {
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        return
    }

    // Convert map ke JSON string dulu untuk JSONB column
    permsJSON, err := json.Marshal(req.Permissions)
    if err != nil {
        log.Printf("failed to marshal permissions: %v", err)
        http.Error(w, "failed to process permissions", http.StatusInternalServerError)
        return
    }

    _, err = db.Exec(
        ctx,
        `UPDATE users
         SET permissions = $2,
             updated_at = NOW()
         WHERE id = $1`,
        req.UserID,
        permsJSON,
    )
    if err != nil {
        log.Printf("failed to update permissions: %v", err)
        http.Error(w, "failed to update permissions", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "permissions updated",
    })
}

// Handler for update role
func (s *Server) UpdateUserRole (w http.ResponseWriter, r *http.Request) {
    var req struct {
        ID int    `json:"id"`
        Role   string `json:"role"`
    }

    // decode JSON dari body
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // validasi role
    if req.Role != "staff" && req.Role != "admin" {
        http.Error(w, "Invalid role value", http.StatusBadRequest)
        return
    }

    db := config.GetDB()
    if db == nil {
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        return
    }

    // update DB
    _, err := db.Exec(r.Context(), `UPDATE users SET role = $2 WHERE id = $1`, req.ID, req.Role)
    if err != nil {
        http.Error(w, "Failed to update role", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}