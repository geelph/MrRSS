<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, watch, onMounted, computed, nextTick } from 'vue';
import { PhArticle, PhTextAlignLeft, PhSpinnerGap, PhTranslate } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { formatDate } from '@/utils/date';
import { settingsDefaults } from '@/config/defaults';

const { t, locale } = useI18n();

interface SummaryResult {
  summary: string;
  sentence_count: number;
  is_too_short: boolean;
  error?: string;
}

interface Props {
  article: Article;
  articleContent: string;
  isLoadingContent: boolean;
}

const props = defineProps<Props>();

// Summary state
const summaryEnabled = ref(settingsDefaults.summary_enabled);
const summaryLength = ref(settingsDefaults.summary_length);
const summaryResult = ref<SummaryResult | null>(null);
const isLoadingSummary = ref(false);
const showSummary = ref(true);

// Translation state
const translationEnabled = ref(settingsDefaults.translation_enabled);
const targetLanguage = ref(settingsDefaults.target_language);
const translatedSummary = ref('');
const translatedTitle = ref('');
const isTranslatingSummary = ref(false);
const isTranslatingTitle = ref(false);
const isTranslatingContent = ref(false);

// Track last translated article ID to prevent duplicate translations
const lastTranslatedArticleId = ref<number | null>(null);

// Computed: check if we should show bilingual title
const showBilingualTitle = computed(() => {
  return (
    translationEnabled.value &&
    translatedTitle.value &&
    translatedTitle.value !== props.article?.title
  );
});

// Load summary and translation settings
async function loadSettings() {
  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    summaryEnabled.value = data.summary_enabled === 'true';
    summaryLength.value = data.summary_length || settingsDefaults.summary_length;
    translationEnabled.value = data.translation_enabled === 'true';
    targetLanguage.value = data.target_language || settingsDefaults.target_language;
  } catch (e) {
    console.error('Error loading settings:', e);
  }
}

// Translate text using the API
async function translateText(text: string): Promise<string> {
  if (!text || !translationEnabled.value) return '';

  try {
    const res = await fetch('/api/articles/translate-text', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        text: text,
        target_language: targetLanguage.value,
      }),
    });

    if (res.ok) {
      const data = await res.json();
      return data.translated_text || '';
    }
  } catch (e) {
    console.error('Error translating text:', e);
  }
  return '';
}

// Generate summary for the current article
async function generateSummary() {
  if (!summaryEnabled.value || !props.article) {
    return;
  }

  isLoadingSummary.value = true;
  summaryResult.value = null;
  translatedSummary.value = '';

  try {
    const res = await fetch('/api/articles/summarize', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        article_id: props.article.id,
        length: summaryLength.value,
      }),
    });

    if (res.ok) {
      summaryResult.value = await res.json();

      // Auto-translate summary if translation is enabled
      if (
        translationEnabled.value &&
        summaryResult.value?.summary &&
        !summaryResult.value.is_too_short
      ) {
        isTranslatingSummary.value = true;
        translatedSummary.value = await translateText(summaryResult.value.summary);
        isTranslatingSummary.value = false;
      }
    }
  } catch (e) {
    console.error('Error generating summary:', e);
  } finally {
    isLoadingSummary.value = false;
  }
}

// Translate title
async function translateTitle() {
  if (!translationEnabled.value || !props.article?.title) return;

  isTranslatingTitle.value = true;
  translatedTitle.value = await translateText(props.article.title);
  isTranslatingTitle.value = false;
}

// Tags that contain text to translate (only direct text blocks)
const textTags = ['P', 'H1', 'H2', 'H3', 'H4', 'H5', 'H6', 'LI', 'BLOCKQUOTE', 'FIGCAPTION'];

// Get visible text content for comparison (strips HTML but keeps structure)
function getVisibleText(el: HTMLElement): string {
  return el.textContent?.trim() || '';
}

