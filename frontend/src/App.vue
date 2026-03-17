<template>
  <div class="shell">
    <ImagePreviewModal />
    <aside class="topbar">
      <div class="topbar-left">
        <img src="/logo.png" alt="高架题库" class="topbar-logo" />
        <div>
          <p class="eyebrow">Ruankao Gaojia</p>
          <h1>高架题库学习系统</h1>
        </div>
      </div>
      <nav class="nav">
        <RouterLink to="/study" class="nav-link">章节刷题</RouterLink>
        <RouterLink to="/favorites" class="nav-link">收藏难题</RouterLink>
        <RouterLink to="/review" class="nav-link">错题复习</RouterLink>
      </nav>
    </aside>
    <div class="shell-body">
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import ImagePreviewModal from './components/ImagePreviewModal.vue'
</script>
