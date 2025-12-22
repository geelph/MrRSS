import { ref, type Ref } from 'vue';
import type { Feed } from '@/types/models';

export interface DropPreview {
  targetFeedId: number | null;
  beforeTarget: boolean; // true = insert before target, false = insert after target
}

export function useDragDrop() {
  const draggingFeedId: Ref<number | null> = ref(null);
  const dragOverCategory: Ref<string | null> = ref(null);
  const dropPreview: Ref<DropPreview> = ref({ targetFeedId: null, beforeTarget: true });

  function onDragStart(feedId: number, event: Event) {
    const dragEvent = event as DragEvent;
    draggingFeedId.value = feedId;
    if (dragEvent.dataTransfer) {
      dragEvent.dataTransfer.effectAllowed = 'move';
      dragEvent.dataTransfer.setData('text/plain', String(feedId));
    }
    // Add dragging class to the source element for visual feedback
    if (dragEvent.target instanceof HTMLElement) {
      const feedItem = dragEvent.target.closest('.feed-item');
      if (feedItem) {
        feedItem.classList.add('dragging');
      }
    }
    console.log('[onDragStart] Started dragging feed:', feedId);
  }

  function onDragEnd() {
    console.log('[onDragEnd] Ended dragging feed:', draggingFeedId.value);
    // Remove dragging class from all feed items
    document.querySelectorAll('.feed-item.dragging').forEach((el) => {
      el.classList.remove('dragging');
    });
    draggingFeedId.value = null;
    dragOverCategory.value = null;
    dropPreview.value = { targetFeedId: null, beforeTarget: true };
  }

  function onDragOver(category: string, targetFeedId: number | null, event: Event) {
    if (!event || !(event instanceof DragEvent)) {
      console.log('[onDragOver] Invalid event:', event);
      return;
    }

    event.preventDefault();

    if (!draggingFeedId.value) {
      console.log('[onDragOver] No dragging feed');
      return;
    }

    // Don't allow dropping on itself
    if (targetFeedId === draggingFeedId.value) {
      dropPreview.value = { targetFeedId: null, beforeTarget: true };
      console.log('[onDragOver] Dropping on itself, clearing preview');
      return;
    }

    dragOverCategory.value = category;

    // Calculate drop position based on mouse Y position relative to element
    let beforeTarget = true;
    if (targetFeedId !== null && event.target instanceof HTMLElement) {
      // Use target instead of currentTarget to get the actual element being hovered
      const target = event.target;
      // Get the feed-item element (might need to traverse up)
      const feedItem = target.closest('.feed-item');
      if (feedItem) {
        const rect = feedItem.getBoundingClientRect();
        const relativeY = event.clientY - rect.top;
        const threshold = rect.height / 2;
        beforeTarget = relativeY < threshold;

        // Debounce: only update if target or position changed significantly
        const newPreview = { targetFeedId, beforeTarget };
        if (
          dropPreview.value.targetFeedId !== newPreview.targetFeedId ||
          dropPreview.value.beforeTarget !== newPreview.beforeTarget
        ) {
          dropPreview.value = newPreview;
        }

        console.log(
          '[onDragOver] category:',
          category,
          'targetFeedId:',
          targetFeedId,
          'relativeY:',
          relativeY.toFixed(1),
          'threshold:',
          threshold.toFixed(1),
          'beforeTarget:',
          beforeTarget
        );
      } else {
        console.log('[onDragOver] Could not find .feed-item element');
      }
    } else {
      console.log('[onDragOver] No specific target, dropping at end. targetFeedId:', targetFeedId);
      // Only update if different
      if (dropPreview.value.targetFeedId !== null) {
        dropPreview.value = { targetFeedId: null, beforeTarget: true };
      }
    }

    console.log('[onDragOver] Updated dropPreview:', dropPreview.value);
  }

  function onDragLeave(category: string, event: Event) {
    if (!event || !(event instanceof DragEvent)) {
      return;
    }

    // Only clear the preview if we're actually leaving the category container
    // Check if the relatedTarget (where we're going) is outside the category
    const target = event.target as HTMLElement;
    const relatedTarget = event.relatedTarget as HTMLElement;

    // If moving to a child element, don't clear the preview
    if (relatedTarget && target.contains(relatedTarget)) {
      return;
    }

    // If moving from one category to another, the new category will handle it
    // Just clear when leaving the drag area entirely
    dropPreview.value = { targetFeedId: null, beforeTarget: true };
    console.log('[onDragLeave] Cleared preview for category:', category);
  }

  async function onDrop(
    currentCategory: string,
    feeds: Feed[]
  ): Promise<{ success: boolean; error?: string }> {
    if (!draggingFeedId.value) {
      return { success: false, error: 'No feed being dragged' };
    }

    const feedId = draggingFeedId.value;
    const targetCategory = dragOverCategory.value || currentCategory;
    const { targetFeedId, beforeTarget } = dropPreview.value;

    console.log('[onDrop] Starting drop operation:', {
      feedId,
      currentCategory,
      targetCategory,
      targetFeedId,
      beforeTarget,
      feedsCount: feeds.length,
    });

    // Sort feeds by position to get correct order
    const sortedFeeds = [...feeds].sort((a, b) => (a.position || 0) - (b.position || 0));

    // Find the dragging feed's current index (0-based)
    const draggingIndex = sortedFeeds.findIndex((f) => f.id === feedId);

    // Calculate the visual target index (0-based)
    // This is where the feed should appear in the sorted list
    let targetIndex = 0;

    if (targetFeedId !== null) {
      const targetIdx = sortedFeeds.findIndex((f) => f.id === targetFeedId);
      if (targetIdx !== -1) {
        if (beforeTarget) {
          // Insert before the target feed
          targetIndex = targetIdx;
        } else {
          // Insert after the target feed
          targetIndex = targetIdx + 1;
        }
      } else {
        // Target feed not found, append to end
        targetIndex = sortedFeeds.length;
      }
    } else {
      // No specific target, append to end
      targetIndex = sortedFeeds.length;
    }

    // Calculate the final position index considering the dragging feed will be removed
    let newPosition = targetIndex;
    if (draggingIndex !== -1 && targetIndex > draggingIndex) {
      // Moving forward: after removing the dragging feed, indices shift down by 1
      newPosition = targetIndex - 1;
    }

    console.log('[onDrop] Calculated position:', {
      feedId,
      targetCategory,
      newPosition,
      draggingIndex,
      targetIndex,
      targetFeedId,
      beforeTarget,
      feedsInCategory: sortedFeeds.length,
    });

    try {
      const response = await fetch('/api/feeds/reorder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          feed_id: feedId,
          category: targetCategory,
          position: newPosition,
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to reorder feed');
      }

      console.log('[onDrop] Successfully reordered feed');
      return { success: true };
    } catch (error) {
      console.error('[onDrop] Error:', error);
      return { success: false, error: error instanceof Error ? error.message : 'Unknown error' };
    }
  }

  return {
    draggingFeedId,
    dragOverCategory,
    dropPreview,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDragLeave,
    onDrop,
  };
}
