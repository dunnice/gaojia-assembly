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

	// 用 ag_chapter_question 实际题数覆盖 all_question_num（目录数可能与同步结果不一致）
	sectionCounts, _ := r.sectionQuestionCounts()
	chapterTotals, _ := r.chapterQuestionTotals()
	for id, node := range allNodes {
		if parents[id] != 0 {
			if cnt, ok := sectionCounts[id]; ok {
				node.AllQuestionNum = cnt
			}
		} else if cnt, ok := chapterTotals[id]; ok {
			node.AllQuestionNum = cnt
		}
	}

	// 先只挂父子关系：子节点 append 到 allNodes 里的父节点（不在此处收集 roots，否则 map 迭代顺序随机，根节点可能先于其子被复制，导致 roots 里副本的 Children 一直为空）
	for _, node := range allNodes {
		if parents[node.ChapterID] == 0 {
			continue
		}
		if parent, ok := allNodes[parents[node.ChapterID]]; ok {
			parent.Children = append(parent.Children, *node)
		}
	}

	// 树已构建完成，再按 parent==0 收集根节点（此时每个节点的 Children 已完整）
	var roots []dto.ChapterTreeNode
	for id, node := range allNodes {
		if parents[id] == 0 {
			roots = append(roots, *node)
		}
	}

	// 稳定排序：SortNo 相同时按 ChapterID 排
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].SortNo != roots[j].SortNo {
			return roots[i].SortNo < roots[j].SortNo
		}
		return roots[i].ChapterID < roots[j].ChapterID
	})
	for i := range roots {
		sort.Slice(roots[i].Children, func(a, b int) bool {
			if roots[i].Children[a].SortNo != roots[i].Children[b].SortNo {
				return roots[i].Children[a].SortNo < roots[i].Children[b].SortNo
			}
			return roots[i].Children[a].ChapterID < roots[i].ChapterID
		})
	}

	return roots, nil
}

// sectionQuestionCounts 返回各小节（section_chapter_id）在 ag_chapter_question 中的实际题数，用于覆盖 all_question_num
func (r *ChapterRepository) sectionQuestionCounts() (map[int64]int, error) {
	rows, err := r.db.Query(`
		SELECT section_chapter_id, COUNT(*) AS cnt
		FROM ag_chapter_question
		WHERE section_chapter_id > 0
		GROUP BY section_chapter_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64]int)
	for rows.Next() {
		var sectionID int64
		var cnt int
		if err := rows.Scan(&sectionID, &cnt); err != nil {
			return nil, err
		}
		m[sectionID] = cnt
	}
	return m, rows.Err()
}

// chapterQuestionTotals 返回各章（chapter_id）在 ag_chapter_question 中的总题数，用于「整章」展示
func (r *ChapterRepository) chapterQuestionTotals() (map[int64]int, error) {
	rows, err := r.db.Query(`
		SELECT chapter_id, COUNT(*) AS cnt
		FROM ag_chapter_question
		GROUP BY chapter_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64]int)
	for rows.Next() {
		var chapterID int64
		var cnt int
		if err := rows.Scan(&chapterID, &cnt); err != nil {
			return nil, err
		}
		m[chapterID] = cnt
	}
	return m, rows.Err()
}

// GetDescendantIDs 返回某章节及其所有子章节的 ID 列表（含自身），按 sort_no, chapter_id 有序
func (r *ChapterRepository) GetDescendantIDs(chapterID int64) ([]int64, error) {
	rows, err := r.db.Query(`
		WITH RECURSIVE tree AS (
			SELECT chapter_id, parent_chapter_id, sort_no FROM ag_chapter WHERE chapter_id = ?
			UNION ALL
			SELECT c.chapter_id, c.parent_chapter_id, c.sort_no FROM ag_chapter c
			INNER JOIN tree t ON c.parent_chapter_id = t.chapter_id
		)
		SELECT t.chapter_id FROM tree t
		JOIN ag_chapter ac ON ac.chapter_id = t.chapter_id
		ORDER BY ac.chapter_level, ac.sort_no, t.chapter_id
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
