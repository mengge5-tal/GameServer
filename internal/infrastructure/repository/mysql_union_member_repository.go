package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
)

// mysqlUnionMemberRepository implements UnionMemberRepository
type mysqlUnionMemberRepository struct {
	db *sql.DB
}

// NewMySQLUnionMemberRepository creates a new MySQL union member repository
func NewMySQLUnionMemberRepository(db *sql.DB) repository.UnionMemberRepository {
	return &mysqlUnionMemberRepository{db: db}
}

// GetByID retrieves a union member by ID
func (r *mysqlUnionMemberRepository) GetByID(id int) (*entity.UnionMember, error) {
	member := &entity.UnionMember{}
	query := `SELECT id, unionid, unionname, memberid, memberlevel, joined_time, roleid 
			  FROM unionmembers WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(&member.ID, &member.UnionID, &member.UnionName,
		&member.MemberID, &member.MemberLevel, &member.JoinedTime, &member.RoleID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取工会成员信息失败: %w", err)
	}
	return member, nil
}

// GetByUserID retrieves a union member by user ID
func (r *mysqlUnionMemberRepository) GetByUserID(userID int) (*entity.UnionMember, error) {
	member := &entity.UnionMember{}
	query := `SELECT id, unionid, unionname, memberid, memberlevel, joined_time, roleid 
			  FROM unionmembers WHERE memberid = ?`

	err := r.db.QueryRow(query, userID).Scan(&member.ID, &member.UnionID, &member.UnionName,
		&member.MemberID, &member.MemberLevel, &member.JoinedTime, &member.RoleID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("按用户ID获取工会成员信息失败: %w", err)
	}
	return member, nil
}

// GetByUnionID retrieves all members of a union
func (r *mysqlUnionMemberRepository) GetByUnionID(unionID int) ([]*entity.UnionMember, error) {
	query := `SELECT id, unionid, unionname, memberid, memberlevel, joined_time, roleid 
			  FROM unionmembers WHERE unionid = ? ORDER BY roleid DESC, joined_time ASC`

	rows, err := r.db.Query(query, unionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会成员列表失败: %w", err)
	}
	defer rows.Close()

	var members []*entity.UnionMember
	for rows.Next() {
		member := &entity.UnionMember{}
		err := rows.Scan(&member.ID, &member.UnionID, &member.UnionName,
			&member.MemberID, &member.MemberLevel, &member.JoinedTime, &member.RoleID)
		if err != nil {
			return nil, fmt.Errorf("扫描工会成员数据失败: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// Create creates a new union member
func (r *mysqlUnionMemberRepository) Create(member *entity.UnionMember) error {
	query := `INSERT INTO unionmembers (unionid, unionname, memberid, memberlevel, roleid) 
			  VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, member.UnionID, member.UnionName, member.MemberID,
		member.MemberLevel, member.RoleID)

	if err != nil {
		return fmt.Errorf("创建工会成员失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取新建工会成员ID失败: %w", err)
	}

	member.ID = int(id)
	return nil
}

// Update updates union member information
func (r *mysqlUnionMemberRepository) Update(member *entity.UnionMember) error {
	query := `UPDATE unionmembers SET unionid = ?, unionname = ?, memberid = ?, 
			  memberlevel = ?, roleid = ? WHERE id = ?`

	_, err := r.db.Exec(query, member.UnionID, member.UnionName, member.MemberID,
		member.MemberLevel, member.RoleID, member.ID)

	if err != nil {
		return fmt.Errorf("更新工会成员信息失败: %w", err)
	}
	return nil
}

// Delete deletes a union member by ID
func (r *mysqlUnionMemberRepository) Delete(id int) error {
	query := "DELETE FROM unionmembers WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除工会成员失败: %w", err)
	}
	return nil
}

// DeleteByUserID deletes a union member by user ID
func (r *mysqlUnionMemberRepository) DeleteByUserID(userID int) error {
	query := "DELETE FROM unionmembers WHERE memberid = ?"
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("按用户ID删除工会成员失败: %w", err)
	}
	return nil
}

// IsUserInUnion checks if user is in any union
func (r *mysqlUnionMemberRepository) IsUserInUnion(userID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM unionmembers WHERE memberid = ?"
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查用户是否在工会失败: %w", err)
	}
	return count > 0, nil
}

// IsUserInSpecificUnion checks if user is in a specific union
func (r *mysqlUnionMemberRepository) IsUserInSpecificUnion(userID, unionID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM unionmembers WHERE memberid = ? AND unionid = ?"
	err := r.db.QueryRow(query, userID, unionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查用户是否在指定工会失败: %w", err)
	}
	return count > 0, nil
}

// GetMemberRole gets user's role in a specific union
func (r *mysqlUnionMemberRepository) GetMemberRole(userID, unionID int) (int, error) {
	var roleID int
	query := "SELECT roleid FROM unionmembers WHERE memberid = ? AND unionid = ?"
	err := r.db.QueryRow(query, userID, unionID).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil // User not in this union
		}
		return -1, fmt.Errorf("获取成员角色失败: %w", err)
	}
	return roleID, nil
}

// UpdateRole updates member's role
func (r *mysqlUnionMemberRepository) UpdateRole(userID, unionID, roleID int) error {
	query := "UPDATE unionmembers SET roleid = ? WHERE memberid = ? AND unionid = ?"
	result, err := r.db.Exec(query, roleID, userID, unionID)
	if err != nil {
		return fmt.Errorf("更新成员角色失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查更新结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("未找到要更新的成员记录")
	}

	return nil
}

// GetMemberCount gets the number of members in a union
func (r *mysqlUnionMemberRepository) GetMemberCount(unionID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM unionmembers WHERE unionid = ?"
	err := r.db.QueryRow(query, unionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取工会成员数量失败: %w", err)
	}
	return count, nil
}

// GetMembersByUnionIDWithPagination retrieves union members with pagination
func (r *mysqlUnionMemberRepository) GetMembersByUnionIDWithPagination(unionID, page, limit int) ([]*entity.UnionMember, int, error) {
	// 首先获取总数
	total, err := r.GetMemberCount(unionID)
	if err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取分页数据，包括用户基本信息
	query := `SELECT um.id, um.unionid, um.unionname, um.memberid, um.memberlevel, 
	                 um.joined_time, um.roleid, u.username, COALESCE(p.experience, 0) as experience
			  FROM unionmembers um
			  LEFT JOIN user u ON um.memberid = u.userid
			  LEFT JOIN playerinfo p ON u.userid = p.userid
			  WHERE um.unionid = ? 
			  ORDER BY um.roleid DESC, um.joined_time ASC
			  LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, unionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("获取工会成员分页列表失败: %w", err)
	}
	defer rows.Close()

	var members []*entity.UnionMember
	for rows.Next() {
		member := &entity.UnionMember{}
		var username sql.NullString
		var userExp sql.NullInt64

		err := rows.Scan(&member.ID, &member.UnionID, &member.UnionName,
			&member.MemberID, &member.MemberLevel, &member.JoinedTime, &member.RoleID,
			&username, &userExp)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描工会成员数据失败: %w", err)
		}

		// 设置用户名和经验值等信息
		if username.Valid {
			member.MemberName = username.String
		}
		if userExp.Valid {
			member.UserExperience = int(userExp.Int64)
		}
		// 设置默认的最后登录时间为空字符串
		member.LastLogin = ""

		members = append(members, member)
	}

	return members, total, nil
}
