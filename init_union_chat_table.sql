-- 初始化工会聊天分表
-- 如果union_chat_tables表存在但没有数据，插入当前月份的分表记录

-- 检查并插入当前月份的分表记录
SET @current_year_month = DATE_FORMAT(NOW(), '%Y-%m');
SET @current_table_name = CONCAT('union_chat_', REPLACE(@current_year_month, '-', '_'));

-- 插入分表管理记录
INSERT IGNORE INTO union_chat_tables (table_name, year_month, is_active) 
VALUES (@current_table_name, @current_year_month, 1);

-- 创建当前月份的实际聊天表
SET @create_sql = CONCAT(
    'CREATE TABLE IF NOT EXISTS ', @current_table_name, ' (',
    'id BIGINT AUTO_INCREMENT PRIMARY KEY,',
    'union_id INT NOT NULL,',
    'user_id INT NOT NULL,',
    'username VARCHAR(50) NOT NULL,',
    'content TEXT NOT NULL,',
    'created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,',
    'updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,',
    'INDEX idx_union_id_created (union_id, created_at DESC),',
    'INDEX idx_user_id (user_id),',
    'INDEX idx_created_at (created_at)',
    ') ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci'
);

PREPARE stmt FROM @create_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 验证数据
SELECT 'union_chat_tables内容:' as info;
SELECT * FROM union_chat_tables ORDER BY created_at DESC;

SELECT CONCAT('当前活跃表: ', @current_table_name) as active_table;