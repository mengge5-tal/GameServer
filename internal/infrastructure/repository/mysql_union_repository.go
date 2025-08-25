package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
)

// mysqlUnionRepository implements UnionRepository
type mysqlUnionRepository struct {
	db *sql.DB
}

// NewMySQLUnionRepository creates a new MySQL union repository
func NewMySQLUnionRepository(db *sql.DB) repository.UnionRepository {
	return &mysqlUnionRepository{db: db}
}

// GetByID retrieves a union by ID
func (r *mysqlUnionRepository) GetByID(unionID int) (*entity.Union, error) {
	union := &entity.Union{}
	query := "SELECT unionid, unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, created_time, union_desc FROM `union` WHERE unionid = ?"
	
	err := r.db.QueryRow(query, unionID).Scan(
		&union.UnionID, &union.UnionName, &union.ChairpersonID, &union.ChairpersonName,
		&union.ChairpersonLevel, &union.UnionLevel, &union.UnionMembers, &union.Experience,
		&union.CreatedTime, &union.UnionDesc,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取工会信息失败: %w", err)
	}
	return union, nil
}

// GetByName retrieves a union by name
func (r *mysqlUnionRepository) GetByName(unionName string) (*entity.Union, error) {
	union := &entity.Union{}
	query := "SELECT unionid, unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, created_time, union_desc FROM `union` WHERE unionname = ?"
	
	err := r.db.QueryRow(query, unionName).Scan(
		&union.UnionID, &union.UnionName, &union.ChairpersonID, &union.ChairpersonName,
		&union.ChairpersonLevel, &union.UnionLevel, &union.UnionMembers, &union.Experience,
		&union.CreatedTime, &union.UnionDesc,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("按名称获取工会信息失败: %w", err)
	}
	return union, nil
}

// Create creates a new union
func (r *mysqlUnionRepository) Create(union *entity.Union) error {
	query := "INSERT INTO `union` (unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, union_desc) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	
	result, err := r.db.Exec(query, union.UnionName, union.ChairpersonID, union.ChairpersonName,
		union.ChairpersonLevel, union.UnionLevel, union.UnionMembers, union.Experience, union.UnionDesc)
	
	if err != nil {
		return fmt.Errorf("创建工会失败: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取新建工会ID失败: %w", err)
	}
	
	union.UnionID = int(id)
	return nil
}

// Update updates union information
func (r *mysqlUnionRepository) Update(union *entity.Union) error {
	query := "UPDATE `union` SET unionname = ?, chairpersonid = ?, chairpersonname = ?, chairpersonlevel = ?, unionlevel = ?, unionmembers = ?, experience = ?, union_desc = ? WHERE unionid = ?"
	
	_, err := r.db.Exec(query, union.UnionName, union.ChairpersonID, union.ChairpersonName,
		union.ChairpersonLevel, union.UnionLevel, union.UnionMembers, union.Experience,
		union.UnionDesc, union.UnionID)
	
	if err != nil {
		return fmt.Errorf("更新工会信息失败: %w", err)
	}
	return nil
}

// Delete deletes a union
func (r *mysqlUnionRepository) Delete(unionID int) error {
	query := "DELETE FROM `union` WHERE unionid = ?"
	_, err := r.db.Exec(query, unionID)
	if err != nil {
		return fmt.Errorf("删除工会失败: %w", err)
	}
	return nil
}

// Exists checks if union name already exists
func (r *mysqlUnionRepository) Exists(unionName string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM `union` WHERE unionname = ?"
	err := r.db.QueryRow(query, unionName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查工会名称是否存在失败: %w", err)
	}
	return count > 0, nil
}

// GetAll retrieves all unions with pagination
func (r *mysqlUnionRepository) GetAll(limit, offset int) ([]*entity.Union, int, error) {
	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM `union`"
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("获取工会总数失败: %w", err)
	}
	
	// Get unions with pagination
	query := "SELECT unionid, unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, created_time, union_desc FROM `union` ORDER BY unionlevel DESC, experience DESC LIMIT ? OFFSET ?"
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询工会列表失败: %w", err)
	}
	defer rows.Close()
	
	var unions []*entity.Union
	for rows.Next() {
		union := &entity.Union{}
		err := rows.Scan(&union.UnionID, &union.UnionName, &union.ChairpersonID, 
			&union.ChairpersonName, &union.ChairpersonLevel, &union.UnionLevel,
			&union.UnionMembers, &union.Experience, &union.CreatedTime, &union.UnionDesc)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描工会数据失败: %w", err)
		}
		unions = append(unions, union)
	}
	
	return unions, total, nil
}

