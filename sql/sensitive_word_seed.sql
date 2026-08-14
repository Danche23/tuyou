-- 敏感词样例 seed（Phase 1 演示 + 冒烟）
-- 导入命令需显式指定 utf8mb4（Windows 下 mysql 客户端默认字符集可能不是 utf8mb4，会导致中文报 Data too long）：
--   mysql --default-character-set=utf8mb4 -h<host> -P<port> -uroot -p<pass> <db> < sensitive_word_seed.sql
-- 建表语句与 internal/model/entity/sensitive_word.go 的 AutoMigrate 字段保持一致（幂等，可重复执行）。
-- 全量词库：从 GitHub 开源中文敏感词库下载后按规则清洗，用 INSERT ... VALUES 追加到下方，本地导入、不提交仓库。
-- 清洗规则：去空行 → 去重 → TrimSpace → 过滤单字符/超长词 → 生成 VALUES 追加。

CREATE TABLE IF NOT EXISTS `sensitive_word` (
  `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
  `word`        varchar(100) NOT NULL,
  `status`      tinyint DEFAULT '1',
  `create_time` datetime(3) NULL,
  `update_time` datetime(3) NULL,
  `delete_time` datetime(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sensitive_word_word` (`word`),
  KEY `idx_sensitive_word_status` (`status`),
  KEY `idx_sensitive_word_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO sensitive_word (word, status, create_time, update_time, delete_time) VALUES
('赌博', 1, NOW(), NOW(), NULL),
('毒品', 1, NOW(), NOW(), NULL),
('色情', 1, NOW(), NOW(), NULL),
('诈骗', 1, NOW(), NOW(), NULL),
('传销', 1, NOW(), NOW(), NULL),
('高利贷', 1, NOW(), NOW(), NULL),
('代开发票', 1, NOW(), NOW(), NULL),
('办证', 1, NOW(), NOW(), NULL),
('博彩', 1, NOW(), NOW(), NULL),
('刷单', 1, NOW(), NOW(), NULL);
