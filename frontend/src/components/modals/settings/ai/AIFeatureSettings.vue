<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhRobot, PhChatCircleText, PhTrash, PhBroom } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const isDeleting = ref(false);

async function clearAllChatSessions() {
  const confirmed = await window.showConfirm({
    title: t('clearAllChats'),
    message: t('clearAllChatsConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isDeleting.value = true;
  try {
    const response = await fetch('/api/ai/chat/sessions/delete-all', {
      method: 'DELETE',
    });

    if (response.ok) {
      const data = await response.json();
      window.showToast(t('clearAllChatsSuccess', { count: data.count || 0 }), 'success');
    } else {
      const errorText = await response.text();
      console.error('Server error:', response.status, errorText);
      window.showToast(t('clearAllChatsFailed'), 'error');
    }
  } catch (error) {
    console.error('Failed to clear chat sessions:', error);
    window.showToast(t('clearAllChatsFailed'), 'error');
  } finally {
    isDeleting.value = false;
  }
}
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhRobot :size="14" class="sm:w-4 sm:h-4" />
      {{ t('aiFeatures') }}
    </label>

    <!-- AI Chat -->
    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhChatCircleText :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('setting.ai.aiChatEnabled') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('setting.ai.aiChatEnabledDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="props.settings.ai_chat_enabled"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              ai_chat_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <!-- Chat Data Management (Sub-setting) -->
    <div
      v-if="props.settings.ai_chat_enabled"
      class="ml-2 sm:ml-4 mt-2 sm:mt-3 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
    >
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhTrash :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('clearAllChats') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('clearAllChatsDesc') }}
            </div>
          </div>
        </div>
        <button
          type="button"
          :disabled="isDeleting"
          class="btn-secondary"
          @click="clearAllChatSessions"
        >
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{ isDeleting ? t('cleaning') : t('clearAllChatsButton') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

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

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}

.setting-group {
  @apply space-y-2 sm:space-y-3;
}
</style>
