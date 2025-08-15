package dto

import "time"

// FriendResponse represents friend response data
type FriendResponse struct {
	ID         int       `json:"id"`
	FriendID   int       `json:"friend_id"`
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
	RequesterLevel    int    `json:"requester_level,omitempty"`
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

// FriendResponseRequest represents friend response request (accept/reject by user ID)
type FriendResponseRequest struct {
	FromUserID int  `json:"fromuserid"`
	Accept     bool `json:"accept"`
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