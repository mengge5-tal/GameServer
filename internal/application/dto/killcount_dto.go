package dto

// KillCountResponse represents kill count response data
type KillCountResponse struct {
	ID     int    `json:"id"`
	UserID int    `json:"userid"`
	Today  string `json:"today"`
	Normal int    `json:"normal"`
	Elite  int    `json:"elite"`
	Boss   int    `json:"boss"`
	Count  int    `json:"count"` // Total kill count
}

// GetKillCountRequest represents get kill count request
type GetKillCountRequest struct {
	UserID int    `json:"userid,omitempty"` // Optional, defaults to current user
	Date   string `json:"date,omitempty"`   // Optional, defaults to today
}

// UpdateKillCountRequest represents update kill count request
type UpdateKillCountRequest struct {
	Normal int `json:"normal"`
	Elite  int `json:"elite"`
	Boss   int `json:"boss"`
}

// IncrementKillCountRequest represents increment kill count request
type IncrementKillCountRequest struct {
	MonsterType string `json:"monster_type"` // "normal", "elite", "boss"
	Count       int    `json:"count"`        // Number to increment, defaults to 1
}

// BatchIncrementKillCountRequest represents batch increment kill count request
type BatchIncrementKillCountRequest struct {
	Normal *int `json:"normal,omitempty"` // Number to increment for normal monsters (optional)
	Elite  *int `json:"elite,omitempty"`  // Number to increment for elite monsters (optional)
	Boss   *int `json:"boss,omitempty"`   // Number to increment for boss monsters (optional)
}

// KillRankingResponse represents kill ranking response data
type KillRankingResponse struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Count    int    `json:"count"` // Total kill count
	Rank     int    `json:"rank"`  // Ranking position
}

// GetKillRankingRequest represents get kill ranking request
type GetKillRankingRequest struct {
	Limit int `json:"limit,omitempty"` // Number of top entries to return, defaults to 100
}

// GetUserKillRankRequest represents get user kill rank request  
type GetUserKillRankRequest struct {
	UserID int `json:"userid,omitempty"` // Optional, defaults to current user
}

// UserKillRankResponse represents user kill rank response
type UserKillRankResponse struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Count    int    `json:"count"` // Total kill count
	Rank     int    `json:"rank"`  // Current ranking position
}