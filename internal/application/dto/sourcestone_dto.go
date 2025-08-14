package dto

// SourceStoneResponse represents source stone response data
type SourceStoneResponse struct {
	SourceStoneID      int    `json:"sourcestoneid"`
	SourceStoneName    string `json:"sourcestonename"`
	SourceStoneQuality string `json:"sourcestonequality"`
	SourceStoneEffect  string `json:"sourcestoneeffect"`
}

// GetSourceStoneRequest represents get source stone by ID request
type GetSourceStoneRequest struct {
	SourceStoneID int `json:"sourcestoneid"`
}