// GetRecommended retrieves recommended unions
func (r *mysqlUnionRepository) GetRecommended(limit int) ([]*entity.Union, error) {
	query := "SELECT unionid, unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, created_time, union_desc FROM `union` WHERE unionmembers < 50 ORDER BY RAND() LIMIT ?"
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("获取推荐工会失败: %w", err)
	}
	defer rows.Close()
	
	var unions []*entity.Union
	for rows.Next() {
		union := &entity.Union{}
		err := rows.Scan(&union.UnionID, &union.UnionName, &union.ChairpersonID, 
			&union.ChairpersonName, &union.ChairpersonLevel, &union.UnionLevel,
			&union.UnionMembers, &union.Experience, &union.CreatedTime, &union.UnionDesc)
		if err != nil {
			return nil, fmt.Errorf("扫描推荐工会数据失败: %w", err)
		}
		unions = append(unions, union)
	}
	
	return unions, nil
}

// Search searches unions by keyword
func (r *mysqlUnionRepository) Search(keyword string, limit, offset int) ([]*entity.Union, int, error) {
	searchPattern := "%" + keyword + "%"
	
	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM `union` WHERE unionname LIKE ? OR chairpersonname LIKE ?"
	err := r.db.QueryRow(countQuery, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索工会总数失败: %w", err)
	}
	
	// Search unions
	query := "SELECT unionid, unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, created_time, union_desc FROM `union` WHERE unionname LIKE ? OR chairpersonname LIKE ? ORDER BY unionlevel DESC, experience DESC LIMIT ? OFFSET ?"
	
	rows, err := r.db.Query(query, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("搜索工会失败: %w", err)
	}
	defer rows.Close()
	
	var unions []*entity.Union
	for rows.Next() {
		union := &entity.Union{}
		err := rows.Scan(&union.UnionID, &union.UnionName, &union.ChairpersonID, 
			&union.ChairpersonName, &union.ChairpersonLevel, &union.UnionLevel,
			&union.UnionMembers, &union.Experience, &union.CreatedTime, &union.UnionDesc)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描搜索结果失败: %w", err)
		}
		unions = append(unions, union)
	}
	
	return unions, total, nil
}

// GetRanking retrieves union ranking
func (r *mysqlUnionRepository) GetRanking(limit int) ([]*entity.Union, error) {
	query := "SELECT unionid, unionname, chairpersonid, chairpersonname, chairpersonlevel, unionlevel, unionmembers, experience, created_time, union_desc FROM `union` ORDER BY unionlevel DESC, experience DESC LIMIT ?"
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("获取工会排行失败: %w", err)
	}
	defer rows.Close()
	
	var unions []*entity.Union
	for rows.Next() {
		union := &entity.Union{}
		err := rows.Scan(&union.UnionID, &union.UnionName, &union.ChairpersonID, 
			&union.ChairpersonName, &union.ChairpersonLevel, &union.UnionLevel,
			&union.UnionMembers, &union.Experience, &union.CreatedTime, &union.UnionDesc)
		if err != nil {
			return nil, fmt.Errorf("扫描排行榜数据失败: %w", err)
		}
		unions = append(unions, union)
	}
	
	return unions, nil
}

// GetUnionRank retrieves specific union's rank
func (r *mysqlUnionRepository) GetUnionRank(unionID int) (int, error) {
	query := "SELECT COUNT(*) + 1 as rank FROM `union` u1 WHERE (u1.unionlevel > (SELECT unionlevel FROM `union` WHERE unionid = ?)) OR (u1.unionlevel = (SELECT unionlevel FROM `union` WHERE unionid = ?) AND u1.experience > (SELECT experience FROM `union` WHERE unionid = ?))"
	
	var rank int
	err := r.db.QueryRow(query, unionID, unionID, unionID).Scan(&rank)
	if err != nil {
		return 0, fmt.Errorf("获取工会排名失败: %w", err)
	}
	return rank, nil
}

// UpdateExperience updates union experience
func (r *mysqlUnionRepository) UpdateExperience(unionID, experience int) error {
	query := "UPDATE `union` SET experience = ? WHERE unionid = ?"
	_, err := r.db.Exec(query, experience, unionID)
	if err != nil {
		return fmt.Errorf("更新工会经验失败: %w", err)
	}
	return nil
}

// UpdateLevel updates union level
func (r *mysqlUnionRepository) UpdateLevel(unionID, level int) error {
	query := "UPDATE `union` SET unionlevel = ? WHERE unionid = ?"
	_, err := r.db.Exec(query, level, unionID)
	if err != nil {
		return fmt.Errorf("更新工会等级失败: %w", err)
	}
	return nil
}

// IncrementMemberCount increments union member count
func (r *mysqlUnionRepository) IncrementMemberCount(unionID int) error {
	query := "UPDATE `union` SET unionmembers = unionmembers + 1 WHERE unionid = ?"
	_, err := r.db.Exec(query, unionID)
	if err != nil {
		return fmt.Errorf("增加工会成员数量失败: %w", err)
	}
	return nil
}

// DecrementMemberCount decrements union member count
func (r *mysqlUnionRepository) DecrementMemberCount(unionID int) error {
	query := "UPDATE `union` SET unionmembers = unionmembers - 1 WHERE unionid = ? AND unionmembers > 0"
	_, err := r.db.Exec(query, unionID)
	if err != nil {
		return fmt.Errorf("减少工会成员数量失败: %w", err)
	}
	return nil
}