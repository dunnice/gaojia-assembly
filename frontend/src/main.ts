import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './styles.css'
import StudyView from './views/StudyView.vue'
import FavoriteView from './views/FavoriteView.vue'
import ReviewView from './views/ReviewView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/study' },
    { path: '/study', component: StudyView },
    { path: '/favorites', component: FavoriteView },
    { path: '/review', component: ReviewView }
  ]
})

createApp(App).use(router).mount('#app')
