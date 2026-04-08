<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Card, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
import { Package, PackageOpen, ArrowRight } from 'lucide-vue-next'

interface AppInfo {
  name: string
  iconPath?: string
  currentHash?: string
  versionCount: number
}

const router = useRouter()
const apps = ref<AppInfo[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    const res = await fetch('/api/app')
    if (!res.ok) {
      throw new Error('Failed to fetch apps')
    }
    apps.value = await res.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    loading.value = false
  }
})

const navigateToApp = (appName: string) => {
  router.push(`/app/${encodeURIComponent(appName)}`)
}

const handleImageError = (e: Event) => {
  const target = e.target as HTMLImageElement
  target.style.display = 'none'
}
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-50 p-4 selection:bg-zinc-800">
    <!-- Background glow -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none z-0">
      <div class="absolute top-0 left-1/2 -translate-x-1/2 w-[80%] h-[50%] bg-gradient-to-b from-indigo-500/5 to-transparent rounded-full blur-3xl"></div>
    </div>

    <div class="relative z-10 max-w-4xl mx-auto py-8">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold tracking-tight text-white mb-2">App Manager</h1>
        <p class="text-zinc-400">Manage your installed AppImage applications</p>
      </div>

      <!-- Loading state -->
      <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="i in 3" :key="i" class="animate-pulse">
          <Card class="bg-zinc-900/60 border-zinc-800/50 h-32">
            <CardContent class="p-6 flex items-center gap-4">
              <div class="w-12 h-12 bg-zinc-800 rounded-xl"></div>
              <div class="flex-1 space-y-2">
                <div class="w-32 h-5 bg-zinc-800 rounded"></div>
                <div class="w-20 h-4 bg-zinc-800 rounded"></div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <!-- Error state -->
      <Card v-else-if="error" class="bg-zinc-900/60 border-zinc-800/50">
        <CardContent class="p-8 text-center">
          <PackageOpen class="w-12 h-12 text-zinc-600 mx-auto mb-4" />
          <CardTitle class="text-zinc-400 mb-2">Failed to load apps</CardTitle>
          <CardDescription class="text-zinc-500">{{ error }}</CardDescription>
        </CardContent>
      </Card>

      <!-- Empty state -->
      <Card v-else-if="apps.length === 0" class="bg-zinc-900/60 border-zinc-800/50">
        <CardContent class="p-12 text-center">
          <PackageOpen class="w-16 h-16 text-zinc-600 mx-auto mb-4" />
          <CardTitle class="text-xl text-zinc-300 mb-2">Empty</CardTitle>
          <CardDescription class="text-zinc-500 max-w-sm mx-auto">
            You haven't installed any AppImage applications yet.
          </CardDescription>
        </CardContent>
      </Card>

      <!-- App list -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card
          v-for="app in apps"
          :key="app.name"
          class="bg-zinc-900/60 backdrop-blur-xl border-zinc-800/50 hover:border-zinc-700/50 transition-all cursor-pointer group"
          @click="navigateToApp(app.name)"
        >
          <CardContent class="p-4 flex items-center gap-4">
            <!-- App icon -->
            <div class="w-14 h-14 rounded-xl bg-zinc-950 border border-zinc-800 flex items-center justify-center overflow-hidden flex-shrink-0">
              <img
                v-if="app.iconPath"
                :src="`/api/icon/${encodeURIComponent(app.name)}`"
                :alt="app.name"
                class="w-10 h-10 object-contain"
                @error="handleImageError"
              />
              <Package v-else class="w-6 h-6 text-zinc-600" />
            </div>

            <!-- App info -->
            <div class="flex-1 min-w-0">
              <h3 class="font-semibold text-white truncate group-hover:text-indigo-400 transition-colors">
                {{ app.name }}
              </h3>
              <p class="text-sm text-zinc-500">
                {{ app.versionCount }} version{{ app.versionCount !== 1 ? 's' : '' }}
              </p>
            </div>

            <!-- Arrow -->
            <ArrowRight class="w-5 h-5 text-zinc-600 group-hover:text-zinc-400 transition-colors flex-shrink-0" />
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
