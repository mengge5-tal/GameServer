package dto

// UserSourceStoneResponse represents user source stone ownership response
type UserSourceStoneResponse struct {
	ID               int `json:"id"`
	UserID           int `json:"userid"`
	SourceStoneID    int `json:"sourcestoneid"`
	SourceStoneCount int `json:"sourcestonecount"`
}

// UserSourceStoneDetailResponse represents user source stone ownership with source stone details
type UserSourceStoneDetailResponse struct {
	ID               int                    `json:"id"`
	UserID           int                    `json:"userid"`
	SourceStoneID    int                    `json:"sourcestoneid"`
	SourceStoneCount int                    `json:"sourcestonecount"`
	SourceStone      *SourceStoneResponse   `json:"sourcestone,omitempty"`
}

// AddUserSourceStoneRequest represents add source stone to user request
type AddUserSourceStoneRequest struct {
	UserID           int `json:"userid"`
	SourceStoneID    int `json:"sourcestoneid"`
	SourceStoneCount int `json:"sourcestonecount"`
}

// UpdateUserSourceStoneRequest represents update user source stone request
type UpdateUserSourceStoneRequest struct {
	SourceStoneID    int `json:"sourcestoneid"`
	SourceStoneCount int `json:"sourcestonecount"`
}

// RemoveUserSourceStoneRequest represents remove source stone from user request
type RemoveUserSourceStoneRequest struct {
	UserID        int `json:"userid"`
	SourceStoneID int `json:"sourcestoneid"`
}

// RemoveUserSourceStoneByIDRequest represents remove user source stone by ID request
type RemoveUserSourceStoneByIDRequest struct {
	ID int `json:"id"`
}

// GetUserSourceStonesRequest represents get user source stones request
type GetUserSourceStonesRequest struct {
	UserID      int  `json:"userid"`
	WithDetails bool `json:"with_details,omitempty"`
}

// CheckUserSourceStoneRequest represents check if user owns source stone request
type CheckUserSourceStoneRequest struct {
	UserID        int `json:"userid"`
	SourceStoneID int `json:"sourcestoneid"`
}

// CheckUserSourceStoneResponse represents check user source stone response
type CheckUserSourceStoneResponse struct {
	UserID           int  `json:"userid"`
	SourceStoneID    int  `json:"sourcestoneid"`
	OwnsSourceStone  bool `json:"owns_sourcestone"`
	SourceStoneCount int  `json:"sourcestonecount,omitempty"`
}