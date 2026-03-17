<template>
  <aside class="sidebar">
    <div class="sidebar-head">
      <h2>{{ title }}</h2>
      <p>{{ description }}</p>
    </div>
    <div class="chapter-list">
      <template v-for="(chapter, idx) in chapters" :key="chapter.chapterId">
        <div class="chapter-item-wrap">
          <button
            class="chapter-item"
            :class="{
              active: isSelected(chapter.chapterId, false),
              'whole-active': isSelected(chapter.chapterId, true)
            }"
            @click="onChapterClick(chapter, !!chapter.children?.length)"
          >
            <span v-if="!chapter.children?.length" class="expand-placeholder"></span>
            <button
              v-else
              type="button"
              class="expand-btn"
              :class="{ expanded: expandedIds.has(chapter.chapterId) }"
              @click.stop="toggleExpand(chapter.chapterId)"
              aria-label="展开/收起"
            />
            <div class="chapter-item-content">
              <div class="chapter-title-row">
                <span class="chapter-name">{{ formatChapterTitle(idx + 1, chapter.chapterName) }}</span>
                <strong>{{ chapter.allQuestionNum }}</strong>
              </div>
              <small v-if="chapter.children?.length" class="sub-hint">
                {{ chapter.children.length }} 个子章节
              </small>
            </div>
          </button>
          <template v-if="chapter.children?.length && expandedIds.has(chapter.chapterId)">
            <button
              class="chapter-item chapter-item-whole"
              :class="{ active: isSelected(chapter.chapterId, true) }"
              @click.stop="onChapterClick(chapter, true)"
            >
              <span class="expand-placeholder"></span>
              <div class="chapter-item-content">
                <span class="whole-label">整章</span>
                <strong>{{ chapter.allQuestionNum }}</strong>
              </div>
            </button>
            <button
              v-for="(child, childIdx) in chapter.children"
              :key="child.chapterId"
              class="chapter-item chapter-item-child"
              :class="{ active: isSelected(child.chapterId, false) }"
              @click.stop="onChapterClick(child, false)"
            >
              <span class="expand-placeholder"></span>
              <div class="chapter-item-content">
                <span class="chapter-name">第{{ childIdx + 1 }}节 {{ stripEdition(child.chapterName) }}</span>
                <strong>{{ child.allQuestionNum }}</strong>
              </div>
            </button>
          </template>
        </div>
      </template>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { ChapterNode } from '../types'
import { formatChapterTitle, stripEdition } from '../utils/chapter'

const props = defineProps<{
  title: string
  description: string
  chapters: ChapterNode[]
  selectedChapterId: number | null
  includeChildren?: boolean
}>()

const emit = defineEmits<{
  (e: 'select', chapterId: number, includeChildren: boolean): void
}>()

const expandedIds = ref<Set<number>>(new Set())

function toggleExpand(chapterId: number) {
  const next = new Set(expandedIds.value)
  if (next.has(chapterId)) {
    next.delete(chapterId)
  } else {
    next.add(chapterId)
  }
  expandedIds.value = next
}

function isSelected(chapterId: number, asWhole: boolean): boolean {
  if (props.selectedChapterId !== chapterId) return false
  if (asWhole) return props.includeChildren === true
  return props.includeChildren !== true
}

function onChapterClick(chapter: ChapterNode, includeChildren: boolean) {
  emit('select', chapter.chapterId, includeChildren)
}

watch(
  () => [props.chapters, props.selectedChapterId] as const,
  ([chapters, selectedId]) => {
    if (!chapters?.length || !selectedId) return
    for (const c of chapters) {
      if (c.chapterId === selectedId) {
        if (c.children?.length) {
          expandedIds.value = new Set([...expandedIds.value, c.chapterId])
        }
        return
      }
      if (c.children?.some((ch) => ch.chapterId === selectedId)) {
        expandedIds.value = new Set([...expandedIds.value, c.chapterId])
        return
      }
    }
  },
  { immediate: true }
)
</script>
