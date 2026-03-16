export interface ChapterNode {
  chapterId: number
  chapterName: string
  chapterLevel: number
  sortNo: number
  allQuestionNum: number
  children: ChapterNode[]
}

export interface QuestionListItem {
  questionId: number
  questionIndex: number
  titleHtml: string
  showTypeName: string
  knowledge: string
  favorite: boolean
  difficult: boolean
  wrongCount: number
  lastWrongAt: string | null
  analyzePreview: string
}

export interface QuestionOption {
  optionNo: number
  optionLabel: string
  optionHtml: string
}

export interface AnswerRecord {
  id: number
  chapterId: number | null
  selectedAnswers: string[]
  correct: boolean
  answeredAt: string
}

export interface QuestionDetail {
  questionId: number
  titleHtml: string
  showTypeName: string
  knowledge: string
  analyzeText: string
  materialText: string
  answers: string[]
  options: QuestionOption[]
  favorite: boolean
  difficult: boolean
  note: string
  attemptCount: number
  wrongCount: number
  lastWrongAt: string | null
  recentRecords: AnswerRecord[]
}

export interface ReviewSummary {
  totalQuestions: number
  favoriteQuestions: number
  difficultQuestions: number
  wrongQuestions: number
}

export interface AnswerSubmitResponse {
  correct: boolean
  correctAnswers: string[]
  totalAttempts: number
  wrongCount: number
  answeredAt: string
}
