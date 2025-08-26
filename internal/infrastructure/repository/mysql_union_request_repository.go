package repository

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"database/sql"
	"fmt"
)

// mysqlUnionRequestRepository implements UnionRequestRepository
type mysqlUnionRequestRepository struct {
	db *sql.DB
}

// NewMySQLUnionRequestRepository creates a new MySQL union request repository
func NewMySQLUnionRequestRepository(db *sql.DB) repository.UnionRequestRepository {
	return &mysqlUnionRequestRepository{db: db}
}

// GetByID retrieves a union request by ID
func (r *mysqlUnionRequestRepository) GetByID(id int) (*entity.UnionRequest, error) {
	request := &entity.UnionRequest{}
	query := `SELECT id, unionid, applicantid, applicantname, applicantlevel, 
			  chairpersonid, request_status, request_time 
			  FROM unionrequests WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(&request.ID, &request.UnionID, &request.ApplicantID,
		&request.ApplicantName, &request.ApplicantLevel, &request.ChairpersonID,
		&request.RequestStatus, &request.RequestTime)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("获取工会申请信息失败: %w", err)
	}
	return request, nil
}

// GetByUnionID retrieves all requests for a union
func (r *mysqlUnionRequestRepository) GetByUnionID(unionID int) ([]*entity.UnionRequest, error) {
	query := `SELECT id, unionid, applicantid, applicantname, applicantlevel, 
			  chairpersonid, request_status, request_time 
			  FROM unionrequests WHERE unionid = ? ORDER BY request_time DESC`
	
	rows, err := r.db.Query(query, unionID)
	if err != nil {
		return nil, fmt.Errorf("获取工会申请列表失败: %w", err)
	}
	defer rows.Close()
	
	var requests []*entity.UnionRequest
	for rows.Next() {
		request := &entity.UnionRequest{}
		err := rows.Scan(&request.ID, &request.UnionID, &request.ApplicantID,
			&request.ApplicantName, &request.ApplicantLevel, &request.ChairpersonID,
			&request.RequestStatus, &request.RequestTime)
		if err != nil {
			return nil, fmt.Errorf("扫描工会申请数据失败: %w", err)
		}
		requests = append(requests, request)
	}
	
	return requests, nil
}

// GetByApplicantID retrieves all requests by an applicant
func (r *mysqlUnionRequestRepository) GetByApplicantID(applicantID int) ([]*entity.UnionRequest, error) {
	query := `SELECT id, unionid, applicantid, applicantname, applicantlevel, 
			  chairpersonid, request_status, request_time 
			  FROM unionrequests WHERE applicantid = ? ORDER BY request_time DESC`
	
	rows, err := r.db.Query(query, applicantID)
	if err != nil {
		return nil, fmt.Errorf("获取申请人申请列表失败: %w", err)
	}
	defer rows.Close()
	
	var requests []*entity.UnionRequest
	for rows.Next() {
		request := &entity.UnionRequest{}
		err := rows.Scan(&request.ID, &request.UnionID, &request.ApplicantID,
			&request.ApplicantName, &request.ApplicantLevel, &request.ChairpersonID,
			&request.RequestStatus, &request.RequestTime)
		if err != nil {
			return nil, fmt.Errorf("扫描申请人申请数据失败: %w", err)
		}
		requests = append(requests, request)
	}
	
	return requests, nil
}

// GetPendingByUnionID retrieves pending requests for a union
func (r *mysqlUnionRequestRepository) GetPendingByUnionID(unionID int) ([]*entity.UnionRequest, error) {
	query := `SELECT id, unionid, applicantid, applicantname, applicantlevel, 
			  chairpersonid, request_status, request_time 
			  FROM unionrequests WHERE unionid = ? AND request_status = 1 
			  ORDER BY request_time ASC`
	
	rows, err := r.db.Query(query, unionID)
	if err != nil {
		return nil, fmt.Errorf("获取待处理工会申请失败: %w", err)
	}
	defer rows.Close()
	
	var requests []*entity.UnionRequest
	for rows.Next() {
		request := &entity.UnionRequest{}
		err := rows.Scan(&request.ID, &request.UnionID, &request.ApplicantID,
			&request.ApplicantName, &request.ApplicantLevel, &request.ChairpersonID,
			&request.RequestStatus, &request.RequestTime)
		if err != nil {
			return nil, fmt.Errorf("扫描待处理申请数据失败: %w", err)
		}
		requests = append(requests, request)
	}
	
	return requests, nil
}

// Create creates a new union request
func (r *mysqlUnionRequestRepository) Create(request *entity.UnionRequest) error {
	query := `INSERT INTO unionrequests (unionid, applicantid, applicantname, applicantlevel, 
			  chairpersonid, request_status) VALUES (?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.Exec(query, request.UnionID, request.ApplicantID, request.ApplicantName,
		request.ApplicantLevel, request.ChairpersonID, request.RequestStatus)
	
	if err != nil {
		return fmt.Errorf("创建工会申请失败: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取新建工会申请ID失败: %w", err)
	}
	
	request.ID = int(id)
	return nil
}

// Update updates union request information
func (r *mysqlUnionRequestRepository) Update(request *entity.UnionRequest) error {
	query := `UPDATE unionrequests SET unionid = ?, applicantid = ?, applicantname = ?, 
			  applicantlevel = ?, chairpersonid = ?, request_status = ? WHERE id = ?`
	
	_, err := r.db.Exec(query, request.UnionID, request.ApplicantID, request.ApplicantName,
		request.ApplicantLevel, request.ChairpersonID, request.RequestStatus, request.ID)
	
	if err != nil {
		return fmt.Errorf("更新工会申请信息失败: %w", err)
	}
	return nil
}

// Delete deletes a union request
func (r *mysqlUnionRequestRepository) Delete(id int) error {
	query := "DELETE FROM unionrequests WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("删除工会申请失败: %w", err)
	}
	return nil
}

