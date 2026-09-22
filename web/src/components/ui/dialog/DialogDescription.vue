<script setup lang="ts">
import type { DialogDescriptionProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { reactiveOmit } from "@vueuse/core"
import { DialogDescription as DialogDescriptionPrimitive } from "reka-ui"
import { cn } from "@/lib/utils"

const props = defineProps<DialogDescriptionProps & { class?: HTMLAttributes["class"] }>()
const delegatedProps = reactiveOmit(props, "class")
</script>

<template>
  <DialogDescriptionPrimitive
    data-slot="dialog-description"
    v-bind="delegatedProps"
    :class="cn('text-sm text-zinc-400', props.class)"
  >
    <slot />
  </DialogDescriptionPrimitive>
</template>
