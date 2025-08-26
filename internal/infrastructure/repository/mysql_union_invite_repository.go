package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
)

// mysqlUnionInviteRepository implements UnionInviteRepository
type mysqlUnionInviteRepository struct {
	db *sql.DB
}

// NewMySQLUnionInviteRepository creates a new MySQL union invite repository
func NewMySQLUnionInviteRepository(db *sql.DB) repository.UnionInviteRepository {
	return &mysqlUnionInviteRepository{db: db}
}

// GetByID retrieves a union invite by ID
func (r *mysqlUnionInviteRepository) GetByID(id int) (*entity.UnionInvite, error) {
	invite := &entity.UnionInvite{}
	query := `SELECT id, invitefromuser, invitetouser, unionid, unionname, 
			  chairpersonid, chairpersonname, chairpersonlevel, unionlevel, creattime, status 
			  FROM inviterequests WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(&invite.ID, &invite.InviteFromUser, &invite.InviteToUser,
		&invite.UnionID, &invite.UnionName, &invite.ChairpersonID, &invite.ChairpersonName,
		&invite.ChairpersonLevel, &invite.UnionLevel, &invite.CreateTime, &invite.Status)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取工会邀请信息失败: %w", err)
	}
	return invite, nil
}

// GetByUserID retrieves all invites for a user (as invitee)
func (r *mysqlUnionInviteRepository) GetByUserID(userID int) ([]*entity.UnionInvite, error) {
	// First get username from user ID
	var username string
	userQuery := "SELECT username FROM user WHERE userid = ?"
	err := r.db.QueryRow(userQuery, userID).Scan(&username)
	if err != nil {
		return nil, fmt.Errorf("获取用户名失败: %w", err)
	}

	query := `SELECT id, invitefromuser, invitetouser, unionid, unionname, 
			  chairpersonid, chairpersonname, chairpersonlevel, unionlevel, creattime, status 
			  FROM inviterequests WHERE invitetouser = ? ORDER BY creattime DESC`
	
	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("获取用户邀请列表失败: %w", err)
	}
	defer rows.Close()
	
	var invites []*entity.UnionInvite
	for rows.Next() {
		invite := &entity.UnionInvite{}
		err := rows.Scan(&invite.ID, &invite.InviteFromUser, &invite.InviteToUser,
			&invite.UnionID, &invite.UnionName, &invite.ChairpersonID, &invite.ChairpersonName,
			&invite.ChairpersonLevel, &invite.UnionLevel, &invite.CreateTime, &invite.Status)
		if err != nil {
			return nil, fmt.Errorf("扫描邀请数据失败: %w", err)
		}
		invites = append(invites, invite)
	}
	
	return invites, nil
}

// GetByUnionID retrieves all invites for a union
func (r *mysqlUnionInviteRepository) GetByUnionID(unionID int) ([]*entity.UnionInvite, error) {
	query := `SELECT id, invitefromuser, invitetouser, unionid, unionname, 
			  chairpersonid, chairpersonname, chairpersonlevel, unionlevel, creattime, status 
			  FROM inviterequests WHERE unionid = ? ORDER BY creattime DESC`
	
	rows, err := r.db.Query(query, unionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会邀请列表失败: %w", err)
	}
	defer rows.Close()
	
	var invites []*entity.UnionInvite
	for rows.Next() {
		invite := &entity.UnionInvite{}
		err := rows.Scan(&invite.ID, &invite.InviteFromUser, &invite.InviteToUser,
			&invite.UnionID, &invite.UnionName, &invite.ChairpersonID, &invite.ChairpersonName,
			&invite.ChairpersonLevel, &invite.UnionLevel, &invite.CreateTime, &invite.Status)
		if err != nil {
			return nil, fmt.Errorf("扫描工会邀请数据失败: %w", err)
		}
		invites = append(invites, invite)
	}
	
	return invites, nil
}

// GetPendingByUserID retrieves pending invites for a user
func (r *mysqlUnionInviteRepository) GetPendingByUserID(userID int) ([]*entity.UnionInvite, error) {
	// First get username from user ID
	var username string
	userQuery := "SELECT username FROM user WHERE userid = ?"
	err := r.db.QueryRow(userQuery, userID).Scan(&username)
	if err != nil {
		return nil, fmt.Errorf("获取用户名失败: %w", err)
	}

	query := `SELECT id, invitefromuser, invitetouser, unionid, unionname, 
			  chairpersonid, chairpersonname, chairpersonlevel, unionlevel, creattime, status 
			  FROM inviterequests WHERE invitetouser = ? AND status = 'pending' 
			  ORDER BY creattime ASC`
	
	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("获取待处理邀请失败: %w", err)
	}
	defer rows.Close()
	
	var invites []*entity.UnionInvite
	for rows.Next() {
		invite := &entity.UnionInvite{}
		err := rows.Scan(&invite.ID, &invite.InviteFromUser, &invite.InviteToUser,
			&invite.UnionID, &invite.UnionName, &invite.ChairpersonID, &invite.ChairpersonName,
			&invite.ChairpersonLevel, &invite.UnionLevel, &invite.CreateTime, &invite.Status)
		if err != nil {
			return nil, fmt.Errorf("扫描待处理邀请数据失败: %w", err)
		}
		invites = append(invites, invite)
	}
	
	return invites, nil
}

// Create creates a new union invite
func (r *mysqlUnionInviteRepository) Create(invite *entity.UnionInvite) error {
	query := `INSERT INTO inviterequests (invitefromuser, invitetouser, unionid, unionname, 
			  chairpersonid, chairpersonname, chairpersonlevel, unionlevel, creattime, status) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query, invite.InviteFromUser, invite.InviteToUser, invite.UnionID,
		invite.UnionName, invite.ChairpersonID, invite.ChairpersonName, invite.ChairpersonLevel,
		invite.UnionLevel, invite.CreateTime, invite.Status)
	
	if err != nil {
		return fmt.Errorf("创建工会邀请失败: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取新建邀请ID失败: %w", err)
	}
	
	invite.ID = int(id)
	return nil
}

