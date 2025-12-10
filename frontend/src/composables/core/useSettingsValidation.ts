/**
 * Composable for validating settings fields
 */
import { computed, type Ref } from 'vue';
import type { SettingsData } from '@/types/settings';

export function useSettingsValidation(settings: Ref<SettingsData>) {
  /**
   * Check if translation settings are valid
   */
  const isTranslationValid = computed(() => {
    if (!settings.value.translation_enabled) {
      return true; // Not enabled, so no validation needed
    }

    if (settings.value.translation_provider === 'deepl') {
      return !!settings.value.deepl_api_key?.trim();
    } else if (settings.value.translation_provider === 'baidu') {
      return !!(settings.value.baidu_app_id?.trim() && settings.value.baidu_secret_key?.trim());
    } else if (settings.value.translation_provider === 'ai') {
      return !!settings.value.ai_api_key?.trim();
    }

    return true; // Google Translate doesn't need API key
  });

  /**
   * Check if summary settings are valid
   */
  const isSummaryValid = computed(() => {
    if (!settings.value.summary_enabled) {
      return true; // Not enabled, so no validation needed
    }

    if (settings.value.summary_provider === 'ai') {
      return !!settings.value.summary_ai_api_key?.trim();
    }

    return true; // Local algorithm doesn't need API key
  });

  /**
   * Check if all settings are valid
   */
  const isValid = computed(() => {
    return isTranslationValid.value && isSummaryValid.value;
  });

  return {
    isTranslationValid,
    isSummaryValid,
    isValid,
  };
}
