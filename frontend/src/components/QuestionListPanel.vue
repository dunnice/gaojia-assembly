<template>
  <section class="panel list-panel">
    <div class="panel-head">
      <div>
        <p class="eyebrow">{{ title }}</p>
        <h2>{{ subtitle }}</h2>
      </div>
      <slot name="actions" />
    </div>

    <div v-if="questions.length" class="question-cards">
      <button
        v-for="item in questions"
        :key="item.questionId"
        class="question-card"
        :class="{ active: item.questionId === selectedQuestionId }"
        @click="$emit('select', item.questionId)"
      >
        <div class="question-card-head">
          <span>#{{ item.questionIndex }}</span>
          <div class="badges">
            <span v-if="item.favorite" class="badge warm">收藏</span>
            <span v-if="item.difficult" class="badge hot">难题</span>
            <span v-if="item.wrongCount" class="badge danger">错 {{ item.wrongCount }}</span>
          </div>
        </div>
        <h3 v-html="item.titleHtml"></h3>
        <p>{{ item.showTypeName }} · {{ item.knowledge || '未标注知识点' }}</p>
        <small v-if="item.lastWrongAt">最近错题：{{ item.lastWrongAt }}</small>
      </button>
    </div>
    <div v-else class="empty-state">当前章节暂无符合条件的题目。</div>
  </section>
</template>

<script setup lang="ts">
import type { QuestionListItem } from '../types'

defineProps<{
  title: string
  subtitle: string
  questions: QuestionListItem[]
  selectedQuestionId: number | null
}>()

defineEmits<{
  (e: 'select', questionId: number): void
}>()
</script>
