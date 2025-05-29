-- 创建数据库
CREATE DATABASE IF NOT EXISTS vote_system CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE vote_system;

-- 创建投票问卷表
CREATE TABLE IF NOT EXISTS polls (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    title VARCHAR(255) NOT NULL COMMENT '投票标题',
    description TEXT COMMENT '投票描述',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否活跃',
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='投票问卷表';

-- 创建选项表
CREATE TABLE IF NOT EXISTS options (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    poll_id BIGINT UNSIGNED NOT NULL COMMENT '投票问卷ID',
    text VARCHAR(255) NOT NULL COMMENT '选项文本',
    vote_count INT DEFAULT 0 COMMENT '票数',
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_poll_id (poll_id),
    FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='投票选项表';

-- 创建投票记录表
CREATE TABLE IF NOT EXISTS votes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    poll_id BIGINT UNSIGNED NOT NULL COMMENT '投票问卷ID',
    option_id BIGINT UNSIGNED NOT NULL COMMENT '选项ID',
    user_ip VARCHAR(45) COMMENT '用户IP地址',
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_poll_id (poll_id),
    INDEX idx_option_id (option_id),
    INDEX idx_user_ip (user_ip),
    -- 暂不启用唯一索引，防止同一ip清除投票记录后无法再次投票
    -- UNIQUE KEY unique_poll_ip (poll_id, user_ip),
    FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE,
    FOREIGN KEY (option_id) REFERENCES options(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='投票记录表';

-- 插入示例数据
INSERT INTO polls (title, description, is_active) VALUES 
('您最喜欢的编程语言是什么？', '请选择您最喜欢的编程语言', TRUE);

-- 获取刚插入的投票ID
SET @poll_id = LAST_INSERT_ID();

-- 插入选项
INSERT INTO options (poll_id, text, vote_count) VALUES 
(@poll_id, 'Go', 0),
(@poll_id, 'Python', 0),
(@poll_id, 'JavaScript', 0),
(@poll_id, 'Java', 0),
(@poll_id, 'TypeScript', 0); 