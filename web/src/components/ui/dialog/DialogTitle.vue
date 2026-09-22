<script setup lang="ts">
import type { DialogTitleProps } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { reactiveOmit } from "@vueuse/core"
import { DialogTitle as DialogTitlePrimitive } from "reka-ui"
import { cn } from "@/lib/utils"

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes["class"] }>()
const delegatedProps = reactiveOmit(props, "class")
</script>

<template>
  <DialogTitlePrimitive
    data-slot="dialog-title"
    v-bind="delegatedProps"
    :class="cn('text-lg font-semibold leading-none', props.class)"
  >
    <slot />
  </DialogTitlePrimitive>
</template>
