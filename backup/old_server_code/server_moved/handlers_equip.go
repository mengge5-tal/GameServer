package server

import (
	"database/sql"
	"encoding/json"
	"log"
)

// Equipment 装备数据结构
type Equipment struct {
	EquipID     int    `json:"equipid"`
	Type        int    `json:"type"` // 装备类型
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
}

// SaveEquipRequest 保存装备信息请求
type SaveEquipRequest struct {
	Equipment Equipment `json:"equipment"`
}

// DelEquipRequest 删除装备请求
type DelEquipRequest struct {
	EquipID int `json:"equipid"`
}

// BatchDelEquipRequest 批量删除装备请求
type BatchDelEquipRequest struct {
	Quality int `json:"quality"`
}

// generateEquipID 生成装备ID的函数
func generateEquipID(db *sql.DB, equipType, quality int) (int, error) {
	// 查询当前type和quality组合的装备数量
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM equip 
		WHERE COALESCE(type, 1) = ? AND quality = ?`,
		equipType, quality).Scan(&count)

	if err != nil {
		return 0, err
	}

	// 生成equipid: [type][quality][6位序号]
	// 例如: type=4, quality=1, count=1 -> 41000001
	sequence := count + 1
	equipID := equipType*10000000 + quality*1000000 + sequence

	return equipID, nil
}

// handleGetEquip 获取装备信息
func handleGetEquip(c *Client, message *Message, db *sql.DB) *Response {
	rows, err := db.Query(`
		SELECT equipid, COALESCE(type, 1), quality, damage, crit, critdamage, damagespeed, 
		       bloodsuck, hp, movespeed, equipname, userid, defense, goodfortune 
		FROM equip WHERE userid = ?`, c.UserID)
	if err != nil {
		log.Printf("Database error getting equipment: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get equipment")
	}
	defer rows.Close()

	var equipments []Equipment
	for rows.Next() {
		var eq Equipment
		err := rows.Scan(&eq.EquipID, &eq.Type, &eq.Quality, &eq.Damage, &eq.Crit, &eq.CritDamage,
			&eq.DamageSpeed, &eq.BloodSuck, &eq.HP, &eq.MoveSpeed, &eq.EquipName,
			&eq.UserID, &eq.Defense, &eq.GoodFortune)
		if err != nil {
			log.Printf("Error scanning equipment: %v", err)
			continue
		}
		equipments = append(equipments, eq)
	}

	return NewSuccessResponse(message.RequestID, equipments)
}

// handleSaveEquip 保存装备信息
func handleSaveEquip(c *Client, message *Message, db *sql.DB) *Response {
	var saveReq SaveEquipRequest

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}

	err = json.Unmarshal(dataBytes, &saveReq)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid equipment data")
	}

	// 验证必需参数
	if saveReq.Equipment.Type <= 0 || saveReq.Equipment.Quality <= 0 {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Type and quality are required and must be positive")
	}

	// 设置用户ID
	saveReq.Equipment.UserID = c.UserID

	// 检查是否是更新操作（equipid不为0）
	if saveReq.Equipment.EquipID != 0 {
		// 更新装备
		_, err := db.Exec(`
			UPDATE equip SET type=?, quality=?, damage=?, crit=?, critdamage=?, damagespeed=?,
			                 bloodsuck=?, hp=?, movespeed=?, equipname=?, defense=?, goodfortune=?
			WHERE equipid=? AND userid=?`,
			saveReq.Equipment.Type, saveReq.Equipment.Quality, saveReq.Equipment.Damage, saveReq.Equipment.Crit,
			saveReq.Equipment.CritDamage, saveReq.Equipment.DamageSpeed, saveReq.Equipment.BloodSuck,
			saveReq.Equipment.HP, saveReq.Equipment.MoveSpeed, saveReq.Equipment.EquipName,
			saveReq.Equipment.Defense, saveReq.Equipment.GoodFortune,
			saveReq.Equipment.EquipID, saveReq.Equipment.UserID)

		if err != nil {
			log.Printf("Database error updating equipment: %v", err)
			return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to update equipment")
		}
	} else {
		// 新增装备 - 生成equipid
		equipID, err := generateEquipID(db, saveReq.Equipment.Type, saveReq.Equipment.Quality)
		if err != nil {
			log.Printf("Error generating equipid: %v", err)
			return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to generate equipment ID")
		}

		saveReq.Equipment.EquipID = equipID

		// 插入新装备
		_, err = db.Exec(`
			INSERT INTO equip (equipid, type, quality, damage, crit, critdamage, damagespeed, 
			                   bloodsuck, hp, movespeed, equipname, userid, defense, goodfortune)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			saveReq.Equipment.EquipID, saveReq.Equipment.Type, saveReq.Equipment.Quality,
			saveReq.Equipment.Damage, saveReq.Equipment.Crit, saveReq.Equipment.CritDamage,
			saveReq.Equipment.DamageSpeed, saveReq.Equipment.BloodSuck, saveReq.Equipment.HP,
			saveReq.Equipment.MoveSpeed, saveReq.Equipment.EquipName, saveReq.Equipment.UserID,
			saveReq.Equipment.Defense, saveReq.Equipment.GoodFortune)

		if err != nil {
			log.Printf("Database error saving equipment: %v", err)
			return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to save equipment")
		}
	}

	return NewSuccessResponse(message.RequestID, saveReq.Equipment)
}

