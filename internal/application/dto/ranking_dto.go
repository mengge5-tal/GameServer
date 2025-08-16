package dto

// PlayerRankingRequest represents a request for player ranking
type PlayerRankingRequest struct {
	RankType string `json:"rank_type"` // level, experience, gamelevel, bloodenergy
	Limit    int    `json:"limit"`     // number of results to return
}

// GetUserRankRequest represents a request for getting a user's rank
type GetUserRankRequest struct {
	UserID   int    `json:"userid,omitempty"` // optional, will use client's user ID if not provided
	RankType string `json:"rank_type"`        // level, experience, gamelevel, bloodenergy
}

// PlayerRankingResponse represents a player's ranking information
type PlayerRankingResponse struct {
	UserID   int    `json:"userid"`
	Username string `json:"username,omitempty"`
	Value    int    `json:"value"`
	Position int    `json:"position"`
}

// GetPlayerRankingResponse represents the response for player ranking query
type GetPlayerRankingResponse struct {
	RankType string                   `json:"rank_type"`
	Rankings []*PlayerRankingResponse `json:"rankings"`
}

// UserRankResponse represents a single user's rank in a specific ranking
type UserRankResponse struct {
	UserID   int    `json:"userid"`
	Username string `json:"username,omitempty"`
	RankType string `json:"rank_type"`
	Value    int    `json:"value"`
	Position int    `json:"position"`
}