<template>
  <section class="panel detail-panel">
    <div v-if="detail">
      <div class="panel-head">
        <div>
          <p class="eyebrow">题目详情</p>
          <h2>{{ detail.showTypeName }} · {{ detail.knowledge || '未标注知识点' }}</h2>
        </div>
        <div class="summary-mini">
          <span>答题 {{ detail.attemptCount }}</span>
          <span>错题 {{ detail.wrongCount }}</span>
        </div>
      </div>

      <article class="question-detail">
        <div v-if="detail.materialText" class="material-box" v-html="detail.materialText" @click="onContentClick"></div>
        <div class="question-title" v-html="detail.titleHtml" @click="onContentClick"></div>

        <div class="options">
          <label v-for="option in detail.options" :key="option.optionLabel" class="option-item">
            <input
              :type="isMultiple ? 'checkbox' : 'radio'"
              :value="option.optionLabel"
              :checked="selectedAnswers.includes(option.optionLabel)"
              @change="onOptionChange(option.optionLabel, $event)"
            />
            <span class="option-label">{{ option.optionLabel }}</span>
            <span class="option-content" v-html="option.optionHtml" @click="onContentClick"></span>
          </label>
        </div>

        <div class="toolbar">
          <button
            type="button"
            class="icon-btn favorite-btn"
            :class="{ active: detail.favorite }"
            :title="detail.favorite ? '取消收藏' : '加入收藏'"
            @click="$emit('toggle-favorite', !detail.favorite)"
          >
            <svg v-if="detail.favorite" class="heart-icon" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
            </svg>
            <svg v-else class="heart-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
            </svg>
          </button>
          <button class="ghost-btn" @click="$emit('toggle-difficult', !detail.difficult)">
            {{ detail.difficult ? '取消难题' : '标记难题' }}
          </button>
        </div>

        <div v-if="submitResult" class="result-box" :class="{ success: submitResult.correct, error: !submitResult.correct }">
          <strong>{{ submitResult.correct ? '回答正确' : '回答错误' }}</strong>
          <span>正确答案：{{ submitResult.correctAnswers.join(', ') }}</span>
        </div>

        <section class="analysis-box">
          <h3>解析</h3>
          <div v-if="detail.analyzeText" class="analysis-content" v-html="decodeHtml(detail.analyzeText)" @click="onContentClick"></div>
          <p v-else class="empty-inline">当前题目暂无解析。</p>
        </section>

        <section class="note-box">
          <div class="note-head">
            <h3>我的备注</h3>
            <button class="ghost-btn" @click="$emit('save-note', noteText)">保存备注</button>
          </div>
          <textarea v-model="noteText" placeholder="记录易错点、记忆口诀、解题思路"></textarea>
        </section>

        <section class="history-box">
          <h3>最近答题记录</h3>
          <div v-if="detail.recentRecords?.length" class="history-list">
            <div v-for="record in detail.recentRecords" :key="record.id" class="history-item">
              <strong :class="{ correct: record.correct, wrong: !record.correct }">
                {{ record.correct ? '正确' : '错误' }}
              </strong>
              <span>作答：{{ record.selectedAnswers?.join(', ') || '空' }}</span>
              <small>{{ record.answeredAt }}</small>
            </div>
          </div>
          <div v-else class="empty-inline">还没有答题记录。</div>
        </section>
      </article>
    </div>
    <div v-else class="empty-state">请选择左侧题目查看详情。</div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useImagePreview } from '../composables/useImagePreview'
import type { AnswerSubmitResponse, QuestionDetail } from '../types'

const { openPreview } = useImagePreview()

function onContentClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (target.tagName === 'IMG' && target instanceof HTMLImageElement) {
    e.preventDefault()
    e.stopPropagation()
    openPreview(target.src)
  }
}

const props = defineProps<{
  detail: QuestionDetail | null
  selectedAnswers: string[]
  submitResult: AnswerSubmitResponse | null
}>()

const emit = defineEmits<{
  (e: 'update:selectedAnswers', value: string[]): void
  (e: 'submit-answer'): void
  (e: 'toggle-favorite', value: boolean): void
  (e: 'toggle-difficult', value: boolean): void
  (e: 'save-note', value: string): void
}>()

const noteText = ref('')

watch(
  () => props.detail,
  (detail) => {
    noteText.value = detail?.note ?? ''
  },
  { immediate: true }
)

const isMultiple = computed(() => (props.detail?.answers.length ?? 0) > 1)

function decodeHtml(html: string): string {
  if (!html) return ''
  return html
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&amp;/g, '&')
}

function onOptionChange(optionLabel: string, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  if (isMultiple.value) {
    const next = checked
      ? [...props.selectedAnswers, optionLabel]
      : props.selectedAnswers.filter((item) => item !== optionLabel)
    emit('update:selectedAnswers', [...new Set(next)])
  } else {
    emit('update:selectedAnswers', checked ? [optionLabel] : [])
  }
  emit('submit-answer')
}
</script>
