// 从 MySQL 迁移数据到 SQLite
// 用法: MYSQL_DSN="user:pass@tcp(127.0.0.1:3306)/ruankao_gaojia" go run ./cmd/migrate -db gaojia.db
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ruankao/gaojia-backend-go/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "ruankao_user:Rk8!vN3#qL7@xP2$hT9^mZ4&cW6@tcp(127.0.0.1:3306)/ruankao_gaojia?charset=utf8mb4&parseTime=true"
		log.Printf("using default MYSQL_DSN (set MYSQL_DSN env to override)")
	}

	dbPath := flag.String("db", "gaojia.db", "SQLite output path")
	flag.Parse()

	mysqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer mysqlDB.Close()
	if err := mysqlDB.Ping(); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}
	log.Println("connected to MySQL")

	sqliteDB, err := db.OpenSQLite(*dbPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer sqliteDB.Close()
	log.Println("SQLite db ready")

	tx, err := sqliteDB.Begin()
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	// 清空目标表（保留 schema）
	for _, t := range []string{"user_answer_record", "user_question_mark", "app_user", "ag_chapter_question", "ag_question_option", "ag_question", "ag_chapter"} {
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", t)); err != nil {
			log.Printf("clear %s: %v (may not exist yet)", t, err)
		}
	}

	// 迁移 ag_chapter (仅必要列)
	if err := migrateChapters(mysqlDB, tx); err != nil {
		log.Fatalf("migrate chapters: %v", err)
	}
	log.Println("migrated ag_chapter")

	// 迁移 ag_question (仅必要列)
	if err := migrateQuestions(mysqlDB, tx); err != nil {
		log.Fatalf("migrate questions: %v", err)
	}
	log.Println("migrated ag_question")

	// 迁移 ag_question_option
	if err := migrateQuestionOptions(mysqlDB, tx); err != nil {
		log.Fatalf("migrate question_options: %v", err)
	}
	log.Println("migrated ag_question_option")

	// 迁移 ag_chapter_question
	if err := migrateChapterQuestions(mysqlDB, tx); err != nil {
		log.Fatalf("migrate chapter_questions: %v", err)
	}
	log.Println("migrated ag_chapter_question")

	// 迁移 app_user
	if err := migrateAppUser(mysqlDB, tx); err != nil {
		log.Fatalf("migrate app_user: %v", err)
	}
	log.Println("migrated app_user")

	// 迁移 user_question_mark
	if err := migrateUserQuestionMark(mysqlDB, tx); err != nil {
		log.Fatalf("migrate user_question_mark: %v", err)
	}
	log.Println("migrated user_question_mark")

	// 迁移 user_answer_record
	if err := migrateUserAnswerRecord(mysqlDB, tx); err != nil {
		log.Fatalf("migrate user_answer_record: %v", err)
	}
	log.Println("migrated user_answer_record")

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}
	log.Println("migration done")
}

func migrateChapters(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`
		SELECT chapter_id, subject_code, parent_chapter_id, chapter_level, chapter_name, sort_no, all_question_num
		FROM ag_chapter
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`
		INSERT INTO ag_chapter (chapter_id, subject_code, parent_chapter_id, chapter_level, chapter_name, sort_no, all_question_num)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var chapterID, parentID int64
		var subjectCode, chapterName string
		var level, sortNo, allNum int
		if err := rows.Scan(&chapterID, &subjectCode, &parentID, &level, &chapterName, &sortNo, &allNum); err != nil {
			return err
		}
		if _, err := ins.Exec(chapterID, nullStr(subjectCode), parentID, level, chapterName, sortNo, allNum); err != nil {
			return err
		}
	}
	return rows.Err()
}

func migrateQuestions(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`
		SELECT question_id, title_html, show_type_name, knowledge, analyze_text, material_text, answer_json
		FROM ag_question
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`
		INSERT INTO ag_question (question_id, title_html, show_type_name, knowledge, analyze_text, material_text, answer_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var qid int64
		var title, showType, knowledge string
		var analyze, material, answer interface{}
		if err := rows.Scan(&qid, &title, &showType, &knowledge, &analyze, &material, &answer); err != nil {
			return err
		}
		if _, err := ins.Exec(qid, title, nullStr(showType), nullStr(knowledge), toStr(analyze), toStr(material), jsonToStr(answer)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func migrateQuestionOptions(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`
		SELECT question_id, option_no, option_label, option_html
		FROM ag_question_option
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`
		INSERT INTO ag_question_option (question_id, option_no, option_label, option_html)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var qid int64
		var no int
		var label, html string
		if err := rows.Scan(&qid, &no, &label, &html); err != nil {
			return err
		}
		if _, err := ins.Exec(qid, no, nullStr(label), html); err != nil {
			return err
		}
	}
	return rows.Err()
}

func migrateChapterQuestions(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`
		SELECT subject_code, chapter_id, question_id, question_index
		FROM ag_chapter_question
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`
		INSERT INTO ag_chapter_question (subject_code, chapter_id, question_id, question_index)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var subj string
		var chapterID, questionID int64
		var idx int
		if err := rows.Scan(&subj, &chapterID, &questionID, &idx); err != nil {
			return err
		}
		if _, err := ins.Exec(nullStr(subj), chapterID, questionID, idx); err != nil {
			return err
		}
	}
	return rows.Err()
}

func migrateAppUser(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`SELECT id, username, display_name FROM app_user`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`INSERT INTO app_user (id, username, display_name) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var id int64
		var username, display string
		if err := rows.Scan(&id, &username, &display); err != nil {
			return err
		}
		if _, err := ins.Exec(id, username, nullStr(display)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func migrateUserQuestionMark(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`SELECT user_id, question_id, favorite, difficult, note FROM user_question_mark`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`
		INSERT INTO user_question_mark (user_id, question_id, favorite, difficult, note)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var uid int64
		var qid int64
		var fav, diff int
		var note string
		if err := rows.Scan(&uid, &qid, &fav, &diff, &note); err != nil {
			return err
		}
		if _, err := ins.Exec(uid, qid, fav, diff, nullStr(note)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func migrateUserAnswerRecord(mysql *sql.DB, tx *sql.Tx) error {
	rows, err := mysql.Query(`
		SELECT user_id, chapter_id, question_id, selected_answer_json, correct_answer_json, is_correct, duration_seconds, answered_at
		FROM user_answer_record
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	ins, err := tx.Prepare(`
		INSERT INTO user_answer_record (user_id, chapter_id, question_id, selected_answer_json, correct_answer_json, is_correct, duration_seconds, answered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer ins.Close()

	for rows.Next() {
		var uid int64
		var chapterID sql.NullInt64
		var qid int64
		var selJSON, corrJSON interface{}
		var isCorrect int
		var dur interface{}
		var answeredAt string
		if err := rows.Scan(&uid, &chapterID, &qid, &selJSON, &corrJSON, &isCorrect, &dur, &answeredAt); err != nil {
			return err
		}
		var chID interface{}
		if chapterID.Valid {
			chID = chapterID.Int64
		}
		if _, err := ins.Exec(uid, chID, qid, jsonToStr(selJSON), jsonToStr(corrJSON), isCorrect, dur, answeredAt); err != nil {
			return err
		}
	}
	return rows.Err()
}

func nullStr(s string) string {
	if s == "" {
		return ""
	}
	return s
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return fmt.Sprintf("%v", v)
}

func jsonToStr(v interface{}) string {
	if v == nil {
		return "[]"
	}
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	}
	return "[]"
}
