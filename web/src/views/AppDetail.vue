<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Card, CardTitle, CardContent } from '@/components/ui/card'
import { ArrowLeft, Package, Calendar, CheckCircle2 } from 'lucide-vue-next'

interface VersionInfo {
  hash: string
  version: string
  install_time: string
}

interface AppDetail {
  name: string
  iconPath?: string
  currentHash?: string
  versions: VersionInfo[]
}

const router = useRouter()
const route = useRoute()

const appName = computed(() => decodeURIComponent(route.params.appName as string))
const app = ref<AppDetail | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    const res = await fetch(`/api/app/${encodeURIComponent(appName.value)}`)
    if (!res.ok) {
      if (res.status === 404) {
        throw new Error('App not found')
      }
      throw new Error('Failed to fetch app details')
    }
    app.value = await res.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    loading.value = false
  }
})

const goBack = () => {
  router.push('/')
}

const formatHash = (hash: string) => {
  return hash.substring(0, 8)
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

    <div class="relative z-10 max-w-3xl mx-auto py-8">
      <!-- Back button -->
      <Button
        variant="ghost"
        class="mb-6 text-zinc-400 hover:text-white -ml-2"
        @click="goBack"
      >
        <ArrowLeft class="w-4 h-4 mr-2" />
        Back to apps
      </Button>

      <!-- Loading state -->
      <div v-if="loading" class="animate-pulse space-y-6">
        <div class="flex items-center gap-4">
          <div class="w-20 h-20 bg-zinc-800 rounded-2xl"></div>
          <div class="space-y-2">
            <div class="w-48 h-8 bg-zinc-800 rounded"></div>
            <div class="w-32 h-4 bg-zinc-800 rounded"></div>
          </div>
        </div>
        <div class="space-y-3">
          <div v-for="i in 3" :key="i" class="h-16 bg-zinc-800 rounded-xl"></div>
        </div>
      </div>

      <!-- Error state -->
      <Card v-else-if="error" class="bg-zinc-900/60 border-zinc-800/50">
        <CardContent class="p-8 text-center">
          <Package class="w-12 h-12 text-zinc-600 mx-auto mb-4" />
          <CardTitle class="text-zinc-400 mb-2">{{ error }}</CardTitle>
          <Button variant="outline" class="mt-4" @click="goBack">
            Go back
          </Button>
        </CardContent>
      </Card>

      <!-- App details -->
      <template v-else-if="app">
        <!-- App header -->
        <div class="flex items-center gap-5 mb-8">
          <div class="w-20 h-20 rounded-2xl bg-zinc-950 border border-zinc-800 flex items-center justify-center overflow-hidden flex-shrink-0">
            <img
              v-if="app.iconPath"
              :src="`/api/icon/${encodeURIComponent(app.name)}`"
              :alt="app.name"
              class="w-14 h-14 object-contain"
              @error="handleImageError"
            />
            <Package v-else class="w-8 h-8 text-zinc-600" />
          </div>
          <div>
            <h1 class="text-3xl font-bold text-white mb-1">{{ app.name }}</h1>
            <p class="text-zinc-400">
              {{ app.versions.length }} version{{ app.versions.length !== 1 ? 's' : '' }} installed
            </p>
          </div>
        </div>

        <!-- Versions list -->
        <div class="space-y-3">
          <h2 class="text-sm font-medium text-zinc-500 uppercase tracking-wider mb-4">Versions</h2>

          <Card
            v-for="version in app.versions"
            :key="version.hash"
            class="bg-zinc-900/60 backdrop-blur-xl border-zinc-800/50"
            :class="{ 'border-indigo-500/30 bg-indigo-500/5': version.hash === app.currentHash }"
          >
            <CardContent class="p-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-4">
                  <!-- Active indicator -->
                  <div
                    v-if="version.hash === app.currentHash"
                    class="flex items-center gap-1.5 text-indigo-400"
                  >
                    <CheckCircle2 class="w-5 h-5" />
                    <span class="text-xs font-medium uppercase tracking-wide">Active</span>
                  </div>
                  <div v-else class="w-5"></div>

                  <div>
                    <div class="flex items-center gap-3">
                      <span class="font-medium text-white">
                        {{ version.version || 'Unknown version' }}
                      </span>
                      <code class="text-xs text-zinc-500 bg-zinc-950 px-2 py-0.5 rounded font-mono">
                        {{ formatHash(version.hash) }}
                      </code>
                    </div>
                    <div class="flex items-center gap-1.5 text-sm text-zinc-500 mt-1">
                      <Calendar class="w-3.5 h-3.5" />
                      <span>Installed {{ version.install_time }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </template>
    </div>
  </div>
</template>
