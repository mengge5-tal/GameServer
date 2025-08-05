package dto

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

// LoginResponse represents login response data
type LoginResponse struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Token    string `json:"token,omitempty"` // Optional token for future use
}

// RegisterResponse represents registration response data
type RegisterResponse struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// UserProfile represents user profile data
type UserProfile struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Experience int  `json:"experience"`
	GameLevel int   `json:"gamelevel"`
	BloodEnergy int `json:"bloodenergy"`
}