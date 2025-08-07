package repository

import "GameServer/internal/domain/entity"

// UserRepository defines the interface for user data access
type UserRepository interface {
	// User operations
	GetByID(id int) (*entity.User, error)
	GetByUsername(username string) (*entity.User, error)
	Create(user *entity.User) error
	Update(user *entity.User) error
	Delete(id int) error
	Exists(username string) (bool, error)

	// Authentication
	VerifyCredentials(username, password string) (*entity.User, error)
	
	// Online status
	UpdateOnlineStatus(userID int, status int) error
}

// PlayerRepository defines the interface for player information data access
type PlayerRepository interface {
	GetByUserID(userID int) (*entity.PlayerInfo, error)
	Create(player *entity.PlayerInfo) error
	Update(player *entity.PlayerInfo) error
	Delete(userID int) error
	UpdateExperience(userID, experience int) error
	UpdateLevel(userID, level int) error
	UpdateBloodEnergy(userID, bloodEnergy int) error
}

// FriendRepository defines the interface for friend data access
type FriendRepository interface {
	GetFriendsByUserID(userID int) ([]*entity.Friend, error)
	GetFriendRequestsByUserID(userID int) ([]*entity.FriendRequest, error)
	CreateFriendRequest(request *entity.FriendRequest) error
	AcceptFriendRequest(requestID int) error
	RejectFriendRequest(requestID int) error
	RemoveFriend(fromUserID, toUserID int) error
	AreFriends(userID1, userID2 int) (bool, error)
	HasPendingRequest(fromUserID, toUserID int) (bool, error)
}

// RankingRepository defines the interface for ranking data access
type RankingRepository interface {
	GetRankingByType(rankType string, limit int) ([]*entity.Ranking, error)
	UpdateUserRanking(userID int, rankType string, value int) error
	GetUserRanking(userID int, rankType string) (*entity.Ranking, error)
	RefreshRankings(rankType string) error
}

// EquipmentRepository defines the interface for equipment data access
type EquipmentRepository interface {
	GetByUserID(userID int) ([]*entity.Equipment, error)
	GetByEquipID(equipID int) (*entity.Equipment, error)
	Create(equipment *entity.Equipment) error
	Update(equipment *entity.Equipment) error
	Delete(equipID int) error
	GetUserEquipmentCount(userID int) (int, error)
	GetMaxSequenceByTypeAndQuality(equipType, quality int) (int, error)
}

// SourceStoneRepository defines the interface for source stone data access
type SourceStoneRepository interface {
	GetByUserID(userID int) ([]*entity.SourceStone, error)
	GetByEquipID(equipID int) (*entity.SourceStone, error)
	Create(sourceStone *entity.SourceStone) error
	Update(sourceStone *entity.SourceStone) error
	Delete(equipID int) error
	UpdateCount(equipID, count int) error
}

// ExperienceRepository defines the interface for experience data access
type ExperienceRepository interface {
	GetByLevel(level int) (*entity.Experience, error)
	GetAllLevels() ([]*entity.Experience, error)
}