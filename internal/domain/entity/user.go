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

// PlayerRankingEntry represents a player's ranking entry
type PlayerRankingEntry struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Value    int    `json:"value"`
	Position int    `json:"position"`
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
	Quality            int    `json:"quality"`
	SourceStoneType    int    `json:"sourcestonetype"`
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

// Union represents a union/guild entity
type Union struct {
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

// Validate validates union data
func (u *Union) Validate() error {
	if len(u.UnionName) < 2 || len(u.UnionName) > 100 {
		return NewDomainError("工会名称长度必须在2-100字符之间")
	}
	if u.ChairpersonID <= 0 {
		return NewDomainError("会长ID必须为正数")
	}
	if len(u.ChairpersonName) < 1 || len(u.ChairpersonName) > 100 {
		return NewDomainError("会长名称长度必须在1-100字符之间")
	}
	if u.ChairpersonLevel <= 0 {
		return NewDomainError("会长等级必须为正数")
	}
	if u.UnionLevel <= 0 {
		return NewDomainError("工会等级必须为正数")
	}
	if u.UnionMembers < 0 {
		return NewDomainError("工会成员数不能为负数")
	}
	if u.Experience < 0 {
		return NewDomainError("工会经验值不能为负数")
	}
	if len(u.UnionDesc) > 1000 {
		return NewDomainError("工会简介不能超过1000字符")
	}
	return nil
}

// UnionMember represents a union member entity
type UnionMember struct {
	ID          int       `json:"id"`
	UnionID     int       `json:"unionid"`
	UnionName   string    `json:"unionname"`
	MemberID    int       `json:"memberid"`
	MemberLevel int       `json:"memberlevel"`
	JoinedTime  time.Time `json:"joined_time"`
	RoleID      int       `json:"roleid"` // 0=普通成员, 1=副会长, 2=会长
}

// Validate validates union member data
func (um *UnionMember) Validate() error {
	if um.UnionID <= 0 {
		return NewDomainError("工会ID必须为正数")
	}
	if len(um.UnionName) < 1 || len(um.UnionName) > 100 {
		return NewDomainError("工会名称长度必须在1-100字符之间")
	}
	if um.MemberID <= 0 {
		return NewDomainError("成员ID必须为正数")
	}
	if um.MemberLevel <= 0 {
		return NewDomainError("成员等级必须为正数")
	}
	if um.RoleID < 0 || um.RoleID > 2 {
		return NewDomainError("成员角色ID必须在0-2之间")
	}
	return nil
}

// GetRoleName returns the role name based on role ID
func (um *UnionMember) GetRoleName() string {
	switch um.RoleID {
	case 0:
		return "普通成员"
	case 1:
		return "副会长"
	case 2:
		return "会长"
	default:
		return "未知角色"
	}
}

// UnionExperience represents union level experience requirements
type UnionExperience struct {
	UnionLevel int `json:"unionlevel"`
	Experience int `json:"experience"`
}

// Validate validates union experience data
func (ue *UnionExperience) Validate() error {
	if ue.UnionLevel <= 0 {
		return NewDomainError("工会等级必须为正数")
	}
	if ue.Experience < 0 {
		return NewDomainError("所需经验值不能为负数")
	}
	return nil
}

// UnionRequest represents a union join request entity
type UnionRequest struct {
	ID             int       `json:"id"`
	UnionID        int       `json:"unionid"`
	ApplicantID    int       `json:"applicantid"`
	ApplicantName  string    `json:"applicantname"`
	ApplicantLevel int       `json:"applicantlevel"`
	ChairpersonID  int       `json:"chairpersonid"`
	RequestStatus  int       `json:"request_status"` // 1=待处理, 2=通过, 3=拒绝
	RequestTime    time.Time `json:"request_time"`
}

// Validate validates union request data
func (ur *UnionRequest) Validate() error {
	if ur.UnionID <= 0 {
		return NewDomainError("工会ID必须为正数")
	}
	if ur.ApplicantID <= 0 {
		return NewDomainError("申请人ID必须为正数")
	}
	if len(ur.ApplicantName) < 1 || len(ur.ApplicantName) > 100 {
		return NewDomainError("申请人名称长度必须在1-100字符之间")
	}
	if ur.ApplicantLevel <= 0 {
		return NewDomainError("申请人等级必须为正数")
	}
	if ur.ChairpersonID <= 0 {
		return NewDomainError("会长ID必须为正数")
	}
	if ur.RequestStatus < 1 || ur.RequestStatus > 3 {
		return NewDomainError("申请状态必须在1-3之间")
	}
	return nil
}

// GetStatusName returns the status name based on status ID
func (ur *UnionRequest) GetStatusName() string {
	switch ur.RequestStatus {
	case 1:
		return "待处理"
	case 2:
		return "已通过"
	case 3:
		return "已拒绝"
	default:
		return "未知状态"
	}
}

// IsProcessed checks if the request has been processed
func (ur *UnionRequest) IsProcessed() bool {
	return ur.RequestStatus == 2 || ur.RequestStatus == 3
}

// UnionRole constants for member roles
const (
	UnionRoleMember    = 0 // 普通成员
	UnionRoleViceLeader = 1 // 副会长
	UnionRoleLeader    = 2 // 会长
)

// UnionRequestStatus constants for request status
const (
	UnionRequestStatusPending  = 1 // 待处理
	UnionRequestStatusApproved = 2 // 已通过
	UnionRequestStatusRejected = 3 // 已拒绝
)

// UnionInvite represents a union invitation entity
type UnionInvite struct {
	ID                int       `json:"id"`
	InviteFromUser    string    `json:"invitefromuser"`
	InviteToUser      string    `json:"invitetouser"`
	UnionID           int       `json:"unionid"`
	UnionName         string    `json:"unionname"`
	ChairpersonID     int       `json:"chairpersonid"`
	ChairpersonName   string    `json:"chairpersonname"`
	ChairpersonLevel  int       `json:"chairpersonlevel"`
	UnionLevel        int       `json:"unionlevel"`
	CreateTime        time.Time `json:"creattime"`
	Status            string    `json:"status"` // pending, accepted, rejected
}

// Validate validates union invite data
func (ui *UnionInvite) Validate() error {
	if len(ui.InviteFromUser) < 1 || len(ui.InviteFromUser) > 100 {
		return NewDomainError("邀请人用户名长度必须在1-100字符之间")
	}
	if len(ui.InviteToUser) < 1 || len(ui.InviteToUser) > 100 {
		return NewDomainError("被邀请人用户名长度必须在1-100字符之间")
	}
	if ui.UnionID <= 0 {
		return NewDomainError("工会ID必须为正数")
	}
	if len(ui.UnionName) < 2 || len(ui.UnionName) > 100 {
		return NewDomainError("工会名称长度必须在2-100字符之间")
	}
	if ui.ChairpersonID <= 0 {
		return NewDomainError("会长ID必须为正数")
	}
	if len(ui.ChairpersonName) < 1 || len(ui.ChairpersonName) > 100 {
		return NewDomainError("会长名称长度必须在1-100字符之间")
	}
	if ui.ChairpersonLevel <= 0 {
		return NewDomainError("会长等级必须为正数")
	}
	if ui.UnionLevel <= 0 {
		return NewDomainError("工会等级必须为正数")
	}
	if ui.Status != "pending" && ui.Status != "accepted" && ui.Status != "rejected" {
		return NewDomainError("邀请状态必须为 pending, accepted 或 rejected")
	}
	return nil
}

// IsPending checks if the invite is pending
func (ui *UnionInvite) IsPending() bool {
	return ui.Status == "pending"
}

// IsProcessed checks if the invite has been processed
func (ui *UnionInvite) IsProcessed() bool {
	return ui.Status == "accepted" || ui.Status == "rejected"
}

// GetStatusName returns the Chinese name for invite status
func (ui *UnionInvite) GetStatusName() string {
	switch ui.Status {
	case "pending":
		return "待处理"
	case "accepted":
		return "已接受"
	case "rejected":
		return "已拒绝"
	default:
		return "未知状态"
	}
}

// UnionInviteStatus constants for invite status
const (
	UnionInviteStatusPending  = "pending"
	UnionInviteStatusAccepted = "accepted"
	UnionInviteStatusRejected = "rejected"
)
