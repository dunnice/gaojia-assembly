package service

import (
	"time"

	"github.com/ruankao/gaojia-backend-go/internal/dto"
	"github.com/ruankao/gaojia-backend-go/internal/repository"
)

type QuestionService struct {
	questionRepo  *repository.QuestionRepository
	chapterRepo   *repository.ChapterRepository
	defaultUserID int64
}

func NewQuestionService(questionRepo *repository.QuestionRepository, chapterRepo *repository.ChapterRepository, defaultUserID int64) *QuestionService {
	return &QuestionService{
		questionRepo:  questionRepo,
		chapterRepo:   chapterRepo,
		defaultUserID: defaultUserID,
	}
}

func (s *QuestionService) ChapterQuestions(chapterID int64, includeChildren, favoriteOnly, difficultOnly, wrongOnly bool) ([]dto.QuestionListItem, error) {
	if includeChildren {
		ids, err := s.chapterRepo.GetDescendantIDs(chapterID)
		if err == nil && len(ids) > 0 {
			return s.questionRepo.FindQuestionsByChapterIDs(s.defaultUserID, ids, favoriteOnly, difficultOnly, wrongOnly)
		}
	}
	// 单章节：先按 chapter_id 直接查（无子节时题目可能直接挂在该 chapter_id 下）
	list, err := s.questionRepo.FindQuestionsByChapterIDs(s.defaultUserID, []int64{chapterID}, favoriteOnly, difficultOnly, wrongOnly)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return list, nil
	}
	// 无结果时，可能是子章节（小节）：爬虫有子节时题目存为 chapter_id=父章、section_chapter_id=小节ID
	parentID, err := s.chapterRepo.GetParentChapterID(chapterID)
	if err != nil || parentID == 0 {
		return list, nil
	}
	list, err = s.questionRepo.FindQuestionsByParentAndSection(s.defaultUserID, parentID, chapterID, favoriteOnly, difficultOnly, wrongOnly)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		return list, nil
	}
	// 再 fallback：按父章下 question_index 范围查（兼容旧数据或按 all_question_num 划分的存储）
	start, end, err := s.chapterRepo.GetChildQuestionIndexRange(parentID, chapterID)
	if err != nil || start > end {
		return list, nil
	}
	return s.questionRepo.FindQuestionsByChapterAndIndexRange(s.defaultUserID, parentID, start, end, favoriteOnly, difficultOnly, wrongOnly)
}

func (s *QuestionService) QuestionDetail(questionID int64) (*dto.QuestionDetailDto, error) {
	return s.questionRepo.FindQuestionDetail(s.defaultUserID, questionID)
}

func (s *QuestionService) UpdateStatus(questionID int64, req *dto.QuestionStatusRequest) error {
	current, err := s.questionRepo.FindQuestionDetail(s.defaultUserID, questionID)
	if err != nil || current == nil {
		return err
	}
	fav := req.Favorite
	diff := req.Difficult
	note := ""
	if req.Note != nil {
		note = *req.Note
	}
	if fav == nil {
		f := current.Favorite
		fav = &f
	}
	if diff == nil {
		d := current.Difficult
		diff = &d
	}
	return s.questionRepo.UpsertQuestionStatus(s.defaultUserID, questionID, fav, diff, note)
}

func (s *QuestionService) SubmitAnswer(questionID int64, req *dto.AnswerSubmitRequest) (*dto.AnswerSubmitResponse, error) {
	correctAnswers, err := s.questionRepo.FindCorrectAnswers(questionID)
	if err != nil {
		return nil, err
	}
	correct := s.questionRepo.IsCorrect(req.SelectedAnswers, correctAnswers)

	if err := s.questionRepo.InsertAnswerRecord(
		s.defaultUserID,
		req.ChapterID,
		questionID,
		req.SelectedAnswers,
		correctAnswers,
		correct,
		req.DurationSeconds,
	); err != nil {
		return nil, err
	}

	attempts, _ := s.questionRepo.CountAttempts(s.defaultUserID, questionID)
	wrongs, _ := s.questionRepo.CountWrongs(s.defaultUserID, questionID)

	return &dto.AnswerSubmitResponse{
		Correct:        correct,
		CorrectAnswers: correctAnswers,
		TotalAttempts:  attempts,
		WrongCount:     wrongs,
		AnsweredAt:     time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *QuestionService) Summary() (*dto.ReviewSummaryDto, error) {
	total, _ := s.questionRepo.TotalQuestions()
	fav, _ := s.questionRepo.FavoriteQuestions(s.defaultUserID)
	diff, _ := s.questionRepo.DifficultQuestions(s.defaultUserID)
	wrong, _ := s.questionRepo.WrongQuestions(s.defaultUserID)
	return &dto.ReviewSummaryDto{
		TotalQuestions:    total,
		FavoriteQuestions: fav,
		DifficultQuestions: diff,
		WrongQuestions:    wrong,
	}, nil
}
