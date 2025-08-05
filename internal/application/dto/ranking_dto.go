package dto

import "time"

// RankingResponse represents ranking response data
type RankingResponse struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userid"`
	Username     string    `json:"username,omitempty"` // Will be populated by service
	RankType     string    `json:"rank_type"`
	RankValue    int       `json:"rank_value"`
	RankPosition int       `json:"rank_position"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetRankingRequest represents get ranking request
type GetRankingRequest struct {
	RankType string `json:"rank_type"` // level, experience, equipment_power
	Limit    int    `json:"limit"`     // Number of top ranks to return
}

// UserRankingResponse represents user's specific ranking
type UserRankingResponse struct {
	UserID       int       `json:"userid"`
	Username     string    `json:"username"`
	RankType     string    `json:"rank_type"`
	RankValue    int       `json:"rank_value"`
	RankPosition int       `json:"rank_position"`
	UpdatedAt    time.Time `json:"updated_at"`
}