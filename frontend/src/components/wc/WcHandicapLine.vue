<template>
  <span v-if="handicapText" class="wc-handicap-line">{{ handicapText }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  homeTeam: string
  awayTeam: string
  handicapValue: number | null | undefined
  handicapTeam: string | null | undefined
}>()

const handicapText = computed(() => {
  if (!props.handicapValue || !props.handicapTeam) return null
  if (props.handicapTeam === 'home') {
    return `${props.homeTeam} chấp ${props.awayTeam} ${props.handicapValue} trái`
  }
  if (props.handicapTeam === 'away') {
    return `${props.awayTeam} chấp ${props.homeTeam} ${props.handicapValue} trái`
  }
  return null
})
</script>

<style scoped>
.wc-handicap-line {
  display: block;
  font-size: 11px;
  color: var(--text-muted, #9ca3af);
  margin-top: 2px;
}
</style>
