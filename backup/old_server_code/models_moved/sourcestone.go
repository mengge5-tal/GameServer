package models

// Sourcestone represents a sourcestone record
type Sourcestone struct {
	EquipID    int `json:"equipid" db:"equipid"`
	SourceType int `json:"sourcetype" db:"sourcetype"`
	Count      int `json:"count" db:"count"`
	Quality    int `json:"quality" db:"quality"`
	UserID     int `json:"userid" db:"userid"`
}

// CreateSourcestoneRequest represents request data for creating a sourcestone
type CreateSourcestoneRequest struct {
	EquipID    int `json:"equipid"`
	SourceType int `json:"sourcetype"`
	Count      int `json:"count"`
	Quality    int `json:"quality"`
}

// UpdateSourcestoneRequest represents request data for updating a sourcestone
type UpdateSourcestoneRequest struct {
	EquipID    int `json:"equipid"`
	SourceType *int `json:"sourcetype,omitempty"`
	Count      *int `json:"count,omitempty"`
	Quality    *int `json:"quality,omitempty"`
}

// DeleteSourcestoneRequest represents request data for deleting a sourcestone
type DeleteSourcestoneRequest struct {
	EquipID int `json:"equipid"`
}

// GetSourcestonesRequest represents request data for querying sourcestones
type GetSourcestonesRequest struct {
	EquipID    *int `json:"equipid,omitempty"`
	SourceType *int `json:"sourcetype,omitempty"`
	Quality    *int `json:"quality,omitempty"`
}