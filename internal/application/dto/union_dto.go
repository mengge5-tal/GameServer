package dto

import "time"

// CreateUnionRequest represents a request to create a union
type CreateUnionRequest struct {
	UnionName       string `json:"unionname"`
	ChairpersonID   int    `json:"chairpersonid"`
	ChairpersonName string `json:"chairpersonname"`
	UnionDesc       string `json:"union_desc,omitempty"`
}

// UpdateUnionRequest represents a request to update union information
type UpdateUnionRequest struct {
	UnionID   int    `json:"unionid"`
	UnionName string `json:"unionname,omitempty"`
	UnionDesc string `json:"union_desc,omitempty"`
}

// UnionResponse represents union information response
type UnionResponse struct {
	UnionID           int       `json:"unionid"`
	UnionName         string    `json:"unionname"`
	ChairpersonID     int       `json:"chairpersonid"`
	ChairpersonName   string    `json:"chairpersonname"`
	ChairpersonLevel  int       `json:"chairpersonlevel"`
	UnionLevel        int       `json:"unionlevel"`
	UnionMembers      int       `json:"unionmembers"`
	Experience        int       `json:"experience"`
	CreatedTime       time.Time `json:"created_time"`
	UnionDesc         string    `json:"union_desc"`
}

// UnionListResponse represents a list of unions for browsing
type UnionListResponse struct {
	Unions []UnionSummary `json:"unions"`
	Total  int            `json:"total"`
}

// UnionSummary represents simplified union information for lists
type UnionSummary struct {
	UnionID           int    `json:"unionid"`
	UnionName         string `json:"unionname"`
	ChairpersonName   string `json:"chairpersonname"`
	ChairpersonLevel  int    `json:"chairpersonlevel"`
	UnionLevel        int    `json:"unionlevel"`
	UnionMembers      int    `json:"unionmembers"`
	UnionDesc         string `json:"union_desc"`
}

// JoinUnionRequest represents a request to join a union
type JoinUnionRequest struct {
	UnionID        int `json:"unionid"`
	ApplicantID    int `json:"applicantid"`
}

// ProcessUnionRequestDTO represents a request to approve/reject union application
type ProcessUnionRequestDTO struct {
	RequestID     int `json:"request_id"`
	ChairpersonID int `json:"chairperson_id"`
	Status        int `json:"status"` // 2=通过, 3=拒绝
}

// UnionRequestResponse represents union request information
type UnionRequestResponse struct {
	ID             int       `json:"id"`
	UnionID        int       `json:"unionid"`
	UnionName      string    `json:"unionname"`
	ApplicantID    int       `json:"applicantid"`
	ApplicantName  string    `json:"applicantname"`
	ApplicantLevel int       `json:"applicantlevel"`
	RequestStatus  int       `json:"request_status"`
	StatusName     string    `json:"status_name"`
	RequestTime    time.Time `json:"request_time"`
}

// UnionRequestListResponse represents a list of union requests
type UnionRequestListResponse struct {
	Requests []UnionRequestResponse `json:"requests"`
	Total    int                    `json:"total"`
}

// LeaveUnionRequest represents a request to leave union
type LeaveUnionRequest struct {
	UnionID  int `json:"unionid"`
	MemberID int `json:"memberid"`
}

// KickMemberRequest represents a request to kick a member from union
type KickMemberRequest struct {
	UnionID       int `json:"unionid"`
	MemberID      int `json:"memberid"`
	ChairpersonID int `json:"chairpersonid"`
}

// PromoteMemberRequest represents a request to promote a member
type PromoteMemberRequest struct {
	UnionID       int `json:"unionid"`
	MemberID      int `json:"memberid"`
	NewRoleID     int `json:"new_role_id"`
	ChairpersonID int `json:"chairpersonid"`
}

// DismissUnionRequest represents a request to dismiss/dissolve a union
type DismissUnionRequest struct {
	UnionID       int `json:"unionid"`
	ChairpersonID int `json:"chairpersonid"`
}

// UnionRankResponse represents user's union ranking information
type UnionRankResponse struct {
	InUnion   bool   `json:"in_union"`
	UnionName string `json:"union_name,omitempty"`
	Rank      int    `json:"rank,omitempty"`
	Message   string `json:"message,omitempty"`
}