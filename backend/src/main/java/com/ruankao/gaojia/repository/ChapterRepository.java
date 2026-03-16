package com.ruankao.gaojia.repository;

import com.ruankao.gaojia.dto.ChapterTreeNode;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Repository
public class ChapterRepository {

    private final JdbcTemplate jdbcTemplate;

    public ChapterRepository(JdbcTemplate jdbcTemplate) {
        this.jdbcTemplate = jdbcTemplate;
    }

    public List<ChapterTreeNode> findTree() {
        String sql = """
                SELECT chapter_id, chapter_name, chapter_level, sort_no, all_question_num, parent_chapter_id
                FROM ag_chapter
                ORDER BY chapter_level, sort_no, chapter_id
                """;
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(sql);
        Map<Long, ChapterTreeNode> allNodes = new HashMap<>();
        Map<Long, Long> parents = new HashMap<>();

        for (Map<String, Object> row : rows) {
            Long chapterId = ((Number) row.get("chapter_id")).longValue();
            ChapterTreeNode node = new ChapterTreeNode(
                    chapterId,
                    String.valueOf(row.get("chapter_name")),
                    ((Number) row.get("chapter_level")).intValue(),
                    ((Number) row.get("sort_no")).intValue(),
                    ((Number) row.get("all_question_num")).intValue()
            );
            allNodes.put(chapterId, node);
            parents.put(chapterId, ((Number) row.get("parent_chapter_id")).longValue());
        }

        List<ChapterTreeNode> roots = allNodes.values().stream()
                .filter(node -> parents.get(node.getChapterId()) == 0L)
                .sorted(Comparator.comparing(ChapterTreeNode::getSortNo))
                .toList();

        allNodes.values().stream()
                .filter(node -> parents.get(node.getChapterId()) != 0L)
                .forEach(node -> {
                    ChapterTreeNode parent = allNodes.get(parents.get(node.getChapterId()));
                    if (parent != null) {
                        parent.getChildren().add(node);
                        parent.getChildren().sort(Comparator.comparing(ChapterTreeNode::getSortNo));
                    }
                });

        return roots;
    }
}
