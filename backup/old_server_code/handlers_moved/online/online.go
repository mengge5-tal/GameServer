package online

import (
	"database/sql"
	"log"
)

// OnlineService 在线状态管理服务
type OnlineService struct {
	db *sql.DB
}

// NewOnlineService 创建在线状态服务
func NewOnlineService(db *sql.DB) *OnlineService {
	return &OnlineService{db: db}
}

// SetUserOnline 设置用户在线状态
func (s *OnlineService) SetUserOnline(userID int) error {
	_, err := s.db.Exec("UPDATE user SET online_status = 1 WHERE userid = ?", userID)
	if err != nil {
		log.Printf("Failed to set user %d online: %v", userID, err)
		return err
	}
	log.Printf("User %d set to online", userID)
	return nil
}

// SetUserOffline 设置用户离线状态
func (s *OnlineService) SetUserOffline(userID int) error {
	_, err := s.db.Exec("UPDATE user SET online_status = 0 WHERE userid = ?", userID)
	if err != nil {
		log.Printf("Failed to set user %d offline: %v", userID, err)
		return err
	}
	log.Printf("User %d set to offline", userID)
	return nil
}

// GetUserOnlineStatus 获取用户在线状态
func (s *OnlineService) GetUserOnlineStatus(userID int) (bool, error) {
	var status int
	err := s.db.QueryRow("SELECT online_status FROM user WHERE userid = ?", userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("Failed to get user %d online status: %v", userID, err)
		return false, err
	}
	return status == 1, nil
}

// GetOnlineUsers 获取所有在线用户
func (s *OnlineService) GetOnlineUsers() ([]int, error) {
	rows, err := s.db.Query("SELECT userid FROM user WHERE online_status = 1")
	if err != nil {
		log.Printf("Failed to get online users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Error scanning user ID: %v", err)
			continue
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// SetAllUsersOffline 设置所有用户为离线状态（服务器启动时调用）
func (s *OnlineService) SetAllUsersOffline() error {
	result, err := s.db.Exec("UPDATE user SET online_status = 0")
	if err != nil {
		log.Printf("Failed to set all users offline: %v", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("Set %d users to offline status", rowsAffected)
	return nil
}