// HasPendingRequest checks if user has pending request to a union
func (r *mysqlUnionRequestRepository) HasPendingRequest(applicantID, unionID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM unionrequests WHERE applicantid = ? AND unionid = ? AND request_status = 1"
	err := r.db.QueryRow(query, applicantID, unionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查是否有待处理申请失败: %w", err)
	}
	return count > 0, nil
}

// ProcessRequest processes a union request (approve/reject)
func (r *mysqlUnionRequestRepository) ProcessRequest(requestID, status int) error {
	query := "UPDATE unionrequests SET request_status = ? WHERE id = ?"
	result, err := r.db.Exec(query, status, requestID)
	if err != nil {
		return fmt.Errorf("处理工会申请失败: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("检查处理结果失败: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("未找到要处理的申请记录")
	}
	
	return nil
}

// GetByUnionIDAndStatus retrieves requests for a union by status
func (r *mysqlUnionRequestRepository) GetByUnionIDAndStatus(unionID, status int) ([]*entity.UnionRequest, error) {
	query := `SELECT id, unionid, applicantid, applicantname, applicantlevel, 
			  chairpersonid, request_status, request_time 
			  FROM unionrequests WHERE unionid = ? AND request_status = ? 
			  ORDER BY request_time ASC`
	
	rows, err := r.db.Query(query, unionID, status)
	if err != nil {
		return nil, fmt.Errorf("获取工会申请列表失败: %w", err)
	}
	defer rows.Close()
	
	var requests []*entity.UnionRequest
	for rows.Next() {
		request := &entity.UnionRequest{}
		err := rows.Scan(&request.ID, &request.UnionID, &request.ApplicantID,
			&request.ApplicantName, &request.ApplicantLevel, &request.ChairpersonID,
			&request.RequestStatus, &request.RequestTime)
		if err != nil {
			return nil, fmt.Errorf("扫描申请数据失败: %w", err)
		}
		requests = append(requests, request)
	}
	
	return requests, nil
}