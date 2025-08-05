package models

// Equipment represents equipment data structure
type Equipment struct {
	EquipID     int    `json:"equipid"`
	UserID      int    `json:"userid"`
	EquipName   string `json:"equipname"`
	Attack      int    `json:"attack"`
	Defense     int    `json:"defense"`
	Health      int    `json:"health"`
	Type        int    `json:"type"`
}