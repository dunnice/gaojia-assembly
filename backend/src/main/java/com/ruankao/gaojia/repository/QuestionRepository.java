package com.ruankao.gaojia.repository;

import com.ruankao.gaojia.dto.AnswerRecordDto;
import com.ruankao.gaojia.dto.QuestionDetailDto;
import com.ruankao.gaojia.dto.QuestionListItem;
import com.ruankao.gaojia.dto.QuestionOptionDto;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.namedparam.MapSqlParameterSource;
import org.springframework.jdbc.core.namedparam.NamedParameterJdbcTemplate;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

@Repository
public class QuestionRepository {

    private static final DateTimeFormatter DATE_TIME_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private final JdbcTemplate jdbcTemplate;
    private final NamedParameterJdbcTemplate namedParameterJdbcTemplate;

    public QuestionRepository(JdbcTemplate jdbcTemplate, NamedParameterJdbcTemplate namedParameterJdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
        this.namedParameterJdbcTemplate = namedParameterJdbcTemplate;
    }

    public List<QuestionListItem> findQuestions(Long userId, Long chapterId, boolean favoriteOnly, boolean difficultOnly, boolean wrongOnly) {
        StringBuilder sql = new StringBuilder("""
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
                LEFT JOIN user_question_mark uqm
                    ON uqm.question_id = cq.question_id AND uqm.user_id = :userId
                LEFT JOIN (
                    SELECT question_id,
                           SUM(CASE WHEN is_correct = 0 THEN 1 ELSE 0 END) AS wrong_count,
                           MAX(CASE WHEN is_correct = 0 THEN answered_at END) AS last_wrong_at
                    FROM user_answer_record
                    WHERE user_id = :userId
                    GROUP BY question_id
                ) stats ON stats.question_id = cq.question_id
                WHERE cq.chapter_id = :chapterId
                """);
        MapSqlParameterSource params = new MapSqlParameterSource()
                .addValue("userId", userId)
                .addValue("chapterId", chapterId);

        if (favoriteOnly) {
            sql.append(" AND COALESCE(uqm.favorite, 0) = 1");
        }
        if (difficultOnly) {
            sql.append(" AND COALESCE(uqm.difficult, 0) = 1");
        }
        if (wrongOnly) {
            sql.append(" AND COALESCE(stats.wrong_count, 0) > 0");
        }
        sql.append(" ORDER BY cq.question_index");

        return namedParameterJdbcTemplate.query(sql.toString(), params, (rs, rowNum) -> new QuestionListItem(
                rs.getLong("question_id"),
                rs.getInt("question_index"),
                rs.getString("title_html"),
                rs.getString("show_type_name"),
                rs.getString("knowledge"),
                rs.getInt("favorite") == 1,
                rs.getInt("difficult") == 1,
                rs.getInt("wrong_count"),
                formatTimestamp(rs.getTimestamp("last_wrong_at")),
                trimAnalyze(rs.getString("analyze_text"))
        ));
    }

    public QuestionDetailDto findQuestionDetail(Long userId, Long questionId) {
        String sql = """
                SELECT
                    q.question_id,
                    q.title_html,
                    q.show_type_name,
                    q.knowledge,
                    q.analyze_text,
                    q.material_text,
                    q.answer_json,
                    COALESCE(uqm.favorite, 0) AS favorite,
                    COALESCE(uqm.difficult, 0) AS difficult,
                    COALESCE(uqm.note, '') AS note,
                    COALESCE(stats.attempt_count, 0) AS attempt_count,
                    COALESCE(stats.wrong_count, 0) AS wrong_count,
                    stats.last_wrong_at
                FROM ag_question q
                LEFT JOIN user_question_mark uqm
                    ON uqm.question_id = q.question_id AND uqm.user_id = ?
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
                """;
        Map<String, Object> detail = jdbcTemplate.queryForMap(sql, userId, userId, questionId);

        List<QuestionOptionDto> options = jdbcTemplate.query("""
                        SELECT option_no, option_label, option_html
                        FROM ag_question_option
                        WHERE question_id = ?
                        ORDER BY option_no
                        """,
                (rs, rowNum) -> new QuestionOptionDto(
                        rs.getInt("option_no"),
                        rs.getString("option_label"),
                        rs.getString("option_html")
                ),
                questionId
        );

        List<AnswerRecordDto> recentRecords = jdbcTemplate.query("""
                        SELECT id, chapter_id, selected_answer_json, is_correct, answered_at
                        FROM user_answer_record
                        WHERE user_id = ? AND question_id = ?
                        ORDER BY answered_at DESC, id DESC
                        LIMIT 10
                        """,
                (rs, rowNum) -> new AnswerRecordDto(
                        rs.getLong("id"),
                        rs.getObject("chapter_id") == null ? null : rs.getLong("chapter_id"),
                        JsonHelper.toStringList(rs.getObject("selected_answer_json")),
                        rs.getInt("is_correct") == 1,
                        formatTimestamp(rs.getTimestamp("answered_at"))
                ),
                userId, questionId
        );

        return new QuestionDetailDto(
                ((Number) detail.get("question_id")).longValue(),
                String.valueOf(detail.get("title_html")),
                String.valueOf(detail.get("show_type_name")),
                String.valueOf(detail.get("knowledge")),
                String.valueOf(detail.get("analyze_text")),
                String.valueOf(detail.get("material_text")),
                JsonHelper.toStringList(detail.get("answer_json")),
                options,
                ((Number) detail.get("favorite")).intValue() == 1,
                ((Number) detail.get("difficult")).intValue() == 1,
                String.valueOf(detail.get("note")),
                ((Number) detail.get("attempt_count")).intValue(),
                ((Number) detail.get("wrong_count")).intValue(),
                formatTimestampObject(detail.get("last_wrong_at")),
                recentRecords
        );
    }

