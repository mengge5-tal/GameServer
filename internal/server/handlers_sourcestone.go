package server

import (
	"database/sql"
	"encoding/json"
	"log"

	"GameServer/internal/models"
)

// HandleCreateSourcestone handles creating a new sourcestone
func handleCreateSourcestone(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}

	var req models.CreateSourcestoneRequest
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid data format")
	}

	if err := json.Unmarshal(dataBytes, &req); err != nil {
		log.Printf("Failed to parse create sourcestone request: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request format")
	}

	// Insert into database
	_, err = db.Exec(`
		INSERT INTO sourcestone (equipid, sourcetype, count, quality, userid) 
		VALUES (?, ?, ?, ?, ?)`,
		req.EquipID, req.SourceType, req.Count, req.Quality, c.UserID)
	
	if err != nil {
		log.Printf("Failed to create sourcestone: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to create sourcestone")
	}

	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"message": "Sourcestone created successfully",
		"equipid": req.EquipID,
	})
}

// HandleGetSourcestones handles retrieving sourcestones for a user
func handleGetSourcestones(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}

	var req models.GetSourcestonesRequest
	if message.Data != nil {
		dataBytes, err := json.Marshal(message.Data)
		if err != nil {
			log.Printf("Failed to marshal data: %v", err)
			return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid data format")
		}

		if err := json.Unmarshal(dataBytes, &req); err != nil {
			log.Printf("Failed to parse get sourcestones request: %v", err)
			return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request format")
		}
	}

	// Build query with optional filters
	query := "SELECT equipid, sourcetype, count, quality, userid FROM sourcestone WHERE userid = ?"
	args := []interface{}{c.UserID}
	
	if req.EquipID != nil {
		query += " AND equipid = ?"
		args = append(args, *req.EquipID)
	}
	if req.SourceType != nil {
		query += " AND sourcetype = ?"
		args = append(args, *req.SourceType)
	}
	if req.Quality != nil {
		query += " AND quality = ?"
		args = append(args, *req.Quality)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Failed to query sourcestones: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to retrieve sourcestones")
	}
	defer rows.Close()

	var sourcestones []models.Sourcestone
	for rows.Next() {
		var ss models.Sourcestone
		err := rows.Scan(&ss.EquipID, &ss.SourceType, &ss.Count, &ss.Quality, &ss.UserID)
		if err != nil {
			log.Printf("Failed to scan sourcestone: %v", err)
			continue
		}
		sourcestones = append(sourcestones, ss)
	}

	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"sourcestones": sourcestones,
		"count":        len(sourcestones),
	})
}

// HandleGetSourcestone handles retrieving a specific sourcestone by equipid
func handleGetSourcestone(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}

	var req struct {
		EquipID int `json:"equipid"`
	}
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid data format")
	}

	if err := json.Unmarshal(dataBytes, &req); err != nil {
		log.Printf("Failed to parse get sourcestone request: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request format")
	}

	var ss models.Sourcestone
	err = db.QueryRow("SELECT equipid, sourcetype, count, quality, userid FROM sourcestone WHERE userid = ? AND equipid = ?",
		c.UserID, req.EquipID).Scan(&ss.EquipID, &ss.SourceType, &ss.Count, &ss.Quality, &ss.UserID)

	if err != nil {
		if err == sql.ErrNoRows {
			return NewErrorResponse(message.RequestID, CodeInvalidParams, "Sourcestone not found")
		}
		log.Printf("Failed to get sourcestone: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get sourcestone")
	}

	return NewSuccessResponse(message.RequestID, ss)
}

// HandleUpdateSourcestone handles updating an existing sourcestone
func handleUpdateSourcestone(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}

	var req models.UpdateSourcestoneRequest
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid data format")
	}

	if err := json.Unmarshal(dataBytes, &req); err != nil {
		log.Printf("Failed to parse update sourcestone request: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request format")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	
	if req.SourceType != nil {
		setParts = append(setParts, "sourcetype = ?")
		args = append(args, *req.SourceType)
	}
	if req.Count != nil {
		setParts = append(setParts, "count = ?")
		args = append(args, *req.Count)
	}
	if req.Quality != nil {
		setParts = append(setParts, "quality = ?")
		args = append(args, *req.Quality)
	}
	
	if len(setParts) == 0 {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "No fields to update")
	}

	query := "UPDATE sourcestone SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE userid = ? AND equipid = ?"
	args = append(args, c.UserID, req.EquipID)

	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Failed to update sourcestone: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to update sourcestone")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Sourcestone not found")
	}

	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"message": "Sourcestone updated successfully",
		"equipid": req.EquipID,
	})
}

// HandleDeleteSourcestone handles deleting a sourcestone
func handleDeleteSourcestone(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}

	var req models.DeleteSourcestoneRequest
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		log.Printf("Failed to marshal data: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid data format")
	}

	if err := json.Unmarshal(dataBytes, &req); err != nil {
		log.Printf("Failed to parse delete sourcestone request: %v", err)
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request format")
	}

	result, err := db.Exec("DELETE FROM sourcestone WHERE userid = ? AND equipid = ?", c.UserID, req.EquipID)
	if err != nil {
		log.Printf("Failed to delete sourcestone: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to delete sourcestone")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Sourcestone not found")
	}

	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"message": "Sourcestone deleted successfully",
		"equipid": req.EquipID,
	})
}

// HandleDeleteAllSourcestones handles deleting all sourcestones for a user
func handleDeleteAllSourcestones(c *Client, message *Message, db *sql.DB) *Response {
	if c.UserID == 0 {
		return NewErrorResponse(message.RequestID, CodeUnauthorized, "User not authenticated")
	}

	result, err := db.Exec("DELETE FROM sourcestone WHERE userid = ?", c.UserID)
	if err != nil {
		log.Printf("Failed to delete all sourcestones: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to delete all sourcestones")
	}

	rowsAffected, _ := result.RowsAffected()
	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"message":       "All sourcestones deleted successfully",
		"deleted_count": rowsAffected,
	})
}