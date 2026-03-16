package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ruankao/gaojia-backend-go/internal/dto"
	"github.com/ruankao/gaojia-backend-go/internal/service"
)

type Handler struct {
	chapterService *service.ChapterService
	questionService *service.QuestionService
}

func NewHandler(chapterService *service.ChapterService, questionService *service.QuestionService) *Handler {
	return &Handler{
		chapterService:  chapterService,
		questionService: questionService,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/chapters/tree", h.ChapterTree)
		api.GET("/chapters/:chapterId/questions", h.ChapterQuestions)
		api.GET("/chapters/:chapterId/favorites", h.ChapterFavorites)
		api.GET("/questions/:questionId", h.QuestionDetail)
		api.POST("/questions/:questionId/status", h.UpdateQuestionStatus)
		api.POST("/questions/:questionId/answer", h.SubmitAnswer)
		api.GET("/review/wrongs", h.ReviewWrongs)
		api.GET("/review/summary", h.ReviewSummary)
	}
}

func (h *Handler) ChapterTree(c *gin.Context) {
	tree, err := h.chapterService.Tree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}

func (h *Handler) ChapterQuestions(c *gin.Context) {
	var path struct {
		ChapterID int64 `uri:"chapterId" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	favoriteOnly := c.Query("favoriteOnly") == "true"
	difficultOnly := c.Query("difficultOnly") == "true"
	wrongOnly := c.Query("wrongOnly") == "true"
	includeChildren := c.Query("includeChildren") == "true"

	list, err := h.questionService.ChapterQuestions(path.ChapterID, includeChildren, favoriteOnly, difficultOnly, wrongOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) ChapterFavorites(c *gin.Context) {
	var path struct {
		ChapterID int64 `uri:"chapterId" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	includeChildren := c.Query("includeChildren") == "true"
	list, err := h.questionService.ChapterQuestions(path.ChapterID, includeChildren, true, false, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) QuestionDetail(c *gin.Context) {
	var path struct {
		QuestionID int64 `uri:"questionId" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	detail, err := h.questionService.QuestionDetail(path.QuestionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *Handler) UpdateQuestionStatus(c *gin.Context) {
	var path struct {
		QuestionID int64 `uri:"questionId" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var req dto.QuestionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.questionService.UpdateStatus(path.QuestionID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) SubmitAnswer(c *gin.Context) {
	var path struct {
		QuestionID int64 `uri:"questionId" binding:"required"`
	}
	if err := c.ShouldBindUri(&path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var req dto.AnswerSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.questionService.SubmitAnswer(path.QuestionID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ReviewWrongs(c *gin.Context) {
	chapterID := c.Query("chapterId")
	if chapterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chapterId required"})
		return
	}
	cid, err := strconv.ParseInt(chapterID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapterId"})
		return
	}
	includeChildren := c.Query("includeChildren") == "true"
	list, err := h.questionService.ChapterQuestions(cid, includeChildren, false, false, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) ReviewSummary(c *gin.Context) {
	summary, err := h.questionService.Summary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}
