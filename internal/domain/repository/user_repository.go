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
	
	// Ranking methods
	GetPlayerRanking(rankType string, limit int) ([]*entity.PlayerRankingEntry, error)
	GetUserRank(userID int, rankType string) (*entity.PlayerRankingEntry, error)
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
	GetRecommendedFriends(userID, userLevel, levelDiff, limit int, onlineOnly bool) ([]*entity.UserWithLevel, error)
}


// EquipmentRepository defines the interface for equipment data access
type EquipmentRepository interface {
	GetByUserID(userID int) ([]*entity.Equipment, error)
	GetByEquipID(equipID int) (*entity.Equipment, error)
	Create(equipment *entity.Equipment) error
	Update(equipment *entity.Equipment) error
	Delete(equipID int) error
	GetUserEquipmentCount(userID int) (int, error)
}

// SourceStoneRepository defines the interface for source stone data access
type SourceStoneRepository interface {
	GetByID(sourcestoneID int) (*entity.SourceStone, error)
	GetAll() ([]*entity.SourceStone, error)
}

// ExperienceRepository defines the interface for experience data access
type ExperienceRepository interface {
	GetByLevel(level int) (*entity.Experience, error)
	GetAllLevels() ([]*entity.Experience, error)
}

// WeaponRepository defines the interface for weapon data access
type WeaponRepository interface {
	GetByID(weaponID int) (*entity.Weapon, error)
	GetAll() ([]*entity.Weapon, error)
	Create(weapon *entity.Weapon) error
	Update(weapon *entity.Weapon) error
	Delete(weaponID int) error
}

// UserWeaponRepository defines the interface for user weapon ownership data access
type UserWeaponRepository interface {
	GetByID(id int) (*entity.UserWeapon, error)
	GetByUserID(userID int) ([]*entity.UserWeapon, error)
	GetByUserAndWeapon(userID, weaponID int) (*entity.UserWeapon, error)
	Create(userWeapon *entity.UserWeapon) error
	Delete(id int) error
	DeleteByUserAndWeapon(userID, weaponID int) error
	UserOwnsWeapon(userID, weaponID int) (bool, error)
}

// UserSourceStoneRepository defines the interface for user source stone ownership data access
type UserSourceStoneRepository interface {
	GetByID(id int) (*entity.UserSourceStone, error)
	GetByUserID(userID int) ([]*entity.UserSourceStone, error)
	GetByUserAndSourceStone(userID, sourcestoneID int) (*entity.UserSourceStone, error)
	Create(userSourceStone *entity.UserSourceStone) error
	Update(userSourceStone *entity.UserSourceStone) error
	Delete(id int) error
	DeleteByUserAndSourceStone(userID, sourcestoneID int) error
	UserOwnsSourceStone(userID, sourcestoneID int) (bool, error)
}

// KillCountRepository defines the interface for kill count data access
type KillCountRepository interface {
	GetByUserIDAndDate(userID int, date string) (*entity.KillCount, error)
	Create(killCount *entity.KillCount) error
	Update(killCount *entity.KillCount) error
	Delete(id int) error
	IncrementKill(userID int, date string, monsterType string, count int) error
	ResetAllToday() error
	GetTodayKillCount(userID int) (*entity.KillCount, error)
	
	// Ranking methods
	GetKillRanking(limit int) ([]*entity.KillRankingEntry, error)
	GetUserKillRank(userID int) (*entity.KillRankingEntry, error)
}