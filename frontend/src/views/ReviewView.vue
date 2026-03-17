<template>
  <main class="workspace review-layout">
    <section class="summary-strip">
      <div class="summary-card">
        <span>题库总量</span>
        <strong>{{ summary?.totalQuestions ?? 0 }}</strong>
      </div>
      <div class="summary-card">
        <span>收藏题</span>
        <strong>{{ summary?.favoriteQuestions ?? 0 }}</strong>
      </div>
      <div class="summary-card">
        <span>难题</span>
        <strong>{{ summary?.difficultQuestions ?? 0 }}</strong>
      </div>
      <div class="summary-card accent">
        <span>错题</span>
        <strong>{{ summary?.wrongQuestions ?? 0 }}</strong>
      </div>
    </section>

    <section class="workspace">
      <ChapterSidebar
        title="错题复习"
        description="按章节查看错题，追踪做错次数和最近错误时间。"
        :chapters="chapters"
        :selected-chapter-id="selectedChapterId"
        :include-children="includeChildren"
        @select="selectChapter"
      />

      <QuestionListPanel
        title="Review Queue"
        :subtitle="currentChapterName"
        :questions="questions"
        :selected-question-id="selectedQuestionId"
        @select="selectQuestion"
      >
        <template #actions>
          <span class="panel-tag panel-tag-danger">仅展示错题</span>
        </template>
      </QuestionListPanel>

      <QuestionDetailPanel
        :detail="detail"
        :selected-answers="selectedAnswers"
        :submit-result="submitResult"
        @update:selected-answers="selectedAnswers = $event"
        @submit-answer="submitAnswer"
        @toggle-favorite="toggleFavorite"
        @toggle-difficult="toggleDifficult"
        @save-note="saveNote"
      />
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '../api'
import { formatChapterTitle, stripEdition } from '../utils/chapter'
import ChapterSidebar from '../components/ChapterSidebar.vue'
import QuestionDetailPanel from '../components/QuestionDetailPanel.vue'
import QuestionListPanel from '../components/QuestionListPanel.vue'
import type { AnswerSubmitResponse, ChapterNode, QuestionDetail, QuestionListItem, ReviewSummary } from '../types'

const summary = ref<ReviewSummary | null>(null)
const chapters = ref<ChapterNode[]>([])
const questions = ref<QuestionListItem[]>([])
const detail = ref<QuestionDetail | null>(null)
const selectedChapterId = ref<number | null>(null)
const includeChildren = ref(false)
const selectedQuestionId = ref<number | null>(null)
const selectedAnswers = ref<string[]>([])
const submitResult = ref<AnswerSubmitResponse | null>(null)
let selectRequestId = 0

const currentChapterName = computed(() => {
  if (!selectedChapterId.value) return '请选择章节'
  for (let i = 0; i < chapters.value.length; i++) {
    const c = chapters.value[i]
    if (c.chapterId === selectedChapterId.value) {
      const formatted = formatChapterTitle(i + 1, c.chapterName)
      return includeChildren.value ? `${formatted}（整章）` : formatted
    }
    const childIdx = c.children?.findIndex((ch) => ch.chapterId === selectedChapterId.value)
    if (childIdx !== undefined && childIdx >= 0 && c.children) {
      return `第${childIdx + 1}节 ${stripEdition(c.children[childIdx].chapterName)}`
    }
  }
  return '请选择章节'
})

async function loadInitial() {
  ;[summary.value, chapters.value] = await Promise.all([api.fetchSummary(), api.fetchChapters()])
  if (chapters.value.length) {
    const first = chapters.value[0]
    await selectChapter(first.chapterId, first.children?.length ? true : false)
  }
}

async function selectChapter(chapterId: number, wholeChapter = false) {
  selectedChapterId.value = chapterId
  includeChildren.value = wholeChapter
  const list = await api.fetchWrongs(chapterId, wholeChapter)
  questions.value = list ?? []
  if (questions.value.length) {
    await selectQuestion(questions.value[0].questionId)
  } else {
    selectedQuestionId.value = null
    detail.value = null
  }
}

async function selectQuestion(questionId: number) {
  selectedQuestionId.value = questionId
  selectedAnswers.value = []
  submitResult.value = null
  const reqId = ++selectRequestId
  const fetched = await api.fetchQuestionDetail(questionId)
  if (reqId === selectRequestId) {
    detail.value = fetched
  }
}

async function submitAnswer() {
  if (!selectedQuestionId.value || !selectedChapterId.value || selectedAnswers.value.length === 0) {
    return
  }
  submitResult.value = await api.submitAnswer(selectedQuestionId.value, {
    chapterId: selectedChapterId.value,
    selectedAnswers: selectedAnswers.value
  })
  detail.value = await api.fetchQuestionDetail(selectedQuestionId.value)
  questions.value = await api.fetchWrongs(selectedChapterId.value, includeChildren.value)
  summary.value = await api.fetchSummary()
}

async function toggleFavorite(value: boolean) {
  if (!selectedQuestionId.value) return
  await api.updateQuestionStatus(selectedQuestionId.value, { favorite: value })
  detail.value = await api.fetchQuestionDetail(selectedQuestionId.value)
  summary.value = await api.fetchSummary()
}

async function toggleDifficult(value: boolean) {
  if (!selectedQuestionId.value) return
  await api.updateQuestionStatus(selectedQuestionId.value, { difficult: value })
  detail.value = await api.fetchQuestionDetail(selectedQuestionId.value)
  summary.value = await api.fetchSummary()
}

async function saveNote(note: string) {
  if (!selectedQuestionId.value) return
  await api.updateQuestionStatus(selectedQuestionId.value, { note })
  detail.value = await api.fetchQuestionDetail(selectedQuestionId.value)
}

onMounted(loadInitial)
</script>
