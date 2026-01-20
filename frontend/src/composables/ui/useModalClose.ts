import { onMounted, onUnmounted } from 'vue';

// Global modal stack to track nested modals
const modalStack: Array<{ zIndex: number; close: () => void }> = [];

export function useModalClose(onClose: () => void, modalZIndex?: number) {
  const zIndex = modalZIndex || 50; // Default z-index for modals

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      event.stopPropagation();

      // Find the modal with the highest z-index
      const highestModal = modalStack.reduce(
        (highest, modal) => {
          return modal.zIndex > (highest?.zIndex || 0) ? modal : highest;
        },
        null as { zIndex: number; close: () => void } | null
      );

      // Only close if this modal is the highest one
      if (highestModal && zIndex === highestModal.zIndex) {
        onClose();
      }
    }
  }

  onMounted(() => {
    modalStack.push({ zIndex, close: onClose });
    document.addEventListener('keydown', handleKeyDown);
  });

  onUnmounted(() => {
    const index = modalStack.findIndex((m) => m.zIndex === zIndex && m.close === onClose);
    if (index !== -1) {
      modalStack.splice(index, 1);
    }
    document.removeEventListener('keydown', handleKeyDown);
  });

  return {
    handleKeyDown,
  };
}
