<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArticleNyTimes, PhImages, PhPlayCircle } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

defineProps<{
  settings: SettingsData;
}>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();
</script>

<template>
  <div class="setting-section">
    <label class="section-label">
      <PhArticleNyTimes :size="16" class="w-4 h-4" />
      {{ t('contentSettings') }}
    </label>

    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhArticleNyTimes :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('enableFullTextFetch') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('enableFullTextFetchDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="settings.full_text_fetch_enabled"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              full_text_fetch_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <!-- Sub-setting: Auto Show All Content -->
    <div
      v-if="settings.full_text_fetch_enabled"
      class="mt-2 sm:mt-3 ml-4 sm:ml-6 pl-3 sm:pl-4 border-l-2 border-border"
    >
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhPlayCircle :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-5 sm:h-5" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-xs sm:text-sm">
              {{ t('autoShowAllContent') }}
            </div>
            <div class="text-[10px] sm:text-xs text-text-secondary hidden sm:block">
              {{ t('autoShowAllContentDesc') }}
            </div>
          </div>
        </div>
        <input
          :checked="settings.auto_show_all_content"
          type="checkbox"
          class="toggle"
          @change="
            (e) =>
              emit('update:settings', {
                ...settings,
                auto_show_all_content: (e.target as HTMLInputElement).checked,
              })
          "
        />
      </div>
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhImages :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('imageGalleryEnabled') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('imageGalleryEnabledDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="settings.image_gallery_enabled"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              image_gallery_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
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

.sub-setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary;
}
</style>
