<script setup lang="ts">
import { computed } from 'vue';
import { PhTag } from '@phosphor-icons/vue';

interface Props {
  labelsJson?: string;
  maxDisplay?: number;
  size?: 'sm' | 'md';
}

const props = withDefaults(defineProps<Props>(), {
  labelsJson: '[]',
  maxDisplay: 3,
  size: 'sm',
});

const labels = computed(() => {
  try {
    const parsed = JSON.parse(props.labelsJson);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
});

const displayedLabels = computed(() => {
  return labels.value.slice(0, props.maxDisplay);
});

const hiddenCount = computed(() => {
  return labels.value.length - props.maxDisplay;
});
</script>

<template>
  <div v-if="labels.length > 0" class="label-badges">
    <div class="label-badge" v-for="label in displayedLabels" :key="label" :class="size">
      <PhTag :size="size === 'sm' ? 10 : 12" class="shrink-0" />
      <span class="label-text">{{ label }}</span>
    </div>
    <div v-if="hiddenCount > 0" class="label-badge more-badge" :class="size">
      <span class="label-text">+{{ hiddenCount }}</span>
    </div>
  </div>
</template>

<style scoped>
.label-badges {
  @apply flex flex-wrap gap-1 items-center;
}

.label-badge {
  @apply flex items-center gap-0.5 px-1.5 py-0.5 rounded-full border border-border bg-bg-tertiary text-text-secondary;
  transition: all 0.2s ease;
}

.label-badge.sm {
  @apply text-[9px] sm:text-[10px];
}

.label-badge.md {
  @apply text-xs sm:text-sm px-2 py-1;
}

.label-badge:hover {
  @apply bg-accent/10 border-accent/30;
}

.label-text {
  @apply truncate max-w-[60px] sm:max-w-[100px];
}

.label-badge.md .label-text {
  @apply max-w-[80px] sm:max-w-[120px];
}

.more-badge {
  @apply bg-bg-secondary border-dashed;
}
</style>
