package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/ruankao/gaojia-backend-go/internal/db"
	"github.com/ruankao/gaojia-backend-go/internal/dto"
)

type QuestionRepository struct {
	db     *sql.DB
	driver db.Driver
}

func NewQuestionRepository(database *sql.DB, driver db.Driver) *QuestionRepository {
	return &QuestionRepository{db: database, driver: driver}
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	s := "?"
	for i := 1; i < n; i++ {
		s += ",?"
	}
	return s
}

func jsonToStringList(raw interface{}) []string {
	if raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case string:
		var out []string
		if json.Unmarshal([]byte(v), &out) == nil {
			return out
		}
		return nil
	case []byte:
		var out []string
		if json.Unmarshal(v, &out) == nil {
			return out
		}
		return nil
	}
	return nil
}

func stringListToJSON(list []string) string {
	if list == nil {
		return "[]"
	}
	b, _ := json.Marshal(list)
	return string(b)
}

func trimAnalyze(s string, maxLen int) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func (r *QuestionRepository) FindQuestions(userID int64, chapterID int64, favoriteOnly, difficultOnly, wrongOnly bool) ([]dto.QuestionListItem, error) {
	return r.FindQuestionsByChapterIDs(userID, []int64{chapterID}, favoriteOnly, difficultOnly, wrongOnly)
}