// Translate content paragraph by paragraph and insert translations after each original
async function translateContentParagraphs() {
  if (!translationEnabled.value || !props.articleContent) return;

  // Prevent duplicate translations for the same article
  if (lastTranslatedArticleId.value === props.article?.id) {
    return;
  }

  isTranslatingContent.value = true;
  lastTranslatedArticleId.value = props.article?.id || null;

  // Wait for content to render
  await nextTick();

  // Find all text elements in the prose content
  const proseContainer = document.querySelector('.prose-content');
  if (!proseContainer) {
    isTranslatingContent.value = false;
    return;
  }

  // Remove any existing translations first
  const existingTranslations = proseContainer.querySelectorAll('.translation-text');
  existingTranslations.forEach((el) => el.remove());

  // Find all translatable elements (only top-level text elements)
  const elements = proseContainer.querySelectorAll(textTags.join(','));

  for (const el of elements) {
    const htmlEl = el as HTMLElement;

    // Skip if inside a translation element
    if (htmlEl.closest('.translation-text')) continue;

    // Skip if already has translation sibling
    if (htmlEl.nextElementSibling?.classList.contains('translation-text')) continue;

    // Get the visible text content for translation
    const visibleText = getVisibleText(htmlEl);
    if (!visibleText || visibleText.length < 2) continue;

    // Translate the visible text
    const translated = await translateText(visibleText);

    // Skip if translation is same as original or empty
    if (!translated || translated === visibleText) continue;

    // Create translation element with same tag type
    const translationEl = document.createElement(htmlEl.tagName.toLowerCase());
    translationEl.className = 'translation-text';
    translationEl.textContent = translated;

    // Copy blockquote styling
    if (htmlEl.tagName === 'BLOCKQUOTE') {
      translationEl.style.borderLeft = '4px solid var(--accent-color)';
      translationEl.style.paddingLeft = '1em';
      translationEl.style.fontStyle = 'italic';
    }

    // Insert translation after original
    htmlEl.parentNode?.insertBefore(translationEl, htmlEl.nextSibling);
  }

  isTranslatingContent.value = false;
}

// Watch for article changes and regenerate summary + translations
watch(
  () => props.article?.id,
  (newId, oldId) => {
    if (newId !== oldId) {
      summaryResult.value = null;
      translatedSummary.value = '';
      translatedTitle.value = '';
      lastTranslatedArticleId.value = null; // Reset translation tracking

      if (props.article) {
        if (summaryEnabled.value) {
          generateSummary();
        }
        if (translationEnabled.value) {
          translateTitle();
        }
      }
    }
  }
);

// Watch for content loading completion only
watch(
  () => props.isLoadingContent,
  (isLoading, wasLoading) => {
    if (wasLoading && !isLoading && props.article) {
      if (summaryEnabled.value) {
        generateSummary();
      }
      if (translationEnabled.value && lastTranslatedArticleId.value !== props.article.id) {
        nextTick(() => translateContentParagraphs());
      }
    }
  }
);

onMounted(async () => {
  await loadSettings();
  if (props.article) {
    if (summaryEnabled.value && props.articleContent) {
      generateSummary();
    }
    if (translationEnabled.value) {
      translateTitle();
      if (props.articleContent && !props.isLoadingContent) {
        nextTick(() => translateContentParagraphs());
      }
    }
  }
});
</script>

