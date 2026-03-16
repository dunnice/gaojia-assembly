package service

import (
	"github.com/ruankao/gaojia-backend-go/internal/dto"
	"github.com/ruankao/gaojia-backend-go/internal/repository"
)

type ChapterService struct {
	chapterRepo *repository.ChapterRepository
}

func NewChapterService(chapterRepo *repository.ChapterRepository) *ChapterService {
	return &ChapterService{chapterRepo: chapterRepo}
}

func (s *ChapterService) Tree() ([]dto.ChapterTreeNode, error) {
	return s.chapterRepo.FindTree()
}
