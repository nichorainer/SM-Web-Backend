package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"strconv"
	"log"

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
		"id": 	user.ID,
		"iat":	now.Unix(),
		"exp":	now.Add(24 * time.Hour).Unix(),
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
			"token":   		tokenString,
			"id": 			user.ID,
			"user_id": 		user.UserID,
			"username":		user.Username,
			"email":   		user.Email,
			"full_name":	user.FullName,
			"role":    		user.Role,
		},
	})
}

// UpdateUser handles updating user profile
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    log.Println("UpdateUser handler triggered") // checking if the handler has started

    // Take id from URL
    userIdStr := chi.URLParam(r, "id")
    log.Printf("URL param id: %s", userIdStr)

    idInt, err := strconv.Atoi(userIdStr)
    if err != nil {
        log.Printf("Invalid user id: %v", err)
        http.Error(w, "Invalid user id", http.StatusBadRequest)
        return
    }

    // Decode body
    var input models.User
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        log.Printf("Invalid request body: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    log.Printf("Decoded input: %+v", input)

    // Extract claims dari JWT
    // claims, err := middleware.ExtractClaims(r)
    // if err != nil {
    //     log.Printf("Unauthorized: %v", err)
    //     http.Error(w, "Unauthorized", http.StatusUnauthorized)
    //     return
    // }

    // Use /users/me
    // if userIdStr == "me" {
    //     idInt = int(claims["id"].(float64)) // JWT numeric claim dibaca float64
    // }

    // Hash password
    // var hashedPassword string
    // if input.Password != "" {
    //     hp, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    //     if err != nil {
    //         log.Printf("Failed to hash password: %v", err)
    //         http.Error(w, "Failed to hash password", http.StatusInternalServerError)
    //         return
    //     }
    //     hashedPassword = string(hp)
    // }

    // Plain password
    var plainPassword string
    if input.Password != "" {
        plainPassword = input.Password
    }
    log.Printf("Plain password length: %d", len(plainPassword))

    // Get DB
    db := config.GetDB()
    if db == nil {
        log.Println("DB pool is nil, did you call InitDB() in main.go?")
        http.Error(w, "Database not initialized", http.StatusInternalServerError)
        return
    }

    // Jalankan query update
    log.Printf("Executing UPDATE for id=%d, full_name=%q, username=%q, email=%q, password=%q",
        idInt, input.FullName, input.Username, input.Email, plainPassword)

    _, err = db.Exec(
        r.Context(),
        `UPDATE users
        SET full_name = COALESCE(NULLIF($2, ''), full_name),
            username   = COALESCE(NULLIF($3, ''), username),
            email      = COALESCE(NULLIF($4, ''), email),
            password_hash = COALESCE(NULLIF($5, ''), password_hash),
            updated_at = NOW()
        WHERE id = $1`,
        idInt,
        input.FullName,
        input.Username,
        input.Email,
        // hashedPassword,
        plainPassword,
    )
    if err != nil {
        log.Printf("UpdateUser Exec error: %v", err)
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }

    // Ambil kembali user yang sudah diupdate
    log.Printf("Fetching updated user id=%d", idInt)

    row := db.QueryRow(
        r.Context(),
        `SELECT id, full_name, username, email
         FROM users WHERE id=$1`,
        idInt,
    )

    var updated models.User
    if err := row.Scan(&updated.ID, &updated.FullName, &updated.Username, &updated.Email); err != nil {
        log.Printf("UpdateUser Scan error: %v", err)
        http.Error(w, "Failed to fetch updated user", http.StatusInternalServerError)
        return
    }
    log.Printf("Updated user: %+v", updated)

    // Return JSON hanya 3 field
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "full_name": updated.FullName,
            "username":  updated.Username,
            "email":     updated.Email,
        },
    }); err != nil {
        log.Printf("UpdateUser Encode error: %v", err)
    }
}