<template>
  <div class="flex-1 overflow-y-auto bg-bg-primary p-3 sm:p-6">
    <div class="max-w-3xl mx-auto bg-bg-primary">
      <!-- Title Section - Bilingual when translation enabled -->
      <div class="mb-3 sm:mb-4">
        <!-- Original Title -->
        <h1 class="text-xl sm:text-3xl font-bold leading-tight text-text-primary">
          {{ article.title }}
        </h1>
        <!-- Translated Title (shown below if different from original) -->
        <h2
          v-if="showBilingualTitle"
          class="text-base sm:text-xl font-medium leading-tight mt-2 text-text-secondary"
        >
          {{ translatedTitle }}
        </h2>
        <!-- Translation loading indicator for title -->
        <div v-if="isTranslatingTitle" class="flex items-center gap-1 mt-1 text-text-secondary">
          <PhSpinnerGap :size="12" class="animate-spin" />
          <span class="text-xs">{{ t('translating') }}</span>
        </div>
      </div>

      <div
        class="text-xs sm:text-sm text-text-secondary mb-4 sm:mb-6 flex flex-wrap items-center gap-2 sm:gap-4"
      >
        <span>{{ article.feed_title }}</span>
        <span class="hidden sm:inline">•</span>
        <span>{{ formatDate(article.published_at, locale === 'zh-CN' ? 'zh-CN' : 'en-US') }}</span>
        <span v-if="translationEnabled" class="flex items-center gap-1 text-accent">
          <PhTranslate :size="14" />
          <span class="text-xs">{{ t('autoTranslateEnabled') }}</span>
        </span>
      </div>

      <!-- Summary Section -->
      <div
        v-if="summaryEnabled && (isLoadingSummary || summaryResult)"
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
          <!-- Loading State -->
          <div v-if="isLoadingSummary" class="flex items-center gap-2 text-text-secondary">
            <PhSpinnerGap :size="16" class="animate-spin" />
            <span class="text-sm">{{ t('generatingSummary') }}</span>
          </div>

          <!-- Too Short Warning -->
          <div v-else-if="summaryResult?.is_too_short" class="text-sm text-text-secondary italic">
            {{ t('summaryTooShort') }}
          </div>

          <!-- Summary Display -->
          <div v-else-if="summaryResult?.summary">
            <!-- Show translated summary only when translation is enabled -->
            <div
              v-if="translationEnabled && translatedSummary"
              class="text-sm text-text-primary leading-relaxed"
            >
              {{ translatedSummary }}
            </div>
            <!-- Show original summary when no translation or as fallback -->
            <p v-else class="text-sm text-text-primary leading-relaxed">
              {{ summaryResult.summary }}
            </p>
            <!-- Translation loading indicator -->
            <div
              v-if="isTranslatingSummary"
              class="flex items-center gap-1 mt-2 text-text-secondary"
            >
              <PhSpinnerGap :size="12" class="animate-spin" />
              <span class="text-xs">{{ t('translating') }}</span>
            </div>
          </div>

          <!-- No Summary Available -->
          <div v-else class="text-sm text-text-secondary italic">
            {{ t('noSummaryAvailable') }}
          </div>
        </div>
      </div>

      <!-- Loading state with proper background -->
      <div
        v-if="isLoadingContent"
        class="flex flex-col items-center justify-center py-12 sm:py-16 bg-bg-primary"
      >
        <div class="relative mb-4 sm:mb-6">
          <!-- Outer pulsing ring -->
          <div
            class="absolute inset-0 rounded-full border-2 sm:border-4 border-accent animate-ping opacity-20"
          ></div>
          <!-- Middle spinning ring -->
          <div
            class="absolute inset-0 rounded-full border-2 sm:border-4 border-t-accent border-r-transparent border-b-transparent border-l-transparent animate-spin"
          ></div>
          <!-- Inner icon -->
          <div class="relative bg-bg-secondary rounded-full p-4 sm:p-6">
            <PhArticle :size="48" class="text-accent sm:w-16 sm:h-16" />
          </div>
        </div>
        <p class="text-base sm:text-lg font-medium text-text-primary mb-1 sm:mb-2">
          {{ t('loadingContent') }}
        </p>
        <p class="text-xs sm:text-sm text-text-secondary px-4 text-center">
          {{ t('fetchingArticleContent') }}
        </p>
      </div>

      <!-- Content display with inline translations -->
      <div v-else-if="articleContent">
        <div
          class="prose prose-sm sm:prose-lg max-w-none text-text-primary prose-content"
          v-html="articleContent"
        ></div>
        <!-- Translation loading indicator -->
        <div v-if="isTranslatingContent" class="flex items-center gap-2 mt-4 text-text-secondary">
          <PhSpinnerGap :size="16" class="animate-spin" />
          <span class="text-sm">{{ t('translatingContent') }}</span>
        </div>
      </div>

      <!-- No content available -->
      <div v-else class="text-center text-text-secondary py-6 sm:py-8">
        <PhArticle :size="48" class="mb-2 sm:mb-3 opacity-50 mx-auto sm:w-16 sm:h-16" />
        <p class="text-sm sm:text-base">{{ t('noContent') }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Prose styling for article content */
.prose {
  color: var(--text-primary);
}
.prose :deep(h1),
.prose :deep(h2),
.prose :deep(h3),
.prose :deep(h4),
.prose :deep(h5),
.prose :deep(h6) {
  color: var(--text-primary);
  font-weight: 600;
  margin-top: 1.5em;
  margin-bottom: 0.75em;
}
.prose :deep(p) {
  margin-bottom: 1em;
  line-height: 1.7;
}
.prose :deep(a) {
  color: var(--accent-color);
  text-decoration: underline;
  cursor: pointer;
}
.prose :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.5rem;
  margin: 1.5em 0;
  cursor: pointer;
  transition: opacity 0.2s;
}
.prose :deep(img:hover) {
  opacity: 0.9;
}
.prose :deep(pre) {
  background-color: var(--bg-secondary);
  padding: 1em;
  border-radius: 0.5rem;
  overflow-x: auto;
  margin: 1em 0;
}
.prose :deep(code) {
  background-color: var(--bg-secondary);
  padding: 0.2em 0.4em;
  border-radius: 0.25rem;
  font-size: 0.9em;
}
.prose :deep(blockquote) {
  border-left: 4px solid var(--accent-color);
  padding-left: 1em;
  margin: 1em 0;
  font-style: italic;
  color: var(--text-secondary);
}
.prose :deep(ul),
.prose :deep(ol) {
  margin: 1em 0;
  padding-left: 2em;
}
.prose :deep(li) {
  margin-bottom: 0.5em;
}

