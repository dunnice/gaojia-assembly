import { ref } from 'vue'

const previewSrc = ref<string | null>(null)

export function useImagePreview() {
  function openPreview(src: string) {
    previewSrc.value = src
  }

  function closePreview() {
    previewSrc.value = null
  }

  return { previewSrc, openPreview, closePreview }
}
