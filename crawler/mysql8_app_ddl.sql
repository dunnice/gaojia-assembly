USE `ruankao_gaojia`;

CREATE TABLE IF NOT EXISTS `app_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(64) NOT NULL,
  `display_name` VARCHAR(128) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_user_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='本地用户';

CREATE TABLE IF NOT EXISTS `user_question_mark` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `question_id` BIGINT NOT NULL,
  `favorite` TINYINT NOT NULL DEFAULT 0,
  `difficult` TINYINT NOT NULL DEFAULT 0,
  `note` VARCHAR(1000) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_question_mark` (`user_id`, `question_id`),
  KEY `idx_uqm_user_favorite` (`user_id`, `favorite`),
  KEY `idx_uqm_user_difficult` (`user_id`, `difficult`),
  KEY `idx_uqm_question` (`question_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='题目标记';

CREATE TABLE IF NOT EXISTS `user_answer_record` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `chapter_id` BIGINT DEFAULT NULL,
  `question_id` BIGINT NOT NULL,
  `selected_answer_json` JSON DEFAULT NULL,
  `correct_answer_json` JSON DEFAULT NULL,
  `is_correct` TINYINT NOT NULL DEFAULT 0,
  `duration_seconds` INT DEFAULT NULL,
  `answered_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_uar_user_question` (`user_id`, `question_id`),
  KEY `idx_uar_user_chapter` (`user_id`, `chapter_id`),
  KEY `idx_uar_wrong_time` (`user_id`, `is_correct`, `answered_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='答题记录';

INSERT INTO `app_user` (`id`, `username`, `display_name`)
VALUES (1, 'local_user', '本地用户')
ON DUPLICATE KEY UPDATE
  `display_name` = VALUES(`display_name`);