// handleDelEquip 删除装备
func handleDelEquip(c *Client, message *Message, db *sql.DB) *Response {
	var delReq DelEquipRequest

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}

	err = json.Unmarshal(dataBytes, &delReq)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid delete data")
	}

	result, err := db.Exec("DELETE FROM equip WHERE equipid = ? AND userid = ?",
		delReq.EquipID, c.UserID)
	if err != nil {
		log.Printf("Database error deleting equipment: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to delete equipment")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Equipment not found")
	}

	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"equipid": delReq.EquipID,
		"deleted": true,
	})
}

// handleBatchDelEquip 批量删除装备
func handleBatchDelEquip(c *Client, message *Message, db *sql.DB) *Response {
	var batchDelReq BatchDelEquipRequest

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}

	err = json.Unmarshal(dataBytes, &batchDelReq)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid batch delete data")
	}

	result, err := db.Exec("DELETE FROM equip WHERE quality = ? AND userid = ?",
		batchDelReq.Quality, c.UserID)
	if err != nil {
		log.Printf("Database error batch deleting equipment: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to batch delete equipment")
	}

	rowsAffected, _ := result.RowsAffected()
	return NewSuccessResponse(message.RequestID, map[string]interface{}{
		"quality":       batchDelReq.Quality,
		"deleted_count": rowsAffected,
	})
}

// PlayerInfo 玩家信息结构
type PlayerInfo struct {
	UserID      int `json:"userid"`
	Level       int `json:"level"`
	Experience  int `json:"experience"`
	GameLevel   int `json:"gamelevel"`
	BloodEnergy int `json:"bloodenergy"`
}

// handleGetPlayerInfo 获取玩家信息
func handleGetPlayerInfo(c *Client, message *Message, db *sql.DB) *Response {
	var player PlayerInfo
	var level, experience, gamelevel, bloodEnergy sql.NullInt64

	err := db.QueryRow("SELECT userid, level, experience, gamelevel, bloodenergy FROM playerinfo WHERE userid = ?",
		c.UserID).Scan(&player.UserID, &level, &experience, &gamelevel, &bloodEnergy)

	if err != nil {
		if err == sql.ErrNoRows {
			// 如果玩家信息不存在，创建默认记录
			_, err = db.Exec("INSERT INTO playerinfo (userid, level, experience, gamelevel, bloodenergy) VALUES (?, 1, 0, 1, 100)",
				c.UserID)
			if err != nil {
				log.Printf("Database error creating player info: %v", err)
				return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to create player info")
			}
			// 设置默认值
			player.UserID = c.UserID
			player.Level = 1
			player.Experience = 0
			player.GameLevel = 1
			player.BloodEnergy = 100
		} else {
			log.Printf("Database error getting player info: %v", err)
			return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get player info")
		}
	} else {
		// 处理可能的 NULL 值，设置默认值
		if level.Valid {
			player.Level = int(level.Int64)
		} else {
			player.Level = 1
		}

		if experience.Valid {
			player.Experience = int(experience.Int64)
		} else {
			player.Experience = 0
		}

		if gamelevel.Valid {
			player.GameLevel = int(gamelevel.Int64)
		} else {
			player.GameLevel = 1
		}

		if bloodEnergy.Valid {
			player.BloodEnergy = int(bloodEnergy.Int64)
		} else {
			player.BloodEnergy = 100
		}
	}

	return NewSuccessResponse(message.RequestID, player)
}

// handleUpdatePlayerInfo 更新玩家信息
func handleUpdatePlayerInfo(c *Client, message *Message, db *sql.DB) *Response {
	var player PlayerInfo

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
	}

	err = json.Unmarshal(dataBytes, &player)
	if err != nil {
		return NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid player data")
	}

	_, err = db.Exec("UPDATE playerinfo SET level=?, experience=?, gamelevel=?, bloodenergy=? WHERE userid=?",
		player.Level, player.Experience, player.GameLevel, player.BloodEnergy, c.UserID)

	if err != nil {
		log.Printf("Database error updating player info: %v", err)
		return NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to update player info")
	}

	player.UserID = c.UserID
	return NewSuccessResponse(message.RequestID, player)
}