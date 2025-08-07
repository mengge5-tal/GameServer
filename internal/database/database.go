package database

import (
	"GameServer/internal/infrastructure/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDatabase(cfg *config.Config) (*sql.DB, error) {
	connectionString := cfg.GetConnectionString()

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// 预热连接池
	if err := WarmupConnectionPool(db, cfg.Database.MaxIdleConns); err != nil {
		log.Printf("Warning: Failed to warmup connection pool: %v", err)
		// 不中断启动流程，只记录警告
	}

	log.Printf("Successfully connected to database at %s:%s", cfg.Database.Host, cfg.Database.Port)
	log.Printf("Connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%s, MaxIdleTime=%s",
		cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime, cfg.Database.ConnMaxIdleTime)
	return db, nil
}

func CheckDatabaseTables(db *sql.DB) error {
	tables := []string{"user", "equip", "playerinfo", "friend", "friend_request", "ranking"}

	for _, table := range tables {
		query := fmt.Sprintf("SHOW TABLES LIKE '%s'", table)
		var tableName string
		err := db.QueryRow(query).Scan(&tableName)
		if err == sql.ErrNoRows {
			log.Printf("Warning: Table '%s' does not exist", table)
		} else if err != nil {
			return fmt.Errorf("error checking table '%s': %v", table, err)
		} else {
			log.Printf("Table '%s' exists", table)
		}
	}

	return nil
}

func CreateMissingTables(db *sql.DB) error {
	log.Println("Creating missing tables...")

	// 用户装备表
	userEquipTable := `
	CREATE TABLE IF NOT EXISTS user_equip (
		id INT AUTO_INCREMENT PRIMARY KEY,
		userid INT NOT NULL,
		equip_slot ENUM('衣服', '鞋子', '戒指', '项链', '头盔', '手套') NOT NULL,
		equipid INT NULL,
		FOREIGN KEY (userid) REFERENCES user(userid) ON DELETE CASCADE,
		FOREIGN KEY (equipid) REFERENCES equip(equipid) ON DELETE SET NULL,
		UNIQUE KEY unique_user_slot (userid, equip_slot)
	)`

	if _, err := db.Exec(userEquipTable); err != nil {
		return fmt.Errorf("failed to create user_equip table: %v", err)
	}
	log.Println("User_equip table created/verified")

	// 好友关系表
	friendTable := `
	CREATE TABLE IF NOT EXISTS friend (
		id INT AUTO_INCREMENT PRIMARY KEY,
		fromuserid INT NOT NULL,
		touserid INT NOT NULL,
		status ENUM('pending', 'accepted', 'blocked') DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_friendship (fromuserid, touserid)
	)`

	if _, err := db.Exec(friendTable); err != nil {
		return fmt.Errorf("failed to create friend table: %v", err)
	}
	log.Println("Friend table created/verified")

	// 好友申请表
	friendRequestTable := `
	CREATE TABLE IF NOT EXISTS friend_request (
		id INT AUTO_INCREMENT PRIMARY KEY,
		fromuserid INT NOT NULL,
		touserid INT NOT NULL,
		message VARCHAR(255) DEFAULT '',
		status ENUM('pending', 'accepted', 'rejected') DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_request (fromuserid, touserid)
	)`

	if _, err := db.Exec(friendRequestTable); err != nil {
		return fmt.Errorf("failed to create friend_request table: %v", err)
	}
	log.Println("Friend_request table created/verified")

	// 排行榜表
	rankingTable := `
	CREATE TABLE IF NOT EXISTS ranking (
		id INT AUTO_INCREMENT PRIMARY KEY,
		userid INT NOT NULL,
		rank_type ENUM('level', 'experience', 'equipment_power') DEFAULT 'level',
		rank_value INT NOT NULL DEFAULT 0,
		rank_position INT NOT NULL DEFAULT 0,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_user_rank_type (userid, rank_type)
	)`

	if _, err := db.Exec(rankingTable); err != nil {
		return fmt.Errorf("failed to create ranking table: %v", err)
	}
	log.Println("Ranking table created/verified")

	// 创建索引 - 使用兼容的语法
	indexes := []struct {
		name string
		sql  string
	}{
		{"idx_friend_fromuserid", "CREATE INDEX idx_friend_fromuserid ON friend(fromuserid)"},
		{"idx_friend_touserid", "CREATE INDEX idx_friend_touserid ON friend(touserid)"},
		{"idx_friend_request_touserid", "CREATE INDEX idx_friend_request_touserid ON friend_request(touserid)"},
		{"idx_ranking_type_value", "CREATE INDEX idx_ranking_type_value ON ranking(rank_type, rank_value DESC)"},
	}

	for _, index := range indexes {
		// 检查索引是否存在
		var indexExists bool
		err := db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND index_name = ?",
			index.name).Scan(&indexExists)

		if err != nil || !indexExists {
			if _, err := db.Exec(index.sql); err != nil {
				log.Printf("Warning: failed to create index %s: %v", index.name, err)
			} else {
				log.Printf("Index %s created successfully", index.name)
			}
		} else {
			log.Printf("Index %s already exists", index.name)
		}
	}

	// 修复用户表的自动递增问题
	_, err := db.Exec("ALTER TABLE user MODIFY COLUMN userid INT AUTO_INCREMENT")
	if err != nil {
		log.Printf("Warning: failed to modify user table userid field: %v", err)
	} else {
		log.Println("User table userid field set to AUTO_INCREMENT")
	}

	// 修复密码字段长度以支持bcrypt哈希（需要60个字符）
	_, err = db.Exec("ALTER TABLE user MODIFY COLUMN password VARCHAR(255)")
	if err != nil {
		log.Printf("Warning: failed to modify user table password field: %v", err)
	} else {
		log.Println("User table password field extended to VARCHAR(255)")
	}

	// 检查并添加在线状态字段
	var onlineStatusExists bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'user' AND column_name = 'online_status'").Scan(&onlineStatusExists)

	if err != nil {
		log.Printf("Warning: failed to check online_status column existence: %v", err)
	} else if !onlineStatusExists {
		_, err = db.Exec("ALTER TABLE user ADD COLUMN online_status INT DEFAULT 0")
		if err != nil {
			log.Printf("Warning: failed to add online_status column to user table: %v", err)
		} else {
			log.Println("Added online_status column to user table")
		}
	} else {
		log.Println("online_status column already exists in user table")
	}

	// 检查并添加装备表的字段
	equipColumns := []struct {
		name         string
		definition   string
		defaultValue string
	}{
		{"type", "INT DEFAULT 1", "1"},
		{"suitid", "INT DEFAULT 0", "0"},
		{"suitname", "VARCHAR(255) DEFAULT ''", "''"},
		{"equip_type_id", "INT DEFAULT 0", "0"},
		{"equip_type_name", "VARCHAR(255) DEFAULT ''", "''"},
	}

	for _, column := range equipColumns {
		var columnExists bool
		err = db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'equip' AND column_name = ?",
			column.name).Scan(&columnExists)

		if err != nil {
			log.Printf("Warning: failed to check %s column existence: %v", column.name, err)
		} else if !columnExists {
			_, err = db.Exec(fmt.Sprintf("ALTER TABLE equip ADD COLUMN %s %s", column.name, column.definition))
			if err != nil {
				log.Printf("Warning: failed to add %s column to equip table: %v", column.name, err)
			} else {
				log.Printf("Added %s column to equip table", column.name)
			}
		} else {
			log.Printf("%s column already exists in equip table", column.name)
		}
	}

	// 检查并移除已弃用的equipname字段
	var equipnameExists bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'equip' AND column_name = 'equipname'").Scan(&equipnameExists)
	if err != nil {
		log.Printf("Warning: failed to check equipname column existence: %v", err)
	} else if equipnameExists {
		log.Println("Warning: equipname column still exists in equip table, consider removing it manually")
		// 不自动删除列，让用户手动处理，以防数据丢失
	}

	// 检查并修改equipid字段为自增
	var isAutoIncrement bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'equip' AND column_name = 'equipid' AND extra LIKE '%auto_increment%'").Scan(&isAutoIncrement)
	if err != nil {
		log.Printf("Warning: failed to check equipid auto_increment: %v", err)
	} else if !isAutoIncrement {
		log.Println("Warning: equipid is not auto_increment, consider modifying it manually: ALTER TABLE equip MODIFY equipid INT AUTO_INCREMENT PRIMARY KEY")
	}

	log.Println("All missing tables and indexes created successfully")
	return nil
}

