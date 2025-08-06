package entity

import "time"

// User represents a user entity in the domain
type User struct {
	ID           int    `json:"userid"`
	Username     string `json:"username"`
	Password     string `json:"-"` // Never expose password in JSON
	OnlineStatus int    `json:"online_status"`
}

// Validate validates user data
func (u *User) Validate() error {
	if len(u.Username) < 3 || len(u.Username) > 20 {
		return NewDomainError("username must be 3-20 characters")
	}
	if u.Username == "" {
		return NewDomainError("username is required")
	}
	return nil
}

// PlayerInfo represents player information entity
type PlayerInfo struct {
	UserID      int `json:"userid"`
	Level       int `json:"level"`
	Experience  int `json:"experience"`
	GameLevel   int `json:"gamelevel"`
	BloodEnergy int `json:"bloodenergy"`
}

// Friend represents a friend relationship entity
type Friend struct {
	ID         int       `json:"id"`
	FromUserID int       `json:"fromuserid"`
	ToUserID   int       `json:"touserid"`
	Status     string    `json:"status"` // pending, accepted, blocked
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FriendRequest represents a friend request entity
type FriendRequest struct {
	ID         int       `json:"id"`
	FromUserID int       `json:"fromuserid"`
	ToUserID   int       `json:"touserid"`
	Message    string    `json:"message"`
	Status     string    `json:"status"` // pending, accepted, rejected
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Ranking represents a ranking entry entity
type Ranking struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userid"`
	RankType     string    `json:"rank_type"`     // level, experience, equipment_power
	RankValue    int       `json:"rank_value"`
	RankPosition int       `json:"rank_position"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Equipment represents an equipment entity
type Equipment struct {
	EquipID     int    `json:"equipid"`
	Quality     int    `json:"quality"`
	Damage      int    `json:"damage"`
	Crit        int    `json:"crit"`
	CritDamage  int    `json:"critdamage"`
	DamageSpeed int    `json:"damagespeed"`
	BloodSuck   int    `json:"bloodsuck"`
	HP          int    `json:"hp"`
	MoveSpeed   int    `json:"movespeed"`
	EquipName   string `json:"equipname"`
	UserID      int    `json:"userid"`
	Defense     int    `json:"denfense"`
	GoodFortune int    `json:"goodfortune"`
	Type        int    `json:"type"`
}

// SourceStone represents a source stone entity
type SourceStone struct {
	EquipID    int `json:"equipid"`
	SourceType int `json:"sourcetype"`
	Count      int `json:"count"`
	Quality    int `json:"quality"`
	UserID     int `json:"userid"`
}

// Experience represents experience level mapping
type Experience struct {
	Level int `json:"level"`
	Value int `json:"value"`
}

// DomainError represents domain-specific errors
type DomainError struct {
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewDomainError(message string) *DomainError {
	return &DomainError{Message: message}
}