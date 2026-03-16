package dto

type ChapterTreeNode struct {
	ChapterID      int64             `json:"chapterId"`
	ChapterName    string            `json:"chapterName"`
	ChapterLevel   int               `json:"chapterLevel"`
	SortNo         int               `json:"sortNo"`
	AllQuestionNum int               `json:"allQuestionNum"`
	Children       []ChapterTreeNode `json:"children"`
}

type QuestionListItem struct {
	QuestionID     int64   `json:"questionId"`
	QuestionIndex  int     `json:"questionIndex"`
	TitleHtml      string  `json:"titleHtml"`
	ShowTypeName   string  `json:"showTypeName"`
	Knowledge      string  `json:"knowledge"`
	Favorite       bool    `json:"favorite"`
	Difficult      bool    `json:"difficult"`
	WrongCount     int     `json:"wrongCount"`
	LastWrongAt    *string `json:"lastWrongAt"`
	AnalyzePreview string  `json:"analyzePreview"`
}

type QuestionOptionDto struct {
	OptionNo    int    `json:"optionNo"`
	OptionLabel string `json:"optionLabel"`
	OptionHtml  string `json:"optionHtml"`
}

type AnswerRecordDto struct {
	ID             int64    `json:"id"`
	ChapterID      *int64   `json:"chapterId"`
	SelectedAnswers []string `json:"selectedAnswers"`
	Correct        bool     `json:"correct"`
	AnsweredAt     string   `json:"answeredAt"`
}

type QuestionDetailDto struct {
	QuestionID   int64              `json:"questionId"`
	TitleHtml    string             `json:"titleHtml"`
	ShowTypeName string             `json:"showTypeName"`
	Knowledge    string             `json:"knowledge"`
	AnalyzeText  string             `json:"analyzeText"`
	MaterialText string             `json:"materialText"`
	Answers      []string           `json:"answers"`
	Options      []QuestionOptionDto `json:"options"`
	Favorite     bool               `json:"favorite"`
	Difficult    bool               `json:"difficult"`
	Note         string             `json:"note"`
	AttemptCount int                `json:"attemptCount"`
	WrongCount   int                `json:"wrongCount"`
	LastWrongAt  *string            `json:"lastWrongAt"`
	RecentRecords []AnswerRecordDto  `json:"recentRecords"`
}

type AnswerSubmitRequest struct {
	ChapterID      int64    `json:"chapterId" binding:"required"`
	SelectedAnswers []string `json:"selectedAnswers" binding:"required"`
	DurationSeconds *int    `json:"durationSeconds"`
}

type AnswerSubmitResponse struct {
	Correct       bool     `json:"correct"`
	CorrectAnswers []string `json:"correctAnswers"`
	TotalAttempts int      `json:"totalAttempts"`
	WrongCount    int      `json:"wrongCount"`
	AnsweredAt   string   `json:"answeredAt"`
}

type QuestionStatusRequest struct {
	Favorite  *bool   `json:"favorite"`
	Difficult *bool   `json:"difficult"`
	Note      *string `json:"note"`
}

type ReviewSummaryDto struct {
	TotalQuestions    int `json:"totalQuestions"`
	FavoriteQuestions int `json:"favoriteQuestions"`
	DifficultQuestions int `json:"difficultQuestions"`
	WrongQuestions    int `json:"wrongQuestions"`
}