func (r *QuestionRepository) HasQuestionsInChapter(chapterID int64) (bool, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM ag_chapter_question WHERE chapter_id = ?", chapterID).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// FindQuestionsByChapterAndIndexRange 按父章节 + question_index 范围查询（用于子章节题目在父章节中的情况）
func (r *QuestionRepository) FindQuestionsByChapterAndIndexRange(userID int64, parentChapterID int64, startIdx, endIdx int, favoriteOnly, difficultOnly, wrongOnly bool) ([]dto.QuestionListItem, error) {
	query := `
		SELECT
			cq.question_id,
			cq.question_index,
			q.title_html,
			q.show_type_name,
			q.knowledge,
			COALESCE(uqm.favorite, 0) AS favorite,
			COALESCE(uqm.difficult, 0) AS difficult,
			COALESCE(stats.wrong_count, 0) AS wrong_count,
			stats.last_wrong_at,
			q.analyze_text
		FROM ag_chapter_question cq
		JOIN ag_question q ON q.question_id = cq.question_id
		LEFT JOIN user_question_mark uqm ON uqm.question_id = cq.question_id AND uqm.user_id = ?
		LEFT JOIN (
			SELECT question_id,
				SUM(CASE WHEN is_correct = 0 THEN 1 ELSE 0 END) AS wrong_count,
				MAX(CASE WHEN is_correct = 0 THEN answered_at END) AS last_wrong_at
			FROM user_answer_record
			WHERE user_id = ?
			GROUP BY question_id
		) stats ON stats.question_id = cq.question_id
		WHERE cq.chapter_id = ? AND cq.question_index >= ? AND cq.question_index <= ?
	`
	args := []interface{}{userID, userID, parentChapterID, startIdx, endIdx}

	if favoriteOnly {
		query += " AND COALESCE(uqm.favorite, 0) = 1"
	}
	if difficultOnly {
		query += " AND COALESCE(uqm.difficult, 0) = 1"
	}
	if wrongOnly {
		query += " AND COALESCE(stats.wrong_count, 0) > 0"
	}
	query += " ORDER BY cq.question_index, cq.question_id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.QuestionListItem
	for rows.Next() {
		var item dto.QuestionListItem
		var favorite, difficult, wrongCount int
		var lastWrongAt sql.NullString
		var analyzeText sql.NullString
		if err := rows.Scan(&item.QuestionID, &item.QuestionIndex, &item.TitleHtml, &item.ShowTypeName, &item.Knowledge,
			&favorite, &difficult, &wrongCount, &lastWrongAt, &analyzeText); err != nil {
			return nil, err
		}
		item.Favorite = favorite == 1
		item.Difficult = difficult == 1
		item.WrongCount = wrongCount
		if lastWrongAt.Valid {
			item.LastWrongAt = &lastWrongAt.String
		}
		if analyzeText.Valid {
			item.AnalyzePreview = trimAnalyze(analyzeText.String, 80)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

// FindQuestionsByParentAndSection 按父章 + section_chapter_id（小节）查题：爬虫有子节时题目存为 chapter_id=父章、section_chapter_id=小节ID
func (r *QuestionRepository) FindQuestionsByParentAndSection(userID int64, parentChapterID, sectionChapterID int64, favoriteOnly, difficultOnly, wrongOnly bool) ([]dto.QuestionListItem, error) {
	query := `
		SELECT
			cq.question_id,
			cq.question_index,
			q.title_html,
			q.show_type_name,
			q.knowledge,
			COALESCE(uqm.favorite, 0) AS favorite,
			COALESCE(uqm.difficult, 0) AS difficult,
			COALESCE(stats.wrong_count, 0) AS wrong_count,
			stats.last_wrong_at,
			q.analyze_text
		FROM ag_chapter_question cq
		JOIN ag_question q ON q.question_id = cq.question_id
		LEFT JOIN user_question_mark uqm ON uqm.question_id = cq.question_id AND uqm.user_id = ?
		LEFT JOIN (
			SELECT question_id,
				SUM(CASE WHEN is_correct = 0 THEN 1 ELSE 0 END) AS wrong_count,
				MAX(CASE WHEN is_correct = 0 THEN answered_at END) AS last_wrong_at
			FROM user_answer_record
			WHERE user_id = ?
			GROUP BY question_id
		) stats ON stats.question_id = cq.question_id
		WHERE cq.chapter_id = ? AND cq.section_chapter_id = ?
	`
	args := []interface{}{userID, userID, parentChapterID, sectionChapterID}
	if favoriteOnly {
		query += " AND COALESCE(uqm.favorite, 0) = 1"
	}
	if difficultOnly {
		query += " AND COALESCE(uqm.difficult, 0) = 1"
	}
	if wrongOnly {
		query += " AND COALESCE(stats.wrong_count, 0) > 0"
	}
	query += " ORDER BY cq.question_index, cq.question_id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []dto.QuestionListItem
	for rows.Next() {
		var item dto.QuestionListItem
		var favorite, difficult, wrongCount int
		var lastWrongAt sql.NullString
		var analyzeText sql.NullString
		if err := rows.Scan(&item.QuestionID, &item.QuestionIndex, &item.TitleHtml, &item.ShowTypeName, &item.Knowledge,
			&favorite, &difficult, &wrongCount, &lastWrongAt, &analyzeText); err != nil {
			return nil, err
		}
		item.Favorite = favorite == 1
		item.Difficult = difficult == 1
		item.WrongCount = wrongCount
		if lastWrongAt.Valid {
			item.LastWrongAt = &lastWrongAt.String
		}
		if analyzeText.Valid {
			item.AnalyzePreview = trimAnalyze(analyzeText.String, 80)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *QuestionRepository) FindQuestionsByChapterIDs(userID int64, chapterIDs []int64, favoriteOnly, difficultOnly, wrongOnly bool) ([]dto.QuestionListItem, error) {
	if len(chapterIDs) == 0 {
		return []dto.QuestionListItem{}, nil
	}
	query := `
		SELECT
			cq.question_id,
			cq.question_index,
			q.title_html,
			q.show_type_name,
			q.knowledge,
			COALESCE(uqm.favorite, 0) AS favorite,
			COALESCE(uqm.difficult, 0) AS difficult,
			COALESCE(stats.wrong_count, 0) AS wrong_count,
			stats.last_wrong_at,
			q.analyze_text
		FROM ag_chapter_question cq
		JOIN ag_question q ON q.question_id = cq.question_id
		LEFT JOIN user_question_mark uqm ON uqm.question_id = cq.question_id AND uqm.user_id = ?
		LEFT JOIN (
			SELECT question_id,
				SUM(CASE WHEN is_correct = 0 THEN 1 ELSE 0 END) AS wrong_count,
				MAX(CASE WHEN is_correct = 0 THEN answered_at END) AS last_wrong_at
			FROM user_answer_record
			WHERE user_id = ?
			GROUP BY question_id
		) stats ON stats.question_id = cq.question_id
		WHERE cq.chapter_id IN (` + placeholders(len(chapterIDs)) + `)
	`
	args := []interface{}{userID, userID}
	for _, id := range chapterIDs {
		args = append(args, id)
	}

	if favoriteOnly {
		query += " AND COALESCE(uqm.favorite, 0) = 1"
	}
	if difficultOnly {
		query += " AND COALESCE(uqm.difficult, 0) = 1"
	}
	if wrongOnly {
		query += " AND COALESCE(stats.wrong_count, 0) > 0"
	}
	query += " ORDER BY cq.chapter_id, cq.question_index, cq.question_id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.QuestionListItem
	for rows.Next() {
		var item dto.QuestionListItem
		var favorite, difficult, wrongCount int
		var lastWrongAt sql.NullString
		var analyzeText sql.NullString
		if err := rows.Scan(&item.QuestionID, &item.QuestionIndex, &item.TitleHtml, &item.ShowTypeName, &item.Knowledge,
			&favorite, &difficult, &wrongCount, &lastWrongAt, &analyzeText); err != nil {
			return nil, err
		}
		item.Favorite = favorite == 1
		item.Difficult = difficult == 1
		item.WrongCount = wrongCount
		if lastWrongAt.Valid {
			item.LastWrongAt = &lastWrongAt.String
		}
		if analyzeText.Valid {
			item.AnalyzePreview = trimAnalyze(analyzeText.String, 80)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *QuestionRepository) FindQuestionDetail(userID int64, questionID int64) (*dto.QuestionDetailDto, error) {
	var d dto.QuestionDetailDto
	var favorite, difficult, attemptCount, wrongCount int
	var note, analyzeText, materialText, answerJSON sql.NullString
	var lastWrongAt sql.NullString

	err := r.db.QueryRow(`
		SELECT
			q.question_id, q.title_html, q.show_type_name, q.knowledge, q.analyze_text, q.material_text, q.answer_json,
			COALESCE(uqm.favorite, 0) AS favorite,
			COALESCE(uqm.difficult, 0) AS difficult,
			COALESCE(uqm.note, '') AS note,
			COALESCE(stats.attempt_count, 0) AS attempt_count,
			COALESCE(stats.wrong_count, 0) AS wrong_count,
			stats.last_wrong_at
		FROM ag_question q
		LEFT JOIN user_question_mark uqm ON uqm.question_id = q.question_id AND uqm.user_id = ?
		LEFT JOIN (
			SELECT question_id,
				COUNT(*) AS attempt_count,
				SUM(CASE WHEN is_correct = 0 THEN 1 ELSE 0 END) AS wrong_count,
				MAX(CASE WHEN is_correct = 0 THEN answered_at END) AS last_wrong_at
			FROM user_answer_record
			WHERE user_id = ?
			GROUP BY question_id
		) stats ON stats.question_id = q.question_id
		WHERE q.question_id = ?
	`, userID, userID, questionID).Scan(
		&d.QuestionID, &d.TitleHtml, &d.ShowTypeName, &d.Knowledge, &analyzeText, &materialText, &answerJSON,
		&favorite, &difficult, &note, &attemptCount, &wrongCount, &lastWrongAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	d.Favorite = favorite == 1
	d.Difficult = difficult == 1
	d.Note = note.String
	d.AttemptCount = attemptCount
	d.WrongCount = wrongCount
	if analyzeText.Valid {
		d.AnalyzeText = analyzeText.String
	}
	if materialText.Valid {
		d.MaterialText = materialText.String
	}
	if answerJSON.Valid && answerJSON.String != "" {
		d.Answers = jsonToStringList(answerJSON.String)
	}
	if d.Answers == nil {
		d.Answers = []string{}
	}
	if lastWrongAt.Valid {
		d.LastWrongAt = &lastWrongAt.String
	}

	// options
	optRows, err := r.db.Query(`SELECT option_no, option_label, option_html FROM ag_question_option WHERE question_id = ? ORDER BY option_no`, questionID)
	if err != nil {
		return nil, err
	}
	defer optRows.Close()
	for optRows.Next() {
		var opt dto.QuestionOptionDto
		if err := optRows.Scan(&opt.OptionNo, &opt.OptionLabel, &opt.OptionHtml); err != nil {
			return nil, err
		}
		d.Options = append(d.Options, opt)
	}

	// recent records
	recRows, err := r.db.Query(`
		SELECT id, chapter_id, selected_answer_json, is_correct, answered_at
		FROM user_answer_record
		WHERE user_id = ? AND question_id = ?
		ORDER BY answered_at DESC, id DESC
		LIMIT 10
	`, userID, questionID)
	if err != nil {
		return nil, err
	}
	defer recRows.Close()
	for recRows.Next() {
		var rec dto.AnswerRecordDto
		var chapterID sql.NullInt64
		var selectedJSON sql.NullString
		var isCorrect int
		if err := recRows.Scan(&rec.ID, &chapterID, &selectedJSON, &isCorrect, &rec.AnsweredAt); err != nil {
			return nil, err
		}
		if chapterID.Valid {
			rec.ChapterID = &chapterID.Int64
		}
		rec.Correct = isCorrect == 1
		if selectedJSON.Valid {
			rec.SelectedAnswers = jsonToStringList(selectedJSON.String)
		}
		d.RecentRecords = append(d.RecentRecords, rec)
	}
	if d.RecentRecords == nil {
		d.RecentRecords = []dto.AnswerRecordDto{}
	}

	return &d, nil
}

func (r *QuestionRepository) FindCorrectAnswers(questionID int64) ([]string, error) {
	var raw sql.NullString
	if err := r.db.QueryRow("SELECT answer_json FROM ag_question WHERE question_id = ?", questionID).Scan(&raw); err != nil {
		return nil, err
	}
	if !raw.Valid || raw.String == "" {
		return []string{}, nil
	}
	return jsonToStringList(raw.String), nil
}

func (r *QuestionRepository) UpsertQuestionStatus(userID, questionID int64, favorite, difficult *bool, note string) error {
	var fav, diff int
	if favorite != nil && *favorite {
		fav = 1
	}
	if difficult != nil && *difficult {
		diff = 1
	}

	var query string
	switch r.driver {
	case db.DriverMySQL:
		query = `
			INSERT INTO user_question_mark (user_id, question_id, favorite, difficult, note)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				favorite = VALUES(favorite),
				difficult = VALUES(difficult),
				note = VALUES(note),
				updated_at = NOW()`
	default:
		query = `
			INSERT INTO user_question_mark (user_id, question_id, favorite, difficult, note)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(user_id, question_id) DO UPDATE SET
				favorite = excluded.favorite,
				difficult = excluded.difficult,
				note = excluded.note,
				updated_at = datetime('now')`
	}
	_, err := r.db.Exec(query, userID, questionID, fav, diff, note)
	return err
}

func (r *QuestionRepository) InsertAnswerRecord(userID, chapterID, questionID int64, selectedAnswers, correctAnswers []string, correct bool, durationSeconds *int) error {
	var dur interface{}
	if durationSeconds != nil {
		dur = *durationSeconds
	} else {
		dur = nil
	}
	_, err := r.db.Exec(`
		INSERT INTO user_answer_record (user_id, chapter_id, question_id, selected_answer_json, correct_answer_json, is_correct, duration_seconds)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, chapterID, questionID, stringListToJSON(selectedAnswers), stringListToJSON(correctAnswers), boolToInt(correct), dur)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *QuestionRepository) CountAttempts(userID, questionID int64) (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM user_answer_record WHERE user_id = ? AND question_id = ?", userID, questionID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) CountWrongs(userID, questionID int64) (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM user_answer_record WHERE user_id = ? AND question_id = ? AND is_correct = 0", userID, questionID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) TotalQuestions() (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM ag_question").Scan(&n)
	return n, err
}

func (r *QuestionRepository) FavoriteQuestions(userID int64) (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM user_question_mark WHERE user_id = ? AND favorite = 1", userID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) DifficultQuestions(userID int64) (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(*) FROM user_question_mark WHERE user_id = ? AND difficult = 1", userID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) WrongQuestions(userID int64) (int, error) {
	var n int
	err := r.db.QueryRow("SELECT COUNT(DISTINCT question_id) FROM user_answer_record WHERE user_id = ? AND is_correct = 0", userID).Scan(&n)
	return n, err
}

func (r *QuestionRepository) IsCorrect(selected, correct []string) bool {
	if len(selected) != len(correct) {
		return false
	}
	m := make(map[string]int)
	for _, s := range selected {
		m[s]++
	}
	for _, c := range correct {
		m[c]--
		if m[c] < 0 {
			return false
		}
	}
	return true
}
