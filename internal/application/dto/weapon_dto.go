package dto

// WeaponResponse represents weapon response
type WeaponResponse struct {
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

// CreateWeaponRequest represents create weapon request
type CreateWeaponRequest struct {
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

// UpdateWeaponRequest represents update weapon request
type UpdateWeaponRequest struct {
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

// DeleteWeaponRequest represents delete weapon request
type DeleteWeaponRequest struct {
	WeaponID int `json:"weapon_id"`
}

// GetWeaponRequest represents get weapon by ID request
type GetWeaponRequest struct {
	WeaponID int `json:"weapon_id"`
}