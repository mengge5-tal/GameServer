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

// SourceStoneResponse represents source stone response
type SourceStoneResponse struct {
	EquipID    int `json:"equipid"`
	SourceType int `json:"sourcetype"`
	Count      int `json:"count"`
	Quality    int `json:"quality"`
	UserID     int `json:"userid"`
}