    public List<String> findCorrectAnswers(Long questionId) {
        String sql = "SELECT answer_json FROM ag_question WHERE question_id = ?";
        Object raw = jdbcTemplate.queryForObject(sql, Object.class, questionId);
        return JsonHelper.toStringList(raw);
    }

    public void upsertQuestionStatus(Long userId, Long questionId, Boolean favorite, Boolean difficult, String note) {
        jdbcTemplate.update("""
                        INSERT INTO user_question_mark (user_id, question_id, favorite, difficult, note)
                        VALUES (?, ?, ?, ?, ?)
                        ON DUPLICATE KEY UPDATE
                            favorite = VALUES(favorite),
                            difficult = VALUES(difficult),
                            note = VALUES(note),
                            updated_at = CURRENT_TIMESTAMP
                        """,
                userId,
                questionId,
                favorite != null && favorite ? 1 : 0,
                difficult != null && difficult ? 1 : 0,
                note == null ? "" : note
        );
    }

    public void insertAnswerRecord(Long userId, Long chapterId, Long questionId, List<String> selectedAnswers, List<String> correctAnswers, boolean correct, Integer durationSeconds) {
        jdbcTemplate.update("""
                        INSERT INTO user_answer_record (
                            user_id, chapter_id, question_id, selected_answer_json, correct_answer_json, is_correct, duration_seconds
                        ) VALUES (?, ?, ?, CAST(? AS JSON), CAST(? AS JSON), ?, ?)
                        """,
                userId,
                chapterId,
                questionId,
                JsonHelper.toJson(selectedAnswers),
                JsonHelper.toJson(correctAnswers),
                correct ? 1 : 0,
                durationSeconds
        );
    }

    public int countAttempts(Long userId, Long questionId) {
        Integer count = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM user_answer_record WHERE user_id = ? AND question_id = ?",
                Integer.class,
                userId,
                questionId
        );
        return count == null ? 0 : count;
    }

    public int countWrongs(Long userId, Long questionId) {
        Integer count = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM user_answer_record WHERE user_id = ? AND question_id = ? AND is_correct = 0",
                Integer.class,
                userId,
                questionId
        );
        return count == null ? 0 : count;
    }

    public int totalQuestions() {
        Integer count = jdbcTemplate.queryForObject("SELECT COUNT(*) FROM ag_question", Integer.class);
        return count == null ? 0 : count;
    }

    public int favoriteQuestions(Long userId) {
        Integer count = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM user_question_mark WHERE user_id = ? AND favorite = 1",
                Integer.class,
                userId
        );
        return count == null ? 0 : count;
    }

    public int difficultQuestions(Long userId) {
        Integer count = jdbcTemplate.queryForObject(
                "SELECT COUNT(*) FROM user_question_mark WHERE user_id = ? AND difficult = 1",
                Integer.class,
                userId
        );
        return count == null ? 0 : count;
    }

    public int wrongQuestions(Long userId) {
        Integer count = jdbcTemplate.queryForObject("""
                        SELECT COUNT(DISTINCT question_id)
                        FROM user_answer_record
                        WHERE user_id = ? AND is_correct = 0
                        """,
                Integer.class,
                userId
        );
        return count == null ? 0 : count;
    }

    public boolean isCorrect(List<String> selectedAnswers, List<String> correctAnswers) {
        return new HashSet<>(selectedAnswers).equals(new HashSet<>(correctAnswers))
                && selectedAnswers.size() == correctAnswers.size();
    }

    private String trimAnalyze(String analyze) {
        if (analyze == null || analyze.isBlank()) {
            return "";
        }
        return analyze.length() <= 80 ? analyze : analyze.substring(0, 80) + "...";
    }

    private String formatTimestamp(Timestamp timestamp) {
        if (timestamp == null) {
            return null;
        }
        return timestamp.toLocalDateTime().format(DATE_TIME_FORMATTER);
    }

    /** 兼容 JDBC 返回 Timestamp 或 LocalDateTime（如 MySQL Connector/J 8+） */
    private String formatTimestampObject(Object value) {
        if (value == null) {
            return null;
        }
        if (value instanceof Timestamp ts) {
            return ts.toLocalDateTime().format(DATE_TIME_FORMATTER);
        }
        if (value instanceof LocalDateTime ldt) {
            return ldt.format(DATE_TIME_FORMATTER);
        }
        if (value instanceof java.util.Date d) {
            return d.toInstant().atZone(java.time.ZoneId.systemDefault()).toLocalDateTime().format(DATE_TIME_FORMATTER);
        }
        return String.valueOf(value);
    }
}
