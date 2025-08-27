package service

import (
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"fmt"
	"log"
	"time"
)

// ChatCleanupService 聊天清理服务
type ChatCleanupService struct {
	privateRepo repository.PrivateMessageRepository
	unionRepo   repository.UnionChatRepository
}

// NewChatCleanupService 创建聊天清理服务
func NewChatCleanupService(
	privateRepo repository.PrivateMessageRepository,
	unionRepo repository.UnionChatRepository,
) *ChatCleanupService {
	return &ChatCleanupService{
		privateRepo: privateRepo,
		unionRepo:   unionRepo,
	}
}

// StartCleanupScheduler 启动清理调度器
func (s *ChatCleanupService) StartCleanupScheduler() {
	// 每天凌晨2点执行清理任务
	go s.scheduleCleanup()
	
	// 每月1号创建新的工会聊天表
	go s.scheduleMonthlyTableCreation()
	
	log.Println("聊天清理调度器已启动")
}

// scheduleCleanup 调度清理任务
func (s *ChatCleanupService) scheduleCleanup() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// 立即执行一次清理（用于测试）
	s.executeCleanup()

	for {
		select {
		case <-ticker.C:
			// 检查是否是凌晨2点
			now := time.Now()
			if now.Hour() == 2 {
				s.executeCleanup()
			}
		}
	}
}

// scheduleMonthlyTableCreation 调度月度表创建任务
func (s *ChatCleanupService) scheduleMonthlyTableCreation() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// 立即检查是否需要创建当月表
	s.ensureCurrentMonthTable()

	for {
		select {
		case <-ticker.C:
			// 每天检查是否需要创建新月份的表
			now := time.Now()
			if now.Day() == 1 && now.Hour() == 1 {
				s.ensureCurrentMonthTable()
			}
		}
	}
}

// executeCleanup 执行清理任务
func (s *ChatCleanupService) executeCleanup() {
	log.Println("开始执行聊天数据清理任务")

	// 清理已读的私聊消息（保留7天）
	if err := s.cleanupReadPrivateMessages(7); err != nil {
		log.Printf("清理已读私聊消息失败: %v", err)
	}

	// 清理过期的未读私聊消息（保留30天）
	if err := s.cleanupOldUnreadMessages(30); err != nil {
		log.Printf("清理过期未读消息失败: %v", err)
	}

	// 停用过期的工会聊天表（保留6个月）
	if err := s.deactivateOldUnionTables(6); err != nil {
		log.Printf("停用过期工会聊天表失败: %v", err)
	}

	log.Println("聊天数据清理任务完成")
}

// cleanupReadPrivateMessages 清理已读的私聊消息
func (s *ChatCleanupService) cleanupReadPrivateMessages(retentionDays int) error {
	// 计算删除阈值时间
	threshold := time.Now().AddDate(0, 0, -retentionDays)
	beforeTime := threshold.Format("2006-01-02 15:04:05")

	// 删除已读消息
	deletedCount, err := s.privateRepo.DeleteReadMessages(beforeTime)
	if err != nil {
		return fmt.Errorf("删除已读消息失败: %v", err)
	}

	log.Printf("已清理 %d 条已读私聊消息（%d天前）", deletedCount, retentionDays)
	return nil
}

// cleanupOldUnreadMessages 清理过期的未读私聊消息
func (s *ChatCleanupService) cleanupOldUnreadMessages(retentionDays int) error {
	// 计算删除阈值时间
	threshold := time.Now().AddDate(0, 0, -retentionDays)
	beforeTime := threshold.Format("2006-01-02 15:04:05")

	// 删除过期未读消息
	deletedCount, err := s.privateRepo.DeleteOldUnreadMessages(beforeTime)
	if err != nil {
		return fmt.Errorf("删除过期未读消息失败: %v", err)
	}

	log.Printf("已清理 %d 条过期未读消息（%d天前）", deletedCount, retentionDays)
	return nil
}

// deactivateOldUnionTables 停用过期的工会聊天表
func (s *ChatCleanupService) deactivateOldUnionTables(retentionMonths int) error {
	// 获取所有表
	tables, err := s.unionRepo.GetAllTables()
	if err != nil {
		return fmt.Errorf("获取工会聊天表列表失败: %v", err)
	}

	// 计算停用阈值时间
	threshold := time.Now().AddDate(0, -retentionMonths, 0)
	thresholdYearMonth := threshold.Format("2006-01")

	deactivatedCount := 0
	for _, table := range tables {
		// 只处理活跃状态且过期的表
		if table.IsActive && table.YearMonth < thresholdYearMonth {
			if err := s.unionRepo.DeactivateTable(table.YearMonth); err != nil {
				log.Printf("停用表 %s 失败: %v", table.TableName, err)
				continue
			}
			deactivatedCount++
			log.Printf("已停用工会聊天表: %s (%s)", table.TableName, table.YearMonth)
		}
	}

	log.Printf("共停用 %d 个工会聊天表（%d个月前）", deactivatedCount, retentionMonths)
	return nil
}

