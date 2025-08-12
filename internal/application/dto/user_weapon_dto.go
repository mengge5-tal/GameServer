package dto

// UserWeaponResponse represents user weapon ownership response
type UserWeaponResponse struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	WeaponID int `json:"weapon_id"`
}

// UserWeaponDetailResponse represents user weapon ownership with weapon details
type UserWeaponDetailResponse struct {
	ID       int             `json:"id"`
	UserID   int             `json:"user_id"`
	WeaponID int             `json:"weapon_id"`
	Weapon   *WeaponResponse `json:"weapon,omitempty"`
}

// AddUserWeaponRequest represents add weapon to user request
type AddUserWeaponRequest struct {
	UserID   int `json:"user_id"`
	WeaponID int `json:"weapon_id"`
}

// RemoveUserWeaponRequest represents remove weapon from user request
type RemoveUserWeaponRequest struct {
	UserID   int `json:"user_id"`
	WeaponID int `json:"weapon_id"`
}

// RemoveUserWeaponByIDRequest represents remove user weapon by ID request
type RemoveUserWeaponByIDRequest struct {
	ID int `json:"id"`
}

// GetUserWeaponsRequest represents get user weapons request
type GetUserWeaponsRequest struct {
	UserID      int  `json:"user_id"`
	WithDetails bool `json:"with_details,omitempty"`
}

// CheckUserWeaponRequest represents check if user owns weapon request
type CheckUserWeaponRequest struct {
	UserID   int `json:"user_id"`
	WeaponID int `json:"weapon_id"`
}

// CheckUserWeaponResponse represents check user weapon response
type CheckUserWeaponResponse struct {
	UserID    int  `json:"user_id"`
	WeaponID  int  `json:"weapon_id"`
	OwnsWeapon bool `json:"owns_weapon"`
}