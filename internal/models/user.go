package models

type User struct {
    ID          int               `json:"id" db:"id"`
    UserID      string            `json:"user_id" db:"user_id"`
    FullName    string            `json:"full_name" db:"full_name"`
    Username    string            `json:"username" db:"username"`
    Email       string            `json:"email" db:"email"`
    Role        string            `json:"role" db:"role"`
    Password    string            `json:"password,omitempty" db:"password"`
    CreatedAt   string            `json:"created_at" db:"created_at"`
    UpdatedAt   string            `json:"updated_at" db:"updated_at"`
    Permissions    map[string]bool `json:"permissions"`
}