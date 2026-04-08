<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'

interface AppMetadata {
  Hash: string
  AppName: string
  Version: string
  Desktop: string
  TmpDir: string
  AlreadyInstalled: boolean
}

const metadata = ref<AppMetadata | null>(null)
const loading = ref(true)
const actionStatus = ref<'idle' | 'installing' | 'completed' | 'cancelled'>('idle')

onMounted(async () => {
  try {
    const res = await fetch('/api/metadata')
    if (res.ok) {
      metadata.value = await res.json()
    }
  } catch (e) {
    console.error("Could not fetch metadata", e)
  } finally {
    loading.value = false
  }
})

const sendAction = async (action: 'install' | 'reinstall' | 'cancel') => {
  if (action === 'install' || action === 'reinstall') {
    actionStatus.value = 'installing'
  } else {
    actionStatus.value = 'cancelled'
  }

  try {
    const res = await fetch('/api/action', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action })
    })

    if (res.ok) {
      actionStatus.value = action === 'cancel' ? 'cancelled' : 'completed'
    }
  } catch (e) {
    console.error("Failed to set action", e)
    actionStatus.value = 'idle'
  }
}

const handleImageError = (e: Event) => {
  const target = e.target as HTMLImageElement
  // A generic box icon if the AppImage icon didn't load properly
  target.src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100' viewBox='0 0 24 24' fill='none' stroke='%236366f1' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z'/%3E%3Cpolyline points='3.27 6.96 12 12.01 20.73 6.96'/%3E%3Cline x1='12' y1='22.08' x2='12' y2='12'/%3E%3C/svg%3E"
  target.className = "w-24 h-24 object-contain rounded-2xl bg-zinc-900 border border-zinc-800 p-4"
}
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-50 flex items-center justify-center p-4 selection:bg-zinc-800">
    <!-- Subtle background glow -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none z-0">
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[120%] h-[120%] bg-gradient-to-br from-indigo-500/5 to-emerald-500/5 rounded-full blur-3xl"></div>
    </div>

    <!-- UI Container -->
    <div class="relative z-10 w-full max-w-sm">
      <div v-if="loading" class="flex flex-col items-center justify-center space-y-4 animate-pulse">
        <div class="w-20 h-20 bg-zinc-800 rounded-2xl"></div>
        <div class="w-32 h-6 bg-zinc-800 rounded-md"></div>
      </div>

      <Card
        v-else-if="metadata"
        class="bg-zinc-900/60 backdrop-blur-xl border-zinc-800/50 rounded-3xl p-6 sm:p-8"
      >
        <div v-if="actionStatus === 'idle'">
          <CardHeader class="items-center text-center p-0 mb-8">
            <div class="flex flex-col items-center gap-3">
              <img
                :src="'/api/icon'"
                :alt="metadata.AppName"
                class="w-24 h-24 object-contain rounded-2xl bg-zinc-900 border border-zinc-800 p-2"
                @error="handleImageError"
              />
              <p class="text-xs text-zinc-400 font-mono bg-zinc-950/50 px-3 py-1 rounded-full overflow-hidden text-ellipsis max-w-[200px]" title="Calculated Hash">
                {{ metadata.Hash.substring(0, 8) }}
              </p>
            </div>
            <CardTitle class="text-2xl font-bold tracking-tight text-white mb-2">{{ metadata.AppName }}</CardTitle>
            <p v-if="metadata.Version" class="text-sm text-zinc-400 font-medium">{{ metadata.Version }}</p>
          </CardHeader>

          <CardContent class="text-center p-0 mb-8 px-2">
            <template v-if="metadata.AlreadyInstalled">
              <CardDescription class="text-[15px] text-zinc-300 font-medium">Reinstall</CardDescription>
              <CardDescription class="text-sm text-zinc-500 mt-2 leading-relaxed">
                The exact same AppImage hash was found. Would you like to reinstall it?
              </CardDescription>
            </template>
            <template v-else>
              <CardDescription class="text-[15px] text-zinc-300 font-medium">Install</CardDescription>
              <CardDescription class="text-sm text-zinc-500 mt-2 leading-relaxed">
                This will extract the AppImage into your local directory.
              </CardDescription>
            </template>
          </CardContent>

          <CardFooter class="flex flex-col space-y-3 p-0">
            <Button
              v-if="metadata.AlreadyInstalled"
              @click="sendAction('reinstall')"
              class="w-full bg-white text-zinc-950 hover:bg-zinc-200 transition-colors py-6 text-base font-semibold rounded-xl"
            >
              Reinstall
            </Button>
            <Button
              v-else
              @click="sendAction('install')"
              class="w-full bg-white text-zinc-950 hover:bg-zinc-200 transition-colors py-6 text-base font-semibold rounded-xl"
            >
              Install
            </Button>

            <button
              @click="sendAction('cancel')"
              class="w-full py-4 text-sm font-medium text-zinc-400 hover:text-white transition-colors"
            >
              Cancel
            </button>
          </CardFooter>
        </div>

        <CardContent v-else-if="actionStatus === 'installing'" class="flex flex-col items-center justify-center py-10 p-0">
          <svg class="animate-spin h-10 w-10 text-indigo-500 mb-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <CardTitle class="text-lg font-semibold text-zinc-200">Installing {{ metadata.AppName }}...</CardTitle>
          <CardDescription class="text-sm text-zinc-500 mt-2">You can close this window now.</CardDescription>
        </CardContent>

        <CardContent v-else-if="actionStatus === 'completed'" class="flex flex-col items-center justify-center py-10 p-0">
          <div class="w-16 h-16 bg-emerald-500/10 rounded-full flex items-center justify-center mb-6">
            <svg class="w-8 h-8 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <CardTitle class="text-lg font-semibold text-emerald-400">Successfully Installed!</CardTitle>
          <CardDescription class="text-sm text-zinc-500 mt-2 text-center">The application has been installed. You may close this window.</CardDescription>
        </CardContent>

        <CardContent v-else-if="actionStatus === 'cancelled'" class="py-10 text-center p-0">
          <CardTitle class="text-lg font-medium text-zinc-400">Installation Cancelled.</CardTitle>
          <CardDescription class="text-sm text-zinc-500 mt-2">You may close this window safely.</CardDescription>
        </CardContent>

      </Card>

      <Card v-else class="text-center p-8 bg-zinc-900 border-zinc-800 rounded-3xl">
        <CardTitle class="text-red-400 font-semibold">Error Loading Metadata</CardTitle>
        <CardDescription class="text-zinc-500 text-sm mt-2">Could not connect to backend.</CardDescription>
      </Card>

    </div>
  </div>
</template>