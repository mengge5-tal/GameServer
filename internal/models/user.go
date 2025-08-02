package models

// LoginRequest represents login request data
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents registration request data
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User represents user information
type User struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
}