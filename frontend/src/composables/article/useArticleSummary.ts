import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';

interface SummarySettings {
  enabled: boolean;
  length: string;
  provider: string;
  triggerMode: string;
}

interface SummaryResult {
  summary: string;
  sentence_count: number;
  is_too_short: boolean;
  limit_reached?: boolean;
  used_fallback?: boolean;
  error?: string;
}

export function useArticleSummary() {
  const { t } = useI18n();
  const summarySettings = ref<SummarySettings>({
    enabled: false,
    length: 'medium',
    provider: 'local',
    triggerMode: 'auto',
  });
  const summaryCache: Ref<Map<number, SummaryResult>> = ref(new Map());
  const loadingSummaries: Ref<Set<number>> = ref(new Set());

  // Load summary settings
  async function loadSummarySettings(): Promise<void> {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      summarySettings.value = {
        enabled: data.summary_enabled === 'true',
        length: data.summary_length || 'medium',
        provider: data.summary_provider || 'local',
        triggerMode: data.summary_trigger_mode || 'manual',
      };
    } catch (e) {
      console.error('Error loading summary settings:', e);
    }
  }

  // Generate summary for an article
  async function generateSummary(
    article: Article,
    content?: string,
    force: boolean = false
  ): Promise<SummaryResult | null> {
    if (!summarySettings.value.enabled) {
      return null;
    }

    // If forcing regeneration, clear cache first
    if (force) {
      summaryCache.value.delete(article.id);
    }

    // Check if article already has a cached summary (only if not forcing)
    // This is the FIRST check - always return cached summary if available
    if (!force && article.summary && article.summary.trim() !== '') {
      const cachedResult: SummaryResult = {
        summary: article.summary,
        sentence_count: 0, // We don't store this in DB
        is_too_short: false,
      };
      // Store in cache but don't mark as loading since it's already done
      summaryCache.value.set(article.id, cachedResult);
      return cachedResult;
    }

    // Check in-memory cache (for summaries generated in current session)
    if (summaryCache.value.has(article.id)) {
      return summaryCache.value.get(article.id) || null;
    }

    // Check if already loading
    if (loadingSummaries.value.has(article.id)) {
      return null;
    }

    loadingSummaries.value.add(article.id);

    try {
      const res = await fetch('/api/articles/summarize', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          article_id: article.id,
          length: summarySettings.value.length,
          content: content,
        }),
      });

      if (res.ok) {
        const data: SummaryResult = await res.json();
        summaryCache.value.set(article.id, data);

        // Show notification if AI limit was reached
        if (data.limit_reached) {
          window.showToast(t('aiLimitReached'), 'warning');
        }

        return data;
      } else {
        // Handle API errors properly
        let errorMessage = `${t('summaryGenerationFailed')}: ${res.status} ${res.statusText}`;

        try {
          const errorData = await res.json();
          if (errorData.error) {
            errorMessage = errorData.error;
          }
        } catch (jsonError) {
          // If we can't parse JSON, use the status text
          console.error('Error parsing error response:', jsonError);
        }

        console.error('Summary generation failed:', errorMessage);

        // Cache the error to show in UI
        const errorResult: SummaryResult = {
          summary: '',
          sentence_count: 0,
          is_too_short: false,
          error: errorMessage,
        };
        summaryCache.value.set(article.id, errorResult);

        return errorResult;
      }
    } catch (e) {
      const errorMessage = `${t('summaryGenerationFailed')}: ${e instanceof Error ? e.message : t('unknownError')}`;
      console.error('Error generating summary:', e);

      // Cache the error to show in UI
      const errorResult: SummaryResult = {
        summary: '',
        sentence_count: 0,
        is_too_short: false,
        error: errorMessage,
      };
      summaryCache.value.set(article.id, errorResult);

      return errorResult;
    } finally {
      loadingSummaries.value.delete(article.id);
    }
  }

  // Get summary from cache
  function getCachedSummary(articleId: number): SummaryResult | null {
    return summaryCache.value.get(articleId) || null;
  }

  // Check if summary is loading
  function isSummaryLoading(articleId: number): boolean {
    return loadingSummaries.value.has(articleId);
  }

  // Clear cache for a specific article or all
  function clearSummaryCache(articleId?: number): void {
    if (articleId !== undefined) {
      summaryCache.value.delete(articleId);
    } else {
      summaryCache.value.clear();
    }
  }

  // Update summary settings from event
  function handleSummarySettingsChange(
    enabled: boolean,
    length: string,
    provider?: string,
    triggerMode?: string
  ): void {
    summarySettings.value = {
      enabled,
      length,
      provider: provider || summarySettings.value.provider,
      triggerMode: triggerMode || summarySettings.value.triggerMode,
    };
    // Clear cache when settings change to regenerate with new settings
    clearSummaryCache();
  }

  return {
    summarySettings,
    loadingSummaries,
    loadSummarySettings,
    generateSummary,
    getCachedSummary,
    isSummaryLoading,
    clearSummaryCache,
    handleSummarySettingsChange,
  };
}
