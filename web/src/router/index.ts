import { createRouter, createWebHashHistory } from 'vue-router'
import AppList from '@/views/AppList.vue'
import AppDetail from '@/views/AppDetail.vue'
import Installer from '@/views/Installer.vue'

const routes = [
  {
    path: '/',
    name: 'AppList',
    component: AppList
  },
  {
    path: '/installer',
    name: 'Installer',
    component: Installer
  },
  {
    path: '/app/:appName',
    name: 'AppDetail',
    component: AppDetail,
    props: true
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