/* Translation text styling - lighter color */
.prose :deep(.translation-text) {
  color: var(--text-secondary);
  opacity: 0.85;
  margin-top: 0.25em;
  margin-bottom: 1em;
  font-size: 0.95em;
  line-height: 1.6;
}

/* Blockquote translation preserves blockquote style */
.prose :deep(blockquote.translation-text) {
  border-left: 4px solid var(--accent-color);
  padding-left: 1em;
  font-style: italic;
  color: var(--text-secondary);
  opacity: 0.75;
}

/* Audio player styling - podcast support */
.prose :deep(audio) {
  width: 100%;
  max-width: 100%;
  margin: 1em 0;
  border-radius: 0.5rem;
  background-color: var(--bg-secondary);
}

/* Video player styling */
.prose :deep(video) {
  width: 100%;
  max-width: 100%;
  height: auto;
  margin: 1.5em 0;
  border-radius: 0.5rem;
  background-color: var(--bg-secondary);
}

/* Iframe for embedded videos (YouTube, Vimeo, etc.) */
.prose :deep(iframe) {
  width: 100%;
  max-width: 100%;
  aspect-ratio: 16 / 9;
  margin: 1.5em 0;
  border-radius: 0.5rem;
  border: none;
  background-color: var(--bg-secondary);
}

/* Figure and figcaption styling for media */
.prose :deep(figure) {
  margin: 1.5em 0;
}

.prose :deep(figcaption) {
  font-size: 0.875em;
  color: var(--text-secondary);
  text-align: center;
  margin-top: 0.5em;
}

/* Enclosure media container styling (RSS media enclosures) */
.prose :deep(.media-enclosure) {
  margin: 1.5em 0;
  padding: 1em;
  background-color: var(--bg-secondary);
  border-radius: 0.5rem;
  border: 1px solid var(--border-color);
}

/* Table styling for better rich text */
.prose :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5em 0;
  font-size: 0.9em;
}

.prose :deep(th),
.prose :deep(td) {
  padding: 0.75em;
  border: 1px solid var(--border-color);
  text-align: left;
}

.prose :deep(th) {
  background-color: var(--bg-secondary);
  font-weight: 600;
}

.prose :deep(tr:nth-child(even)) {
  background-color: var(--bg-secondary);
}

/* Horizontal rule styling */
.prose :deep(hr) {
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 2em 0;
}

/* Definition list styling */
.prose :deep(dl) {
  margin: 1em 0;
}

.prose :deep(dt) {
  font-weight: 600;
  margin-top: 0.5em;
}

.prose :deep(dd) {
  margin-left: 1.5em;
  color: var(--text-secondary);
}
</style>
