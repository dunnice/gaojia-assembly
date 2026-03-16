-- SQLite schema for gaojia-backend
-- 从 MySQL 迁移适配

CREATE TABLE IF NOT EXISTS ag_chapter (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  chapter_id INTEGER NOT NULL UNIQUE,
  subject_code TEXT NOT NULL DEFAULT '',
  parent_chapter_id INTEGER NOT NULL DEFAULT 0,
  chapter_level INTEGER NOT NULL,
  chapter_name TEXT NOT NULL,
  sort_no INTEGER NOT NULL DEFAULT 0,
  all_question_num INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_chapter_parent ON ag_chapter(subject_code, parent_chapter_id);

CREATE TABLE IF NOT EXISTS ag_question (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  question_id INTEGER NOT NULL UNIQUE,
  title_html TEXT NOT NULL,
  show_type_name TEXT NOT NULL DEFAULT '',
  knowledge TEXT NOT NULL DEFAULT '',
  analyze_text TEXT,
  material_text TEXT,
  answer_json TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS ag_question_option (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  question_id INTEGER NOT NULL,
  option_no INTEGER NOT NULL,
  option_label TEXT NOT NULL DEFAULT '',
  option_html TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(question_id, option_no)
);

CREATE INDEX IF NOT EXISTS idx_option_question ON ag_question_option(question_id);

CREATE TABLE IF NOT EXISTS ag_chapter_question (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  subject_code TEXT NOT NULL DEFAULT '',
  chapter_id INTEGER NOT NULL,
  question_id INTEGER NOT NULL,
  question_index INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(chapter_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_cq_question ON ag_chapter_question(question_id);
CREATE INDEX IF NOT EXISTS idx_cq_chapter ON ag_chapter_question(chapter_id);

CREATE TABLE IF NOT EXISTS app_user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS user_question_mark (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  question_id INTEGER NOT NULL,
  favorite INTEGER NOT NULL DEFAULT 0,
  difficult INTEGER NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(user_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_uqm_user_favorite ON user_question_mark(user_id, favorite);
CREATE INDEX IF NOT EXISTS idx_uqm_user_difficult ON user_question_mark(user_id, difficult);
CREATE INDEX IF NOT EXISTS idx_uqm_question ON user_question_mark(question_id);

CREATE TABLE IF NOT EXISTS user_answer_record (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  chapter_id INTEGER,
  question_id INTEGER NOT NULL,
  selected_answer_json TEXT,
  correct_answer_json TEXT,
  is_correct INTEGER NOT NULL DEFAULT 0,
  duration_seconds INTEGER,
  answered_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_uar_user_question ON user_answer_record(user_id, question_id);
CREATE INDEX IF NOT EXISTS idx_uar_user_chapter ON user_answer_record(user_id, chapter_id);
CREATE INDEX IF NOT EXISTS idx_uar_wrong ON user_answer_record(user_id, is_correct, answered_at);

INSERT OR IGNORE INTO app_user (id, username, display_name) VALUES (1, 'local_user', '本地用户');
