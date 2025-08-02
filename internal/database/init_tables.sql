-- 好友关系表
CREATE TABLE IF NOT EXISTS friend (
    id INT AUTO_INCREMENT PRIMARY KEY,
    fromuserid INT NOT NULL,
    touserid INT NOT NULL,
    status ENUM('pending', 'accepted', 'blocked') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_friendship (fromuserid, touserid),
    FOREIGN KEY (fromuserid) REFERENCES user(userid) ON DELETE CASCADE,
    FOREIGN KEY (touserid) REFERENCES user(userid) ON DELETE CASCADE
);

-- 好友申请表
CREATE TABLE IF NOT EXISTS friend_request (
    id INT AUTO_INCREMENT PRIMARY KEY,
    fromuserid INT NOT NULL,
    touserid INT NOT NULL,
    message VARCHAR(255) DEFAULT '',
    status ENUM('pending', 'accepted', 'rejected') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_request (fromuserid, touserid),
    FOREIGN KEY (fromuserid) REFERENCES user(userid) ON DELETE CASCADE,
    FOREIGN KEY (touserid) REFERENCES user(userid) ON DELETE CASCADE
);

-- 排行榜表（可以根据不同类型进行排名）
CREATE TABLE IF NOT EXISTS ranking (
    id INT AUTO_INCREMENT PRIMARY KEY,
    userid INT NOT NULL,
    rank_type ENUM('level', 'experience', 'equipment_power') DEFAULT 'level',
    rank_value INT NOT NULL DEFAULT 0,
    rank_position INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_rank_type (userid, rank_type),
    FOREIGN KEY (userid) REFERENCES user(userid) ON DELETE CASCADE,
    INDEX idx_rank_type_value (rank_type, rank_value DESC),
    INDEX idx_rank_position (rank_type, rank_position)
);

-- 创建索引以提高查询性能
CREATE INDEX idx_friend_fromuserid ON friend(fromuserid);
CREATE INDEX idx_friend_touserid ON friend(touserid);
CREATE INDEX idx_friend_request_touserid ON friend_request(touserid);
CREATE INDEX idx_ranking_type_value ON ranking(rank_type, rank_value DESC);