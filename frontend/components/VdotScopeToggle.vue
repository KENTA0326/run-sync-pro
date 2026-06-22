<template>
  <div class="inline-flex rounded border border-gray-200 bg-white p-1 text-sm">
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="rounded px-3 py-1"
      :class="modelValue === option.value ? 'bg-blue-500 text-white' : 'text-gray-700 hover:bg-gray-100'"
      @click="select(option.value)"
    >
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
export type VdotTrendScope = 'all' | 'recent3'

const props = defineProps<{
  modelValue: VdotTrendScope
}>()

const emit = defineEmits<{
  'update:modelValue': [value: VdotTrendScope]
}>()

const options: { value: VdotTrendScope; label: string }[] = [
  { value: 'all', label: '全期間' },
  { value: 'recent3', label: '直近3ヶ月' },
]

function select(value: VdotTrendScope) {
  if (value !== props.modelValue) {
    emit('update:modelValue', value)
  }
}
</script>
