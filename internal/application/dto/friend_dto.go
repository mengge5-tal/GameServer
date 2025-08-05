package dto

import "time"

// FriendResponse represents friend response data
type FriendResponse struct {
	ID         int       `json:"id"`
	FromUserID int       `json:"fromuserid"`
	ToUserID   int       `json:"touserid"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Additional friend info
	FriendUsername string `json:"friend_username,omitempty"`
	FriendLevel    int    `json:"friend_level,omitempty"`
}

// FriendRequestResponse represents friend request response data
type FriendRequestResponse struct {
	ID         int       `json:"id"`
	FromUserID int       `json:"fromuserid"`
	ToUserID   int       `json:"touserid"`
	Message    string    `json:"message"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Additional requester info
	RequesterUsername string `json:"requester_username,omitempty"`
}

// AddFriendRequest represents add friend request
type AddFriendRequest struct {
	ToUserID int    `json:"touserid"`
	Message  string `json:"message"`
}

// FriendActionRequest represents friend action request (accept/reject)
type FriendActionRequest struct {
	RequestID int `json:"request_id"`
}

// RemoveFriendRequest represents remove friend request
type RemoveFriendRequest struct {
	FriendUserID int `json:"friend_userid"`
}

// FriendRankResponse represents friend ranking response
type FriendRankResponse struct {
	UserID       int    `json:"userid"`
	Username     string `json:"username"`
	Level        int    `json:"level"`
	Experience   int    `json:"experience"`
	RankPosition int    `json:"rank_position"`
}