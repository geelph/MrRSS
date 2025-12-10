import { ref } from 'vue';

export function useArticleLabels() {
  const isGeneratingLabels = ref(false);

  /**
   * Generate labels for an article
   */
  async function generateLabels(articleId: number): Promise<string[]> {
    isGeneratingLabels.value = true;
    try {
      const response = await fetch('/api/label/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ article_id: articleId }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        if (errorData.error === 'missing_ai_api_key') {
          throw new Error('AI API key is required for AI-based labeling');
        }
        throw new Error('Failed to generate labels');
      }

      const data = await response.json();
      return data.labels || [];
    } finally {
      isGeneratingLabels.value = false;
    }
  }

  /**
   * Update labels for an article
   */
  async function updateLabels(articleId: number, labels: string[]): Promise<void> {
    const response = await fetch('/api/label/update', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ article_id: articleId, labels }),
    });

    if (!response.ok) {
      throw new Error('Failed to update labels');
    }
  }

  /**
   * Parse labels from JSON string
   */
  function parseLabels(labelsJson: string | undefined): string[] {
    if (!labelsJson) return [];
    try {
      const parsed = JSON.parse(labelsJson);
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }

  return {
    isGeneratingLabels,
    generateLabels,
    updateLabels,
    parseLabels,
  };
}
