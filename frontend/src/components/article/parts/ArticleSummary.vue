<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue';
import { PhTextAlignLeft, PhSpinnerGap, PhPlay, PhWarning, PhClock } from '@phosphor-icons/vue';
import { useI18n } from 'vue-i18n';

interface Props {
  summaryResult: {
    summary: string;
    sentence_count: number;
    is_too_short: boolean;
    error?: string;
  } | null;
  isLoadingSummary: boolean;
  translatedSummary: string;
  isTranslatingSummary: boolean;
  translationEnabled: boolean;
  summaryProvider?: string;
  summaryTriggerMode?: string;
  isLoadingContent?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  summaryProvider: 'local',
  summaryTriggerMode: 'auto',
  isLoadingContent: false,
});

const emit = defineEmits<{
  'generate-summary': [];
}>();

const { t } = useI18n();

const showSummary = ref(true);

// Enhanced loading states
const loadingTime = ref(0);
const loadingStartTime = ref<number | null>(null);

// Track loading time for better UX
watch(
  () => props.isLoadingSummary,
  (isLoading: boolean) => {
    if (isLoading && !loadingStartTime.value) {
      loadingStartTime.value = Date.now();
      loadingTime.value = 0;
    } else if (!isLoading && loadingStartTime.value) {
      loadingStartTime.value = null;
      loadingTime.value = 0;
    }
  }
);

let intervalId: number | null = null;

// Update loading time display
intervalId = window.setInterval(() => {
  if (props.isLoadingSummary && loadingStartTime.value) {
    loadingTime.value = Math.floor((Date.now() - loadingStartTime.value) / 1000);
  }
}, 100);

// Cleanup interval on unmount
onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId);
  }
});

// Check if should show manual trigger button
const shouldShowManualTrigger = computed(() => {
  return (
    props.summaryProvider === 'ai' &&
    props.summaryTriggerMode === 'manual' &&
    !props.summaryResult &&
    !props.isLoadingSummary &&
    !props.isLoadingContent
  );
});

// Generate summary on button click
function handleGenerateSummary() {
  emit('generate-summary');
}
</script>

<template>
  <!-- Summary Section -->
  <div
    v-if="summaryResult || isLoadingSummary || shouldShowManualTrigger"
    class="mb-6 p-4 rounded-lg border border-border bg-bg-secondary"
  >
    <!-- Summary Header -->
    <div
      class="flex items-center justify-between cursor-pointer"
      @click="showSummary = !showSummary"
    >
      <div class="flex items-center gap-2 text-accent font-medium">
        <PhTextAlignLeft :size="20" />
        <span>{{ t('articleSummary') }}</span>
      </div>
      <span class="text-xs text-text-secondary">
        {{ showSummary ? '▲' : '▼' }}
      </span>
    </div>

    <!-- Summary Content -->
    <div v-if="showSummary" class="mt-3">
      <!-- Enhanced Loading State -->
      <div v-if="isLoadingSummary" class="flex flex-col items-center gap-3 py-4">
        <!-- Loading Animation -->
        <div class="flex items-center gap-3">
          <PhSpinnerGap :size="24" class="animate-spin text-accent" />
          <div class="flex flex-col gap-1">
            <div class="text-sm font-medium text-text-primary">
              {{
                props.summaryProvider === 'ai' ? t('generatingAISummary') : t('generatingSummary')
              }}
            </div>
            <div class="flex items-center gap-2 text-xs text-text-secondary">
              <PhClock :size="12" />
              <span>{{ t('generatingSummaryTime', { seconds: loadingTime }) }}</span>
            </div>
          </div>
        </div>

        <!-- Progress Indicator -->
        <div class="w-full max-w-xs">
          <div class="text-xs text-text-secondary text-center mb-2">
            {{
              loadingTime < 3
                ? t('summaryInitiating')
                : loadingTime < 8
                  ? t('summaryProcessing')
                  : t('summaryAlmostDone')
            }}
          </div>
          <div class="w-full bg-bg-tertiary rounded-full h-1.5 overflow-hidden">
            <div
              class="bg-accent h-full transition-all duration-300 ease-out rounded-full"
              :style="{ width: `${Math.min(loadingTime * 12, 95)}%` }"
            />
          </div>
        </div>

        <!-- Provider-specific Tips -->
        <div
          v-if="props.summaryProvider === 'ai'"
          class="text-xs text-text-secondary text-center italic"
        >
          {{ t('summaryProcessingTip') }}
        </div>
      </div>

      <!-- Manual Trigger Button (only shown for AI manual mode when no summary) -->
      <div v-else-if="shouldShowManualTrigger" class="flex flex-col items-center gap-3 py-4">
        <div class="text-sm text-text-secondary text-center">
          {{ t('summaryManualTriggerDesc') }}
        </div>
        <button
          class="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-lg hover:bg-accent/90 transition-colors"
          @click="handleGenerateSummary"
        >
          <PhPlay :size="16" />
          <span class="text-sm font-medium">{{ t('generateSummary') }}</span>
        </button>
      </div>

      <!-- Too Short Warning -->
      <div v-else-if="summaryResult?.is_too_short" class="text-sm text-text-secondary italic">
        {{ t('summaryTooShort') }}
      </div>

      <!-- Summary Display -->
      <div v-else-if="summaryResult?.summary">
        <!-- Regenerate Button -->
        <div class="flex justify-end mb-2">
          <button
            class="flex items-center gap-1 px-2 py-1 text-xs bg-bg-secondary text-text-secondary rounded hover:bg-bg-tertiary transition-colors"
            :disabled="isLoadingSummary"
            @click="handleGenerateSummary"
          >
            <PhSpinnerGap v-if="isLoadingSummary" :size="12" class="animate-spin" />
            <PhPlay v-else :size="12" />
            <span>{{ t('regenerateSummary') }}</span>
          </button>
        </div>

        <!-- Show translated summary only when translation is enabled -->
        <div
          v-if="translationEnabled && translatedSummary"
          class="text-sm text-text-primary leading-relaxed select-text"
        >
          {{ translatedSummary }}
        </div>
        <!-- Show original summary when no translation or as fallback -->
        <p v-else class="text-sm text-text-primary leading-relaxed select-text">
          {{ summaryResult.summary }}
        </p>
        <!-- Translation loading indicator -->
        <div v-if="isTranslatingSummary" class="flex items-center gap-1 mt-2 text-text-secondary">
          <PhSpinnerGap :size="12" class="animate-spin" />
          <span class="text-xs">{{ t('translating') }}</span>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="summaryResult?.error" class="flex flex-col items-center gap-3 py-4">
        <div class="flex items-center gap-2 text-accent-error">
          <PhWarning :size="20" />
          <span class="text-sm font-medium">{{ t('summaryGenerationFailed') }}</span>
        </div>
        <div class="text-xs text-text-secondary text-center max-w-xs">
          {{ summaryResult.error }}
        </div>
        <!-- Retry Button for AI Summary -->
        <button
          v-if="props.summaryProvider === 'ai'"
          class="flex items-center gap-2 px-3 py-1.5 bg-accent text-white rounded-md hover:bg-accent/90 transition-colors text-xs"
          @click="handleGenerateSummary"
        >
          <PhPlay :size="14" />
          <span>{{ t('retrySummary') }}</span>
        </button>
      </div>

      <!-- No Summary Available -->
      <div v-else class="text-sm text-text-secondary italic">{{ t('noSummaryAvailable') }}</div>
    </div>
  </div>
</template>
