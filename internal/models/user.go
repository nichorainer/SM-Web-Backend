package models

type User struct {
    ID        string `json:"id"`
    FullName  string `json:"full_name"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Password  string `json:"password,omitempty"`
    Role      string `json:"role"`
    AvatarUrl string `json:"avatarUrl"`
}