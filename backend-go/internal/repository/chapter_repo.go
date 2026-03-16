package repository

import (
	"database/sql"
	"sort"

	"github.com/ruankao/gaojia-backend-go/internal/dto"
)

type ChapterRepository struct {
	db *sql.DB
}

func NewChapterRepository(db *sql.DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

func (r *ChapterRepository) FindTree() ([]dto.ChapterTreeNode, error) {
	rows, err := r.db.Query(`
		SELECT chapter_id, chapter_name, chapter_level, sort_no, all_question_num, parent_chapter_id
		FROM ag_chapter
		ORDER BY chapter_level, sort_no, chapter_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allNodes := make(map[int64]*dto.ChapterTreeNode)
	parents := make(map[int64]int64)

	for rows.Next() {
		var node dto.ChapterTreeNode
		var parentID int64
		if err := rows.Scan(&node.ChapterID, &node.ChapterName, &node.ChapterLevel, &node.SortNo, &node.AllQuestionNum, &parentID); err != nil {
			return nil, err
		}
		node.Children = []dto.ChapterTreeNode{}
		allNodes[node.ChapterID] = &node
		parents[node.ChapterID] = parentID
	}

	var roots []dto.ChapterTreeNode
	for _, node := range allNodes {
		if parents[node.ChapterID] == 0 {
			roots = append(roots, *node)
		} else if parent, ok := allNodes[parents[node.ChapterID]]; ok {
			parent.Children = append(parent.Children, *node)
		}
	}

	sort.Slice(roots, func(i, j int) bool { return roots[i].SortNo < roots[j].SortNo })
	for i := range roots {
		sort.Slice(roots[i].Children, func(a, b int) bool { return roots[i].Children[a].SortNo < roots[i].Children[b].SortNo })
	}

	return roots, nil
}

// GetDescendantIDs 返回某章节及其所有子章节的 ID 列表（含自身）
func (r *ChapterRepository) GetDescendantIDs(chapterID int64) ([]int64, error) {
	rows, err := r.db.Query(`
		WITH RECURSIVE tree AS (
			SELECT chapter_id, parent_chapter_id FROM ag_chapter WHERE chapter_id = ?
			UNION ALL
			SELECT c.chapter_id, c.parent_chapter_id FROM ag_chapter c
			INNER JOIN tree t ON c.parent_chapter_id = t.chapter_id
		)
		SELECT chapter_id FROM tree
	`, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetParentChapterID 返回父章节 ID，若为根章节则返回 0
func (r *ChapterRepository) GetParentChapterID(chapterID int64) (int64, error) {
	var parentID int64
	err := r.db.QueryRow("SELECT parent_chapter_id FROM ag_chapter WHERE chapter_id = ?", chapterID).Scan(&parentID)
	if err != nil {
		return 0, err
	}
	return parentID, nil
}

// GetChildQuestionIndexRange 子章节题目在父章节中的 question_index 范围 [start, end]（1-based）
// 根据同层兄弟的 all_question_num 累加得出
func (r *ChapterRepository) GetChildQuestionIndexRange(parentID, childID int64) (start, end int, err error) {
	rows, err := r.db.Query(`
		SELECT chapter_id, all_question_num FROM ag_chapter
		WHERE parent_chapter_id = ?
		ORDER BY sort_no, chapter_id
	`, parentID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	start = 1
	for rows.Next() {
		var cid int64
		var num int
		if err := rows.Scan(&cid, &num); err != nil {
			return 0, 0, err
		}
		if cid == childID {
			return start, start + num - 1, nil
		}
		start += num
	}
	return 0, 0, sql.ErrNoRows
}
