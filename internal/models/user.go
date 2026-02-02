package models

type User struct {
    ID       int    `json:"id"`
    UserID   string `json:"user_id"`
    FullName string `json:"full_name"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    Password string `json:"password,omitempty"`
}