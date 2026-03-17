<template>
  <Teleport to="body">
    <Transition name="img-preview">
      <div
        v-if="previewSrc"
        class="img-preview-overlay"
        @click.self="closePreview"
      >
        <button type="button" class="img-preview-close" aria-label="关闭" @click="closePreview">×</button>
        <img
          :src="previewSrc"
          class="img-preview-img"
          alt="预览"
          @click.stop
        />
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useImagePreview } from '../composables/useImagePreview'

const { previewSrc, closePreview } = useImagePreview()
</script>
