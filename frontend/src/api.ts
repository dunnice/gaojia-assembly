import axios from 'axios'
import type { AnswerSubmitResponse, ChapterNode, QuestionDetail, QuestionListItem, ReviewSummary } from './types'

const baseURL = import.meta.env.VITE_API_BASE ?? 'http://localhost:5170/api'

const client = axios.create({
  baseURL,
  timeout: 10000
})

export const api = {
  async fetchChapters() {
    const { data } = await client.get<ChapterNode[]>('/chapters/tree')
    return data
  },
  async fetchQuestions(chapterId: number, params?: Record<string, boolean>) {
    const { data } = await client.get<QuestionListItem[]>(`/chapters/${chapterId}/questions`, { params })
    return data
  },
  async fetchFavorites(chapterId: number, includeChildren?: boolean) {
    const { data } = await client.get<QuestionListItem[]>(`/chapters/${chapterId}/favorites`, {
      params: includeChildren ? { includeChildren: true } : undefined
    })
    return data
  },
  async fetchWrongs(chapterId: number, includeChildren?: boolean) {
    const { data } = await client.get<QuestionListItem[]>('/review/wrongs', {
      params: { chapterId, ...(includeChildren ? { includeChildren: true } : {}) }
    })
    return data
  },
  async fetchQuestionDetail(questionId: number) {
    const { data } = await client.get<QuestionDetail>(`/questions/${questionId}`)
    return data
  },
  async updateQuestionStatus(questionId: number, payload: { favorite?: boolean; difficult?: boolean; note?: string }) {
    await client.post(`/questions/${questionId}/status`, payload)
  },
  async submitAnswer(questionId: number, payload: { chapterId: number; selectedAnswers: string[]; durationSeconds?: number }) {
    const { data } = await client.post<AnswerSubmitResponse>(`/questions/${questionId}/answer`, payload)
    return data
  },
  async fetchSummary() {
    const { data } = await client.get<ReviewSummary>('/review/summary')
    return data
  }
}
