-- MySQL schema for gaojia-backend
-- 与 SQLite schema 结构一致，便于切换

CREATE TABLE IF NOT EXISTS ag_chapter (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  chapter_id BIGINT NOT NULL,
  subject_code VARCHAR(32) NOT NULL DEFAULT '',
  parent_chapter_id BIGINT NOT NULL DEFAULT 0,
  chapter_level INT NOT NULL,
  chapter_name VARCHAR(255) NOT NULL,
  sort_no INT NOT NULL DEFAULT 0,
  all_question_num INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_chapter_id (chapter_id),
  KEY idx_chapter_parent (subject_code, parent_chapter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ag_question (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  question_id BIGINT NOT NULL,
  title_html MEDIUMTEXT NOT NULL,
  show_type_name VARCHAR(64) NOT NULL DEFAULT '',
  knowledge VARCHAR(255) NOT NULL DEFAULT '',
  analyze_text MEDIUMTEXT,
  material_text MEDIUMTEXT,
  answer_json JSON,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_question_id (question_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ag_question_option (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  question_id BIGINT NOT NULL,
  option_no INT NOT NULL,
  option_label VARCHAR(8) NOT NULL DEFAULT '',
  option_html TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_question_option (question_id, option_no),
  KEY idx_option_question (question_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ag_chapter_question (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  subject_code VARCHAR(32) NOT NULL DEFAULT '',
  chapter_id BIGINT NOT NULL,
  section_chapter_id BIGINT NOT NULL DEFAULT 0,
  question_id BIGINT NOT NULL,
  question_index INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_chapter_question (chapter_id, question_id),
  KEY idx_cq_question (question_id),
  KEY idx_cq_chapter (chapter_id),
  KEY idx_section_chapter (section_chapter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS app_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  display_name VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_question_mark (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  question_id BIGINT NOT NULL,
  favorite TINYINT NOT NULL DEFAULT 0,
  difficult TINYINT NOT NULL DEFAULT 0,
  note VARCHAR(1000) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_question (user_id, question_id),
  KEY idx_uqm_user_favorite (user_id, favorite),
  KEY idx_uqm_user_difficult (user_id, difficult),
  KEY idx_uqm_question (question_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_answer_record (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  chapter_id BIGINT DEFAULT NULL,
  question_id BIGINT NOT NULL,
  selected_answer_json JSON,
  correct_answer_json JSON,
  is_correct TINYINT NOT NULL DEFAULT 0,
  duration_seconds INT DEFAULT NULL,
  answered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_uar_user_question (user_id, question_id),
  KEY idx_uar_user_chapter (user_id, chapter_id),
  KEY idx_uar_wrong (user_id, is_correct, answered_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO app_user (id, username, display_name) VALUES (1, 'local_user', '本地用户');