// Update updates union invite information
func (r *mysqlUnionInviteRepository) Update(invite *entity.UnionInvite) error {
	query := `UPDATE inviterequests SET invitefromuser = ?, invitetouser = ?, unionid = ?, 
			  unionname = ?, chairpersonid = ?, chairpersonname = ?, chairpersonlevel = ?, 
			  unionlevel = ?, status = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, invite.InviteFromUser, invite.InviteToUser, invite.UnionID,
		invite.UnionName, invite.ChairpersonID, invite.ChairpersonName, invite.ChairpersonLevel,
		invite.UnionLevel, invite.Status, invite.ID)
	
	if err != nil {
		return fmt.Errorf("更新工会邀请信息失败: %w", err)
	}
	return nil
}

// Delete deletes a union invite
func (r *mysqlUnionInviteRepository) Delete(id int) error {
	query := "DELETE FROM inviterequests WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除工会邀请失败: %w", err)
	}
	return nil
}

// HasPendingInvite checks if there's a pending invite between users for a union
func (r *mysqlUnionInviteRepository) HasPendingInvite(fromUserName, toUserName string, unionID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM inviterequests WHERE invitefromuser = ? AND invitetouser = ? AND unionid = ? AND status = 'pending'"
	err := r.db.QueryRow(query, fromUserName, toUserName, unionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查是否有待处理邀请失败: %w", err)
	}
	return count > 0, nil
}

// ProcessInvite processes a union invite (accept/reject)
func (r *mysqlUnionInviteRepository) ProcessInvite(inviteID int, status string) error {
	query := "UPDATE inviterequests SET status = ? WHERE id = ?"
	result, err := r.db.Exec(query, status, inviteID)
	if err != nil {
		return fmt.Errorf("处理工会邀请失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查处理结果失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("未找到要处理的邀请记录")
	}
	
	return nil
}