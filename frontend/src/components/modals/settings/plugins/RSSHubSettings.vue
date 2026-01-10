<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLink, PhKey, PhTestTube } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { useAppStore } from '@/stores/app';

const { t } = useI18n();
const store = useAppStore();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const isTesting = ref(false);

// Check if there are any RSSHub feeds
const hasRSSHubFeeds = computed(() => {
  return store.feeds && store.feeds.some((f) => f.url.startsWith('rsshub://'));
});

// Handle RSSHub toggle - prevent disabling if there are RSSHub feeds
function handleToggleRSSHub(e: Event) {
  const target = e.target as HTMLInputElement;
  const newValue = target.checked;

  // Prevent disabling if there are RSSHub feeds
  if (!newValue && hasRSSHubFeeds.value) {
    window.showToast(t('rsshubCannotDisableWithFeeds'), 'error');
    // Reset checkbox to enabled
    target.checked = true;
    return;
  }

  emit('update:settings', {
    ...props.settings,
    rsshub_enabled: newValue,
  });
}

// Test RSSHub connection
async function testConnection() {
  isTesting.value = true;

  try {
    const response = await fetch('/api/rsshub/test-connection', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        endpoint: props.settings.rsshub_endpoint,
        api_key: props.settings.rsshub_api_key,
      }),
    });

    const result = await response.json();

    if (result.success) {
      window.showToast(t('rsshubConnectionSuccessful'), 'success');
    } else {
      window.showToast(result.error || t('rsshubConnectionFailed'), 'error');
    }
  } catch (error) {
    window.showToast(error instanceof Error ? error.message : t('rsshubConnectionFailed'), 'error');
  } finally {
    isTesting.value = false;
  }
}
</script>

<template>
  <!-- Enable RSSHub -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/rsshub.svg"
        alt="RSSHub"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">{{ t('rsshubEnabled') }}</div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('rsshubEnabledDesc') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.rsshub_enabled"
      class="toggle"
      @change="handleToggleRSSHub"
    />
  </div>

  <div
    v-if="props.settings.rsshub_enabled"
    class="ml-2 sm:ml-4 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
  >
    <div class="tip-box">
      <PhInfo :size="16" class="text-accent shrink-0 sm:w-5 sm:h-5" />
      <span class="text-xs sm:text-sm">{{ t('rsshubNotSuggestOfficial') }}</span>
    </div>

    <!-- Endpoint -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhLink :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('rsshubEndpoint') }} <span class="text-red-500">*</span>
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubEndpointDesc') }}
          </div>
        </div>
      </div>
      <input
        type="text"
        :value="props.settings.rsshub_endpoint"
        placeholder="https://rsshub.app"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              rsshub_endpoint: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- API Key -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('rsshubAPIKey') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubAPIKeyDesc') }}
          </div>
        </div>
      </div>
      <input
        type="password"
        :value="props.settings.rsshub_api_key"
        :placeholder="t('rsshubOptional')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              rsshub_api_key: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- Test Connection -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhTestTube :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('rsshubTestConnection') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubTestConnectionDesc') }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <button :disabled="isTesting" class="btn-secondary" @click="testConnection">
          {{ isTesting ? t('rsshubTesting') : t('rsshubTestConnection') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.tip-box {
  @apply flex items-center gap-2 sm:gap-3 py-2 sm:py-2.5 px-2.5 sm:px-3 rounded-lg;
  background-color: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.3);
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

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}

.btn-secondary:disabled {
  @apply cursor-not-allowed opacity-50;
}
</style>
