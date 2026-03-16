<template>
  <main class="workspace">
    <ChapterSidebar
      title="收藏难题"
      description="按章节查看你标注的收藏题和难题。"
      :chapters="chapters"
      :selected-chapter-id="selectedChapterId"
      :include-children="includeChildren"
      @select="selectChapter"
    />

    <QuestionListPanel
      title="Favorite Queue"
      :subtitle="currentChapterName"
      :questions="questions"
      :selected-question-id="selectedQuestionId"
      @select="selectQuestion"
    >
      <template #actions>
        <span class="panel-tag">仅展示收藏</span>
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
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '../api'
import ChapterSidebar from '../components/ChapterSidebar.vue'
import QuestionDetailPanel from '../components/QuestionDetailPanel.vue'
import QuestionListPanel from '../components/QuestionListPanel.vue'
import type { AnswerSubmitResponse, ChapterNode, QuestionDetail, QuestionListItem } from '../types'

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
  for (const c of chapters.value) {
    if (c.chapterId === selectedChapterId.value) {
      return includeChildren.value ? `${c.chapterName}（整章）` : c.chapterName
    }
    const child = c.children?.find((ch) => ch.chapterId === selectedChapterId.value)
    if (child) return child.chapterName
  }
  return '请选择章节'
})

async function loadChapters() {
  chapters.value = await api.fetchChapters()
  if (chapters.value.length) {
    const first = chapters.value[0]
    await selectChapter(first.chapterId, first.children?.length ? true : false)
  }
}

async function selectChapter(chapterId: number, wholeChapter = false) {
  selectedChapterId.value = chapterId
  includeChildren.value = wholeChapter
  const list = await api.fetchFavorites(chapterId, wholeChapter)
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
  questions.value = await api.fetchFavorites(selectedChapterId.value, includeChildren.value)
}

async function toggleFavorite(value: boolean) {
  if (!selectedQuestionId.value) return
  await api.updateQuestionStatus(selectedQuestionId.value, { favorite: value })
  if (selectedChapterId.value) {
    questions.value = await api.fetchFavorites(selectedChapterId.value, includeChildren.value)
  }
  detail.value = value ? await api.fetchQuestionDetail(selectedQuestionId.value) : null
}

async function toggleDifficult(value: boolean) {
  if (!selectedQuestionId.value) return
  await api.updateQuestionStatus(selectedQuestionId.value, { difficult: value })
  detail.value = await api.fetchQuestionDetail(selectedQuestionId.value)
}

async function saveNote(note: string) {
  if (!selectedQuestionId.value) return
  await api.updateQuestionStatus(selectedQuestionId.value, { note })
  detail.value = await api.fetchQuestionDetail(selectedQuestionId.value)
}

onMounted(loadChapters)
</script>
