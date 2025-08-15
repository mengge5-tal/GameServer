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
	RankType     string    `json:"rank_type"` // level, experience, equipment_power
	RankValue    int       `json:"rank_value"`
	RankPosition int       `json:"rank_position"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Equipment represents an equipment entity
type Equipment struct {
	EquipID        int    `json:"equipid"`
	Quality        int    `json:"quality"`
	Damage         int    `json:"damage"`
	Crit           int    `json:"crit"`
	CritDamage     int    `json:"critdamage"`
	DamageSpeed    int    `json:"damagespeed"`
	BloodSuck      int    `json:"bloodsuck"`
	HP             int    `json:"hp"`
	MoveSpeed      int    `json:"movespeed"`
	EquipName      string `json:"equipname"`
	SuitID         int    `json:"suitid"`
	SuitName       string `json:"suitname"`
	EquipTypeID    int    `json:"equip_type_id"`
	EquipTypeName  string `json:"equip_type_name"`
	UserID         int    `json:"userid"`
	Defense        int    `json:"defense"`
	GoodFortune    int    `json:"goodfortune"`
	Type           int    `json:"type"`
}

// SourceStone represents a source stone entity
type SourceStone struct {
	SourceStoneID      int    `json:"sourcestoneid"`
	SourceStoneName    string `json:"sourcestonename"`
	SourceStoneQuality string `json:"sourcestonequality"`
	SourceStoneEffect  string `json:"sourcestoneeffect"`
}

// Experience represents experience level mapping
type Experience struct {
	Level int `json:"level"`
	Value int `json:"value"`
}

// Weapon represents a weapon entity
type Weapon struct {
	WeaponID            int    `json:"weapon_id"`
	WeaponName          string `json:"weapon_name"`
	AttackPower         int    `json:"attack_power"`
	AttackSpeed         int    `json:"attack_speed"`
	CriticalStrikeRate  int    `json:"critical_strike_rate"`
	CriticalStrikeDamage int    `json:"critical_strike_damage"`
	LuckyValue          int    `json:"lucky_value"`
	EnhancementLevel    int    `json:"enhancement_level"`
	GrowthValue         int    `json:"growth_value"`
	Quality             int    `json:"quality"`
}

// UserWeapon represents a user's weapon ownership
type UserWeapon struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	WeaponID int `json:"weapon_id"`
}

// UserSourceStone represents a user's source stone ownership
type UserSourceStone struct {
	ID                int `json:"id"`
	UserID            int `json:"userid"`
	SourceStoneID     int `json:"sourcestoneid"`
	SourceStoneCount  int `json:"sourcestonecount"`
}

// UserEquip represents equipped items for a user
type UserEquip struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userid"`
	EquipSlot string `json:"equip_slot"` // 衣服, 鞋子, 戒指, 项链, 头盔, 手套
	EquipID   *int   `json:"equipid"`    // nullable, nil means no equipment in this slot
}

// ValidEquipSlots defines valid equipment slot types
var ValidEquipSlots = []string{"衣服", "鞋子", "戒指", "项链", "头盔", "手套"}

// Validate validates UserEquip data
func (ue *UserEquip) Validate() error {
	if ue.UserID <= 0 {
		return NewDomainError("user ID must be positive")
	}
	
	// Check if equip slot is valid
	isValidSlot := false
	for _, slot := range ValidEquipSlots {
		if ue.EquipSlot == slot {
			isValidSlot = true
			break
		}
	}
	if !isValidSlot {
		return NewDomainError("invalid equipment slot type")
	}
	
	// If EquipID is provided, it must be positive
	if ue.EquipID != nil && *ue.EquipID <= 0 {
		return NewDomainError("equipment ID must be positive")
	}
	
	return nil
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

// UserWithLevel represents a user with level information for recommendations
type UserWithLevel struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Level    int    `json:"level"`
}

// KillCount represents daily kill count statistics for a user
type KillCount struct {
	ID     int    `json:"id"`
	UserID int    `json:"userid"`
	Today  string `json:"today"` // Date in YYYY-MM-DD format
	Normal int    `json:"normal"`
	Elite  int    `json:"elite"`
	Boss   int    `json:"boss"`
	Count  int    `json:"count"` // Total kill count (normal*1 + elite*5 + boss*10)
}

// CalculateTotalKillCount calculates the total kill count based on kill types
func (kc *KillCount) CalculateTotalKillCount() {
	kc.Count = kc.Normal*1 + kc.Elite*5 + kc.Boss*10
}

// KillRankingEntry represents a kill ranking entry with user information
type KillRankingEntry struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Count    int    `json:"count"` // Total kill count
	Rank     int    `json:"rank"`  // Ranking position
}
