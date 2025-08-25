package dto

import "time"

// UnionMemberResponse represents union member information
type UnionMemberResponse struct {
	ID          int       `json:"id"`
	UnionID     int       `json:"unionid"`
	UnionName   string    `json:"unionname"`
	MemberID    int       `json:"memberid"`
	MemberName  string    `json:"membername,omitempty"` // 可能需要从用户表获取
	MemberLevel int       `json:"memberlevel"`
	JoinedTime  time.Time `json:"joined_time"`
	RoleID      int       `json:"roleid"`
	RoleName    string    `json:"rolename"`
}

// UnionMemberListResponse represents a list of union members
type UnionMemberListResponse struct {
	Members []UnionMemberResponse `json:"members"`
	Total   int                   `json:"total"`
}

// UnionMemberSummary represents simplified member information
type UnionMemberSummary struct {
	MemberID    int    `json:"memberid"`
	MemberName  string `json:"membername"`
	MemberLevel int    `json:"memberlevel"`
	RoleID      int    `json:"roleid"`
	RoleName    string `json:"rolename"`
}

// GetUnionMembersRequest represents a request to get union members
type GetUnionMembersRequest struct {
	UnionID int `json:"unionid"`
	Page    int `json:"page,omitempty"`
	Limit   int `json:"limit,omitempty"`
}

// UpdateMemberRoleRequest represents a request to update member role
type UpdateMemberRoleRequest struct {
	UnionID       int `json:"unionid"`
	MemberID      int `json:"memberid"`
	NewRoleID     int `json:"new_role_id"`
	OperatorID    int `json:"operator_id"` // 执行操作的用户ID
}