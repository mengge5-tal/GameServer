package websocket

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
)

// AuthServiceInterface defines the interface for auth service used by websocket handlers
type AuthServiceInterface interface {
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	GetUserProfile(userID int) (*dto.UserProfile, error)
	Logout(userID int) error
}

// PlayerServiceInterface defines the interface for player service used by websocket handlers
type PlayerServiceInterface interface {
	GetPlayerInfo(userID int) (*dto.PlayerInfoResponse, error)
	UpdatePlayer(req *dto.UpdatePlayerRequest) error
	GetUserEquipment(userID int) ([]*dto.EquipmentResponse, error)
	SaveEquipment(req *dto.SaveEquipmentRequest) (*dto.EquipmentResponse, error)
	DeleteEquipment(equipID, userID int) error
}

// FriendServiceInterface defines the interface for friend service used by websocket handlers
type FriendServiceInterface interface {
	GetFriends(userID int) ([]*dto.FriendResponse, error)
	GetFriendRequests(userID int) ([]*dto.FriendRequestResponse, error)
	SendFriendRequest(fromUserID int, req *dto.AddFriendRequest) error
	AcceptFriendRequest(userID int, req *dto.FriendActionRequest) error
	RejectFriendRequest(userID int, req *dto.FriendActionRequest) error
	RemoveFriend(userID int, req *dto.RemoveFriendRequest) error
	GetFriendRanking(userID int) ([]*dto.FriendRankResponse, error)
	AcceptFriendRequestByUserID(toUserID, fromUserID int) error
	RejectFriendRequestByUserID(toUserID, fromUserID int) error
	GetRecommendedFriends(userID int) ([]*dto.RecommendedFriendResponse, error)
}

// RankingServiceInterface defines the interface for ranking service used by websocket handlers
type RankingServiceInterface interface {
	GetRanking(req *dto.GetRankingRequest) ([]*dto.RankingResponse, error)
	GetUserRanking(userID int, rankType string) (*dto.UserRankingResponse, error)
	UpdateUserRankings(userID int) error
	RefreshAllRankings() error
}

// UserEquipServiceInterface defines the interface for user equipment service used by websocket handlers
type UserEquipServiceInterface interface {
	GetUserEquippedItems(userID int) (map[string]interface{}, error)
	EquipItem(userID int, slot string, equipID int) error
	UnequipItem(userID int, slot string) error
	GetEquippedItemsBySlot(userID int, slot string) (interface{}, error)
	GetEquipmentStats(userID int) (map[string]int, error)
}

// ExperienceServiceInterface defines the interface for experience service used by websocket handlers
type ExperienceServiceInterface interface {
	GetByLevel(level int) (*entity.Experience, error)
	GetAllLevels() ([]*entity.Experience, error)
}

// WeaponServiceInterface defines the interface for weapon service used by websocket handlers
type WeaponServiceInterface interface {
	GetWeaponByID(weaponID int) (*dto.WeaponResponse, error)
	GetAllWeapons() ([]*dto.WeaponResponse, error)
	CreateWeapon(req *dto.CreateWeaponRequest) (*dto.WeaponResponse, error)
	UpdateWeapon(req *dto.UpdateWeaponRequest) (*dto.WeaponResponse, error)
	DeleteWeapon(weaponID int) error
}

// UserWeaponServiceInterface defines the interface for user weapon service used by websocket handlers
type UserWeaponServiceInterface interface {
	GetUserWeapons(userID int, withDetails bool) (interface{}, error)
	AddUserWeapon(req *dto.AddUserWeaponRequest) (*dto.UserWeaponResponse, error)
	RemoveUserWeapon(req *dto.RemoveUserWeaponRequest) error
	RemoveUserWeaponByID(id int) error
	CheckUserWeapon(req *dto.CheckUserWeaponRequest) (*dto.CheckUserWeaponResponse, error)
}

// SourceStoneServiceInterface defines the interface for source stone service used by websocket handlers
type SourceStoneServiceInterface interface {
	GetSourceStoneByID(sourcestoneID int) (*dto.SourceStoneResponse, error)
	GetAllSourceStones() ([]*dto.SourceStoneResponse, error)
}

// UserSourceStoneServiceInterface defines the interface for user source stone service used by websocket handlers
type UserSourceStoneServiceInterface interface {
	GetUserSourceStones(userID int, withDetails bool) (interface{}, error)
	AddUserSourceStone(req *dto.AddUserSourceStoneRequest) (*dto.UserSourceStoneResponse, error)
	UpdateUserSourceStone(req *dto.UpdateUserSourceStoneRequest) (*dto.UserSourceStoneResponse, error)
	RemoveUserSourceStone(req *dto.RemoveUserSourceStoneRequest) error
	RemoveUserSourceStoneByID(id int) error
	CheckUserSourceStone(req *dto.CheckUserSourceStoneRequest) (*dto.CheckUserSourceStoneResponse, error)
}