// ensureCurrentMonthTable 确保当前月份的工会聊天表存在
func (s *ChatCleanupService) ensureCurrentMonthTable() {
	currentYearMonth := time.Now().Format("2006-01")

	// 检查当前月份的表是否存在
	_, err := s.unionRepo.GetTableByYearMonth(currentYearMonth)
	if err != nil {
		// 表不存在，需要创建
		if err := s.createMonthlyTable(currentYearMonth); err != nil {
			log.Printf("创建当前月份工会聊天表失败: %v", err)
			return
		}
	}

	// 停用其他月份的活跃表
	if err := s.deactivateOtherTables(currentYearMonth); err != nil {
		log.Printf("停用其他活跃表失败: %v", err)
	}
}

// createMonthlyTable 创建月度工会聊天表
func (s *ChatCleanupService) createMonthlyTable(yearMonth string) error {
	// 生成表名
	tableMonth := yearMonth[:4] + yearMonth[5:] // 将 "2025-08" 转换为 "202508"
	tableName := fmt.Sprintf("union_messages_%s", tableMonth)

	// 检查物理表是否存在
	exists, err := s.unionRepo.TableExists(tableName)
	if err != nil {
		return fmt.Errorf("检查表存在性失败: %v", err)
	}

	// 如果物理表不存在，创建它
	if !exists {
		if err := s.unionRepo.CreateMonthlyTable(tableName); err != nil {
			return fmt.Errorf("创建物理表失败: %v", err)
		}
		log.Printf("已创建工会聊天物理表: %s", tableName)
	}

	// 创建表元数据记录
	tableEntity := &entity.UnionChatTable{
		TableName: tableName,
		YearMonth: yearMonth,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := s.unionRepo.CreateTable(tableEntity); err != nil {
		// 如果记录已存在，更新为活跃状态
		tableEntity.IsActive = true
		if updateErr := s.unionRepo.UpdateTable(tableEntity); updateErr != nil {
			return fmt.Errorf("创建/更新表记录失败: create=%v, update=%v", err, updateErr)
		}
	}

	log.Printf("已创建/激活工会聊天表: %s (%s)", tableName, yearMonth)
	return nil
}

// deactivateOtherTables 停用其他月份的活跃表
func (s *ChatCleanupService) deactivateOtherTables(currentYearMonth string) error {
	// 获取所有活跃表
	tables, err := s.unionRepo.GetAllTables()
	if err != nil {
		return fmt.Errorf("获取表列表失败: %v", err)
	}

	for _, table := range tables {
		// 停用非当前月份的活跃表
		if table.IsActive && table.YearMonth != currentYearMonth {
			if err := s.unionRepo.DeactivateTable(table.YearMonth); err != nil {
				log.Printf("停用表 %s 失败: %v", table.TableName, err)
				continue
			}
			log.Printf("已停用工会聊天表: %s (%s)", table.TableName, table.YearMonth)
		}
	}

	return nil
}

// ForceCleanup 强制执行清理（用于管理员接口）
func (s *ChatCleanupService) ForceCleanup() error {
	log.Println("执行强制清理任务")
	s.executeCleanup()
	return nil
}

// GetCleanupStats 获取清理统计信息
func (s *ChatCleanupService) GetCleanupStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计工会聊天表信息
	tables, err := s.unionRepo.GetAllTables()
	if err != nil {
		return nil, fmt.Errorf("获取工会聊天表失败: %v", err)
	}

	activeTables := 0
	inactiveTables := 0
	for _, table := range tables {
		if table.IsActive {
			activeTables++
		} else {
			inactiveTables++
		}
	}

	stats["union_tables"] = map[string]interface{}{
		"total":    len(tables),
		"active":   activeTables,
		"inactive": inactiveTables,
	}

	// 添加清理配置信息
	stats["cleanup_config"] = map[string]interface{}{
		"read_message_retention_days":      7,
		"unread_message_retention_days":    30,
		"union_table_retention_months":     6,
		"cleanup_schedule":                 "每天凌晨2点",
		"monthly_table_creation_schedule":  "每月1号凌晨1点",
	}

	return stats, nil
}