func CheckTableStructure(db *sql.DB) error {
	// 检查user表结构
	log.Println("Checking user table structure...")
	rows, err := db.Query("DESCRIBE user")
	if err != nil {
		log.Printf("Error describing user table: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			log.Printf("user table - Field: %s, Type: %s, Key: %s",
				field.String, fieldType.String, key.String)
		}
	}

	// 检查equip表结构
	log.Println("Checking equip table structure...")
	rows, err = db.Query("DESCRIBE equip")
	if err != nil {
		log.Printf("Error describing equip table: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			log.Printf("equip table - Field: %s, Type: %s, Key: %s",
				field.String, fieldType.String, key.String)
		}
	}

	// 检查playerinfo表结构
	log.Println("Checking playerinfo table structure...")
	rows, err = db.Query("DESCRIBE playerinfo")
	if err != nil {
		log.Printf("Error describing playerinfo table: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			log.Printf("playerinfo table - Field: %s, Type: %s, Key: %s",
				field.String, fieldType.String, key.String)
		}
	}

	return nil
}

// WarmupConnectionPool 预热数据库连接池
func WarmupConnectionPool(db *sql.DB, targetConnections int) error {
	log.Printf("Warming up connection pool with %d connections...", targetConnections)

	// 创建多个并发连接来预热连接池
	done := make(chan bool, targetConnections)
	errors := make(chan error, targetConnections)

	for i := 0; i < targetConnections; i++ {
		go func(connNum int) {
			// 执行一个简单的查询来建立连接
			var result int
			err := db.QueryRow("SELECT 1").Scan(&result)
			if err != nil {
				log.Printf("Failed to warmup connection %d: %v", connNum, err)
				errors <- err
			} else {
				log.Printf("Connection %d warmed up successfully", connNum)
				done <- true
			}
		}(i)
	}

	// 等待所有连接完成或失败
	successCount := 0
	errorCount := 0
	for i := 0; i < targetConnections; i++ {
		select {
		case <-done:
			successCount++
		case <-errors:
			errorCount++
		}
	}

	log.Printf("Connection pool warmup completed: %d successful, %d failed", successCount, errorCount)

	// 获取连接池统计信息
	stats := db.Stats()
	log.Printf("Connection pool stats after warmup: Open=%d, InUse=%d, Idle=%d",
		stats.OpenConnections, stats.InUse, stats.Idle)

	if errorCount > targetConnections/2 {
		return fmt.Errorf("too many connection failures during warmup: %d/%d", errorCount, targetConnections)
	}

	return nil
}

// GetConnectionPoolStats 获取连接池统计信息
func GetConnectionPoolStats(db *sql.DB) map[string]interface{} {
	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
