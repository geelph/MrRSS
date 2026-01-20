<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhTextT, PhTextIndent, PhTextAa } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { getRecommendedFonts } from '@/utils/fontDetector';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Font categories
const availableFonts = ref<{
  serif: string[];
  sansSerif: string[];
  monospace: string[];
}>({
  serif: [],
  sansSerif: [],
  monospace: [],
});

// Computed values for display (handle string/number conversion)
const displayContentSize = computed(() => {
  return parseInt(props.settings.content_font_size as any) || 16;
});
const displayLineHeight = computed(() => {
  return parseFloat(props.settings.content_line_height as any) || 1.6;
});

// Load system fonts on mount
onMounted(() => {
  try {
    availableFonts.value = getRecommendedFonts();
  } catch (error) {
    console.error('Failed to detect system fonts:', error);
  }
});
</script>

<template>
  <div class="setting-section">
    <label class="section-label">
      <PhTextT :size="16" class="w-4 h-4" />
      {{ t('typography') }}
    </label>

    <!-- Content Font Family -->
    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhTextT :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('contentFontFamily') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('contentFontFamilyDesc') }}
          </div>
        </div>
      </div>
      <select
        :value="settings.content_font_family"
        class="input-field w-36 sm:w-48 text-xs sm:text-sm max-h-60"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              content_font_family: (e.target as HTMLSelectElement).value,
            })
        "
      >
        <optgroup :label="t('fontSystem')">
          <option value="system">{{ t('fontSystemDefault') }}</option>
        </optgroup>

        <optgroup v-if="availableFonts.serif.length > 0" :label="t('fontSerif')">
          <option value="serif">{{ t('fontSerifDefault') }}</option>
          <option
            v-for="font in availableFonts.serif"
            :key="font"
            :value="font"
            :style="{ fontFamily: font + ', serif' }"
          >
            {{ font }}
          </option>
        </optgroup>

        <optgroup v-if="availableFonts.sansSerif.length > 0" :label="t('fontSansSerif')">
          <option value="sans-serif">{{ t('fontSansSerifDefault') }}</option>
          <option
            v-for="font in availableFonts.sansSerif"
            :key="font"
            :value="font"
            :style="{ fontFamily: font + ', sans-serif' }"
          >
            {{ font }}
          </option>
        </optgroup>

        <optgroup v-if="availableFonts.monospace.length > 0" :label="t('fontMonospace')">
          <option value="monospace">{{ t('fontMonospaceDefault') }}</option>
          <option
            v-for="font in availableFonts.monospace"
            :key="font"
            :value="font"
            :style="{ fontFamily: font + ', monospace' }"
          >
            {{ font }}
          </option>
        </optgroup>
      </select>
    </div>

    <!-- Content Font Size -->
    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhTextAa :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('contentFontSize') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('contentFontSizeDesc') }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <input
          type="number"
          :value="displayContentSize"
          class="input-field w-20 sm:w-24 text-xs sm:text-sm"
          @input="
            (e) => {
              const value = parseInt((e.target as HTMLInputElement).value);
              emit('update:settings', {
                ...settings,
                content_font_size: isNaN(value) ? 16 : value,
              });
            }
          "
        />
        <span class="text-sm text-text-secondary">px</span>
      </div>
    </div>

    <!-- Content Line Height -->
    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhTextIndent :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('contentLineHeight') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('contentLineHeightDesc') }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <input
          type="number"
          :value="displayLineHeight"
          step="0.1"
          class="input-field w-20 sm:w-24 text-xs sm:text-sm"
          @input="
            (e) => {
              const value = parseFloat((e.target as HTMLInputElement).value);
              emit('update:settings', {
                ...settings,
                content_line_height: isNaN(value) ? '1.6' : value.toString(),
              });
            }
          "
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.section-label {
  @apply font-semibold mb-3 sm:mb-4 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2;
}

.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}

.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}

.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}

.toggle:checked::after {
  transform: translateX(20px);
}

.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
</style>
