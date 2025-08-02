package friend

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// 好友信息结构
type Friend struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
	Status   string `json:"status"`
	AddedAt  string `json:"added_at"`
}

// 好友申请结构
type FriendRequest struct {
	ID           int    `json:"id"`
	FromUserID   int    `json:"fromuserid"`
	FromUsername string `json:"fromusername"`
	ToUserID     int    `json:"touserid"`
	ToUsername   string `json:"tousername"`
	Message      string `json:"message"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// 排行榜条目结构
type RankEntry struct {
	UserID    int    `json:"userid"`
	Username  string `json:"username"`
	RankType  string `json:"rank_type"`
	RankValue int    `json:"rank_value"`
	Position  int    `json:"position"`
	UpdatedAt string `json:"updated_at"`
}

// 处理好友相关消息
func (c *Client) HandleFriend(message *Message) {
	switch message.Action {
	case ActionGetFriends:
		c.HandleGetFriends(message)
	case ActionAddFriend:
		c.HandleAddFriend(message)
	case ActionDelFriend:
		c.HandleDelFriend(message)
	case ActionFriendRequest:
		c.HandleFriendRequest(message)
	case ActionFriendResponse:
		c.HandleFriendResponse(message)
	default:
		response := NewErrorResponse(message.RequestID, CodeInvalidRequest, "Unknown friend action")
		c.SendResponse(response)
	}
}

// 获取好友列表
func (c *Client) HandleGetFriends(message *Message) {
	rows, err := c.Hub.DB.Query(`
		SELECT f.touserid, u.username, f.status, f.created_at 
		FROM friend f 
		JOIN user u ON f.touserid = u.userid 
		WHERE f.fromuserid = ? AND f.status = 'accepted'
		UNION
		SELECT f.fromuserid, u.username, f.status, f.created_at 
		FROM friend f 
		JOIN user u ON f.fromuserid = u.userid 
		WHERE f.touserid = ? AND f.status = 'accepted'`,
		c.UserID, c.UserID)

	if err != nil {
		log.Printf("Database error getting friends: %v", err)
		response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get friends")
		c.SendResponse(response)
		return
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var friend Friend
		var createdAt time.Time
		err := rows.Scan(&friend.UserID, &friend.Username, &friend.Status, &createdAt)
		if err != nil {
			log.Printf("Error scanning friend: %v", err)
			continue
		}
		friend.AddedAt = createdAt.Format("2006-01-02 15:04:05")
		friends = append(friends, friend)
	}

	response := NewSuccessResponse(message.RequestID, friends)
	c.SendResponse(response)
}

// 添加好友请求
type AddFriendRequest struct {
	ToUserID int    `json:"touserid"`
	Message  string `json:"message"`
}

// 添加好友（发送好友申请）
func (c *Client) HandleAddFriend(message *Message) {
	var addReq AddFriendRequest

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
		c.SendResponse(response)
		return
	}

	err = json.Unmarshal(dataBytes, &addReq)
	if err != nil {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid add friend data")
		c.SendResponse(response)
		return
	}

	// 检查目标用户是否存在
	var targetUsername string
	err = c.Hub.DB.QueryRow("SELECT username FROM user WHERE userid = ?", addReq.ToUserID).Scan(&targetUsername)
	if err != nil {
		var response *Response
		if err == sql.ErrNoRows {
			response = NewErrorResponse(message.RequestID, CodeUserNotFound, "Target user not found")
		} else {
			response = NewErrorResponse(message.RequestID, CodeDatabaseError, "Database error")
		}
		c.SendResponse(response)
		return
	}

	// 检查是否已经是好友
	var existingFriend int
	err = c.Hub.DB.QueryRow(`
		SELECT 1 FROM friend 
		WHERE (fromuserid = ? AND touserid = ?) 
		   OR (fromuserid = ? AND touserid = ?)`,
		c.UserID, addReq.ToUserID, addReq.ToUserID, c.UserID).Scan(&existingFriend)

	if err != sql.ErrNoRows {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Already friends or request exists")
		c.SendResponse(response)
		return
	}

	// 插入好友申请
	_, err = c.Hub.DB.Exec(`
		INSERT INTO friend_request (fromuserid, touserid, message, status) 
		VALUES (?, ?, ?, 'pending')
		ON DUPLICATE KEY UPDATE message = VALUES(message), status = 'pending', updated_at = CURRENT_TIMESTAMP`,
		c.UserID, addReq.ToUserID, addReq.Message)

	if err != nil {
		log.Printf("Database error adding friend request: %v", err)
		response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to send friend request")
		c.SendResponse(response)
		return
	}

	// 通知目标用户（如果在线）
	if targetClient := c.Hub.GetClientByUserID(addReq.ToUserID); targetClient != nil {
		notification := map[string]interface{}{
			"type":         "friend_request",
			"fromuserid":   c.UserID,
			"fromusername": "", // 需要查询用户名
			"message":      addReq.Message,
		}

		// 查询发送者用户名
		var senderUsername string
		c.Hub.DB.QueryRow("SELECT username FROM user WHERE userid = ?", c.UserID).Scan(&senderUsername)
		notification["fromusername"] = senderUsername

		notifResp := NewSuccessResponse("", notification)
		targetClient.SendResponse(notifResp)
	}

	response := NewSuccessResponse(message.RequestID, map[string]interface{}{
		"touserid": addReq.ToUserID,
		"status":   "request_sent",
	})
	c.SendResponse(response)
}

// 删除好友请求
type DelFriendRequest struct {
	FriendUserID int `json:"frienduserid"`
}

// 删除好友
func (c *Client) HandleDelFriend(message *Message) {
	var delReq DelFriendRequest

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
		c.SendResponse(response)
		return
	}

	err = json.Unmarshal(dataBytes, &delReq)
	if err != nil {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid delete friend data")
		c.SendResponse(response)
		return
	}

	// 删除好友关系
	result, err := c.Hub.DB.Exec(`
		DELETE FROM friend 
		WHERE (fromuserid = ? AND touserid = ?) 
		   OR (fromuserid = ? AND touserid = ?)`,
		c.UserID, delReq.FriendUserID, delReq.FriendUserID, c.UserID)

	if err != nil {
		log.Printf("Database error deleting friend: %v", err)
		response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to delete friend")
		c.SendResponse(response)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Friend relationship not found")
		c.SendResponse(response)
		return
	}

	response := NewSuccessResponse(message.RequestID, map[string]interface{}{
		"frienduserid": delReq.FriendUserID,
		"deleted":      true,
	})
	c.SendResponse(response)
}

// 好友申请响应请求
type FriendResponseRequest struct {
	FromUserID int  `json:"fromuserid"`
	Accept     bool `json:"accept"`
}

// 响应好友申请
func (c *Client) HandleFriendResponse(message *Message) {
	var respReq FriendResponseRequest

	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid request data")
		c.SendResponse(response)
		return
	}

	err = json.Unmarshal(dataBytes, &respReq)
	if err != nil {
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid response data")
		c.SendResponse(response)
		return
	}

	// 检查好友申请是否存在
	var requestID int
	err = c.Hub.DB.QueryRow(`
		SELECT id FROM friend_request 
		WHERE fromuserid = ? AND touserid = ? AND status = 'pending'`,
		respReq.FromUserID, c.UserID).Scan(&requestID)

	if err != nil {
		var response *Response
		if err == sql.ErrNoRows {
			response = NewErrorResponse(message.RequestID, CodeInvalidParams, "Friend request not found")
		} else {
			response = NewErrorResponse(message.RequestID, CodeDatabaseError, "Database error")
		}
		c.SendResponse(response)
		return
	}

	if respReq.Accept {
		// 接受好友申请
		// 更新申请状态
		_, err = c.Hub.DB.Exec(`
			UPDATE friend_request SET status = 'accepted' WHERE id = ?`, requestID)
		if err != nil {
			log.Printf("Database error updating friend request: %v", err)
		}

		// 添加好友关系
		_, err = c.Hub.DB.Exec(`
			INSERT INTO friend (fromuserid, touserid, status) VALUES (?, ?, 'accepted')`,
			respReq.FromUserID, c.UserID)
		if err != nil {
			log.Printf("Database error adding friend relationship: %v", err)
			response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to add friend")
			c.SendResponse(response)
			return
		}
	} else {
		// 拒绝好友申请
		_, err = c.Hub.DB.Exec(`
			UPDATE friend_request SET status = 'rejected' WHERE id = ?`, requestID)
		if err != nil {
			log.Printf("Database error updating friend request: %v", err)
		}
	}

	response := NewSuccessResponse(message.RequestID, map[string]interface{}{
		"fromuserid": respReq.FromUserID,
		"accepted":   respReq.Accept,
	})
	c.SendResponse(response)
}

// 处理好友申请（获取收到的申请）
func (c *Client) HandleFriendRequest(message *Message) {
	rows, err := c.Hub.DB.Query(`
		SELECT fr.id, fr.fromuserid, u.username, fr.touserid, fr.message, fr.status, fr.created_at
		FROM friend_request fr
		JOIN user u ON fr.fromuserid = u.userid
		WHERE fr.touserid = ? AND fr.status = 'pending'
		ORDER BY fr.created_at DESC`,
		c.UserID)

	if err != nil {
		log.Printf("Database error getting friend requests: %v", err)
		response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get friend requests")
		c.SendResponse(response)
		return
	}
	defer rows.Close()

	var requests []FriendRequest
	for rows.Next() {
		var req FriendRequest
		var createdAt time.Time
		err := rows.Scan(&req.ID, &req.FromUserID, &req.FromUsername, &req.ToUserID,
			&req.Message, &req.Status, &createdAt)
		if err != nil {
			log.Printf("Error scanning friend request: %v", err)
			continue
		}
		req.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		requests = append(requests, req)
	}

	response := NewSuccessResponse(message.RequestID, requests)
	c.SendResponse(response)
}

// 处理排行榜相关消息
func (c *Client) HandleRank(message *Message) {
	switch message.Action {
	case ActionGetAllRank:
		c.HandleGetAllRank(message)
	case ActionGetSelfRank:
		c.HandleGetSelfRank(message)
	default:
		response := NewErrorResponse(message.RequestID, CodeInvalidRequest, "Unknown rank action")
		c.SendResponse(response)
	}
}

// 获取排行榜请求
type GetRankRequest struct {
	RankType string `json:"rank_type"` // level, experience, equipment_power
	Limit    int    `json:"limit"`     // 限制返回条数，默认100
}

// 获取排行榜
func (c *Client) HandleGetAllRank(message *Message) {
	var rankReq GetRankRequest
	rankReq.RankType = "level" // 默认按等级排行
	rankReq.Limit = 100        // 默认返回前100名

	if message.Data != nil {
		dataBytes, err := json.Marshal(message.Data)
		if err == nil {
			json.Unmarshal(dataBytes, &rankReq)
		}
	}

	// 验证排行类型
	if rankReq.RankType != "level" && rankReq.RankType != "experience" && rankReq.RankType != "equipment_power" {
		rankReq.RankType = "level"
	}

	if rankReq.Limit <= 0 || rankReq.Limit > 100 {
		rankReq.Limit = 100
	}

	// 根据排行类型查询
	var query string
	switch rankReq.RankType {
	case "level":
		query = `
			SELECT p.userid, u.username, p.level as rank_value, 0 as position
			FROM playerinfo p
			JOIN user u ON p.userid = u.userid
			ORDER BY p.level DESC, p.experience DESC
			LIMIT ?`
	case "experience":
		query = `
			SELECT p.userid, u.username, p.experience as rank_value, 0 as position
			FROM playerinfo p
			JOIN user u ON p.userid = u.userid
			ORDER BY p.experience DESC, p.level DESC
			LIMIT ?`
	default:
		response := NewErrorResponse(message.RequestID, CodeInvalidParams, "Invalid rank type")
		c.SendResponse(response)
		return
	}

	rows, err := c.Hub.DB.Query(query, rankReq.Limit)
	if err != nil {
		log.Printf("Database error getting rank: %v", err)
		response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get ranking")
		c.SendResponse(response)
		return
	}
	defer rows.Close()

	var rankings []RankEntry
	position := 1
	for rows.Next() {
		var rank RankEntry
		err := rows.Scan(&rank.UserID, &rank.Username, &rank.RankValue, &rank.Position)
		if err != nil {
			log.Printf("Error scanning rank entry: %v", err)
			continue
		}
		rank.RankType = rankReq.RankType
		rank.Position = position
		position++
		rankings = append(rankings, rank)
	}

	response := NewSuccessResponse(message.RequestID, map[string]interface{}{
		"rank_type": rankReq.RankType,
		"rankings":  rankings,
	})
	c.SendResponse(response)
}

// 获取个人排名
func (c *Client) HandleGetSelfRank(message *Message) {
	var rankReq GetRankRequest
	rankReq.RankType = "level" // 默认按等级排行

	if message.Data != nil {
		dataBytes, err := json.Marshal(message.Data)
		if err == nil {
			json.Unmarshal(dataBytes, &rankReq)
		}
	}

	// 验证排行类型
	if rankReq.RankType != "level" && rankReq.RankType != "experience" && rankReq.RankType != "equipment_power" {
		rankReq.RankType = "level"
	}

	// 获取个人排名
	var query string
	switch rankReq.RankType {
	case "level":
		query = `
			SELECT COUNT(*) + 1 as position, 
			       (SELECT level FROM playerinfo WHERE userid = ?) as rank_value
			FROM playerinfo p1
			JOIN playerinfo p2 ON p2.userid = ?
			WHERE p1.level > p2.level 
			   OR (p1.level = p2.level AND p1.experience > p2.experience)`
	case "experience":
		query = `
			SELECT COUNT(*) + 1 as position,
			       (SELECT experience FROM playerinfo WHERE userid = ?) as rank_value
			FROM playerinfo p1
			JOIN playerinfo p2 ON p2.userid = ?
			WHERE p1.experience > p2.experience
			   OR (p1.experience = p2.experience AND p1.level > p2.level)`
	}

	var position, rankValue int
	err := c.Hub.DB.QueryRow(query, c.UserID, c.UserID).Scan(&position, &rankValue)
	if err != nil {
		log.Printf("Database error getting self rank: %v", err)
		response := NewErrorResponse(message.RequestID, CodeDatabaseError, "Failed to get self ranking")
		c.SendResponse(response)
		return
	}

	// 获取用户名
	var username string
	c.Hub.DB.QueryRow("SELECT username FROM user WHERE userid = ?", c.UserID).Scan(&username)

	selfRank := RankEntry{
		UserID:    c.UserID,
		Username:  username,
		RankType:  rankReq.RankType,
		RankValue: rankValue,
		Position:  position,
	}

	response := NewSuccessResponse(message.RequestID, selfRank)
	c.SendResponse(response)
}
