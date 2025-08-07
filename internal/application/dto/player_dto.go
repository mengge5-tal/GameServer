package dto

// PlayerInfoResponse represents player information response
type PlayerInfoResponse struct {
	UserID      int `json:"userid"`
	Level       int `json:"level"`
	Experience  int `json:"experience"`
	GameLevel   int `json:"gamelevel"`
	BloodEnergy int `json:"bloodenergy"`
}

// UpdatePlayerRequest represents update player request
type UpdatePlayerRequest struct {
	UserID      int  `json:"userid"`
	Level       *int `json:"level,omitempty"`
	Experience  *int `json:"experience,omitempty"`
	GameLevel   *int `json:"gamelevel,omitempty"`
	BloodEnergy *int `json:"bloodenergy,omitempty"`
}

// EquipmentResponse represents equipment response
type EquipmentResponse struct {
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
	Defense     int    `json:"defense"`
	GoodFortune int    `json:"goodfortune"`
	Type        int    `json:"type"`
}

// SaveEquipmentRequest represents save equipment request
type SaveEquipmentRequest struct {
	Equipment EquipmentData `json:"equipment"`
	// Support both formats: nested and direct
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
	Defense     int    `json:"defense"`
	GoodFortune int    `json:"goodfortune"`
	Type        int    `json:"type"`
}

// EquipmentData represents the equipment data within save request
type EquipmentData struct {
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
	Defense     int    `json:"defense"`
	GoodFortune int    `json:"goodfortune"`
	Type        int    `json:"type"`
}

// GetEffectiveEquipmentData returns the equipment data, prioritizing nested format
func (req *SaveEquipmentRequest) GetEffectiveEquipmentData() EquipmentData {
	// If equipment field has data (nested format), use it
	if req.Equipment.Type > 0 || req.Equipment.Quality > 0 {
		return req.Equipment
	}
	// Otherwise use direct fields
	return EquipmentData{
		EquipID:     req.EquipID,
		Quality:     req.Quality,
		Damage:      req.Damage,
		Crit:        req.Crit,
		CritDamage:  req.CritDamage,
		DamageSpeed: req.DamageSpeed,
		BloodSuck:   req.BloodSuck,
		HP:          req.HP,
		MoveSpeed:   req.MoveSpeed,
		EquipName:   req.EquipName,
		UserID:      req.UserID,
		Defense:     req.Defense,
		GoodFortune: req.GoodFortune,
		Type:        req.Type,
	}
}

// SourceStoneResponse represents source stone response
type SourceStoneResponse struct {
	EquipID    int `json:"equipid"`
	SourceType int `json:"sourcetype"`
	Count      int `json:"count"`
	Quality    int `json:"quality"`
	UserID     int `json:"userid"`
}
