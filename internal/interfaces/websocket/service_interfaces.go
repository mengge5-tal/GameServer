package websocket

import "GameServer/internal/application/dto"

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
	SaveEquipment(req *dto.SaveEquipmentRequest) error
	DeleteEquipment(equipID, userID int) error
	GetUserSourceStones(userID int) ([]*dto.SourceStoneResponse, error)
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
}

// RankingServiceInterface defines the interface for ranking service used by websocket handlers
type RankingServiceInterface interface {
	GetRanking(req *dto.GetRankingRequest) ([]*dto.RankingResponse, error)
	GetUserRanking(userID int, rankType string) (*dto.UserRankingResponse, error)
	UpdateUserRankings(userID int) error
	RefreshAllRankings() error
}