package repository

import "GameServer/internal/domain/entity"

// UnionRepository defines the interface for union data access
type UnionRepository interface {
	// Union operations
	GetByID(unionID int) (*entity.Union, error)
	GetByName(unionName string) (*entity.Union, error)
	Create(union *entity.Union) error
	Update(union *entity.Union) error
	Delete(unionID int) error
	Exists(unionName string) (bool, error)
	
	// Union listing and search
	GetAll(limit, offset int) ([]*entity.Union, int, error) // returns unions, total, error
	GetRecommended(limit int) ([]*entity.Union, error)
	Search(keyword string, limit, offset int) ([]*entity.Union, int, error)
	
	// Union ranking
	GetRanking(limit int) ([]*entity.Union, error)
	GetUnionRank(unionID int) (int, error)
	
	// Union experience and level
	UpdateExperience(unionID, experience int) error
	UpdateLevel(unionID, level int) error
	IncrementMemberCount(unionID int) error
	DecrementMemberCount(unionID int) error
	
	// Union information updates
	UpdateUnionInfo(unionID int, unionName, description string) error
}

// UnionMemberRepository defines the interface for union member data access
type UnionMemberRepository interface {
	// Member operations
	GetByID(id int) (*entity.UnionMember, error)
	GetByUserID(userID int) (*entity.UnionMember, error)
	GetByUnionID(unionID int) ([]*entity.UnionMember, error)
	Create(member *entity.UnionMember) error
	Update(member *entity.UnionMember) error
	Delete(id int) error
	DeleteByUserID(userID int) error
	
	// Member validation
	IsUserInUnion(userID int) (bool, error)
	IsUserInSpecificUnion(userID, unionID int) (bool, error)
	GetMemberRole(userID, unionID int) (int, error)
	
	// Member management
	UpdateRole(userID, unionID, roleID int) error
	GetMemberCount(unionID int) (int, error)
	GetMembersByUnionIDWithPagination(unionID, page, limit int) ([]*entity.UnionMember, int, error)
	SearchMembersByUnionIDAndKeyword(unionID int, keyword string, page, limit int) ([]*entity.UnionMember, int, error)
}

// UnionRequestRepository defines the interface for union request data access
type UnionRequestRepository interface {
	// Request operations
	GetByID(id int) (*entity.UnionRequest, error)
	GetByUnionID(unionID int) ([]*entity.UnionRequest, error)
	GetByApplicantID(applicantID int) ([]*entity.UnionRequest, error)
	GetPendingByUnionID(unionID int) ([]*entity.UnionRequest, error)
	GetByUnionIDAndStatus(unionID, status int) ([]*entity.UnionRequest, error)
	Create(request *entity.UnionRequest) error
	Update(request *entity.UnionRequest) error
	Delete(id int) error
	
	// Request validation
	HasPendingRequest(applicantID, unionID int) (bool, error)
	
	// Request processing
	ProcessRequest(requestID, status int) error
}

// UnionExperienceRepository defines the interface for union experience data access
type UnionExperienceRepository interface {
	GetByLevel(level int) (*entity.UnionExperience, error)
	GetAllLevels() ([]*entity.UnionExperience, error)
	GetNextLevel(currentLevel int) (*entity.UnionExperience, error)
}

// UnionInviteRepository defines the interface for union invite data access
type UnionInviteRepository interface {
	// Invite operations
	GetByID(id int) (*entity.UnionInvite, error)
	GetByUserID(userID int) ([]*entity.UnionInvite, error)
	GetByUnionID(unionID int) ([]*entity.UnionInvite, error)
	GetPendingByUserID(userID int) ([]*entity.UnionInvite, error)
	Create(invite *entity.UnionInvite) error
	Update(invite *entity.UnionInvite) error
	Delete(id int) error
	
	// Invite validation
	HasPendingInvite(fromUserName, toUserName string, unionID int) (bool, error)
	
	// Invite processing
	ProcessInvite(inviteID int, status string) error
}