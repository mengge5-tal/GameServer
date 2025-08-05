package repository

import (
	"database/sql"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
)

// mysqlFriendRepository implements FriendRepository
type mysqlFriendRepository struct {
	db *sql.DB
}

// NewMySQLFriendRepository creates a new MySQL friend repository
func NewMySQLFriendRepository(db *sql.DB) repository.FriendRepository {
	return &mysqlFriendRepository{db: db}
}

// GetFriendsByUserID retrieves all friends for a user
func (r *mysqlFriendRepository) GetFriendsByUserID(userID int) ([]*entity.Friend, error) {
	query := `SELECT id, fromuserid, touserid, status, created_at, updated_at 
			  FROM friend WHERE (fromuserid = ? OR touserid = ?) AND status = 'accepted'`
	
	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []*entity.Friend
	for rows.Next() {
		friend := &entity.Friend{}
		err := rows.Scan(
			&friend.ID, &friend.FromUserID, &friend.ToUserID,
			&friend.Status, &friend.CreatedAt, &friend.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}

	return friends, rows.Err()
}

// GetFriendRequestsByUserID retrieves all friend requests for a user
func (r *mysqlFriendRepository) GetFriendRequestsByUserID(userID int) ([]*entity.FriendRequest, error) {
	query := `SELECT id, fromuserid, touserid, message, status, created_at, updated_at 
			  FROM friend_request WHERE touserid = ? AND status = 'pending'`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entity.FriendRequest
	for rows.Next() {
		request := &entity.FriendRequest{}
		err := rows.Scan(
			&request.ID, &request.FromUserID, &request.ToUserID,
			&request.Message, &request.Status, &request.CreatedAt, &request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, rows.Err()
}

// CreateFriendRequest creates a new friend request
func (r *mysqlFriendRepository) CreateFriendRequest(request *entity.FriendRequest) error {
	query := `INSERT INTO friend_request (fromuserid, touserid, message, status) 
			  VALUES (?, ?, ?, ?)`
	
	result, err := r.db.Exec(query, request.FromUserID, request.ToUserID, request.Message, request.Status)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	request.ID = int(id)
	return nil
}

// AcceptFriendRequest accepts a friend request and creates friendship
func (r *mysqlFriendRepository) AcceptFriendRequest(requestID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the request details
	var fromUserID, toUserID int
	query := "SELECT fromuserid, touserid FROM friend_request WHERE id = ?"
	err = tx.QueryRow(query, requestID).Scan(&fromUserID, &toUserID)
	if err != nil {
		return err
	}

	// Update request status
	updateQuery := "UPDATE friend_request SET status = 'accepted' WHERE id = ?"
	_, err = tx.Exec(updateQuery, requestID)
	if err != nil {
		return err
	}

	// Create friend relationship
	insertQuery := "INSERT INTO friend (fromuserid, touserid, status) VALUES (?, ?, 'accepted')"
	_, err = tx.Exec(insertQuery, fromUserID, toUserID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RejectFriendRequest rejects a friend request
func (r *mysqlFriendRepository) RejectFriendRequest(requestID int) error {
	query := "UPDATE friend_request SET status = 'rejected' WHERE id = ?"
	_, err := r.db.Exec(query, requestID)
	return err
}

// RemoveFriend removes a friendship
func (r *mysqlFriendRepository) RemoveFriend(fromUserID, toUserID int) error {
	query := "DELETE FROM friend WHERE (fromuserid = ? AND touserid = ?) OR (fromuserid = ? AND touserid = ?)"
	_, err := r.db.Exec(query, fromUserID, toUserID, toUserID, fromUserID)
	return err
}

// AreFriends checks if two users are friends
func (r *mysqlFriendRepository) AreFriends(userID1, userID2 int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM friend 
			  WHERE ((fromuserid = ? AND touserid = ?) OR (fromuserid = ? AND touserid = ?)) 
			  AND status = 'accepted'`
	err := r.db.QueryRow(query, userID1, userID2, userID2, userID1).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasPendingRequest checks if there's a pending friend request
func (r *mysqlFriendRepository) HasPendingRequest(fromUserID, toUserID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM friend_request 
			  WHERE fromuserid = ? AND touserid = ? AND status = 'pending'`
	err := r.db.QueryRow(query, fromUserID, toUserID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}