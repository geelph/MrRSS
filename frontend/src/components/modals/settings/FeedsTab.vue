<script setup>
import { store } from '../../../store.js';
import { ref, computed } from 'vue';

const emit = defineEmits(['import-opml', 'export-opml', 'cleanup-database', 'add-feed', 'edit-feed', 'delete-feed', 'batch-delete', 'batch-move']);

const selectedFeeds = ref([]);

const isAllSelected = computed(() => {
    return store.feeds && store.feeds.length > 0 && selectedFeeds.value.length === store.feeds.length;
});

function toggleSelectAll(e) {
    if (!store.feeds) return;
    if (e.target.checked) {
        selectedFeeds.value = store.feeds.map(f => f.id);
    } else {
        selectedFeeds.value = [];
    }
}

function handleImportOPML(event) {
    emit('import-opml', event);
}

function handleExportOPML() {
    emit('export-opml');
}

function handleCleanupDatabase() {
    emit('cleanup-database');
}

function handleAddFeed() {
    emit('add-feed');
}

function handleEditFeed(feed) {
    emit('edit-feed', feed);
}

function handleDeleteFeed(id) {
    emit('delete-feed', id);
}

function handleBatchDelete() {
    if (selectedFeeds.value.length === 0) return;
    emit('batch-delete', selectedFeeds.value);
    selectedFeeds.value = [];
}

function handleBatchMove() {
    if (selectedFeeds.value.length === 0) return;
    emit('batch-move', selectedFeeds.value);
    selectedFeeds.value = [];
}
</script>

<template>
    <div class="space-y-6">
        <div class="setting-group">
            <label class="block font-semibold mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2">
                <i class="ph ph-hard-drives text-base"></i>
                {{ store.i18n.t('dataManagement') }}
            </label>
            <div class="flex gap-3 mb-3">
                <button @click="$refs.opmlInput.click()" class="btn-secondary flex-1 justify-center">
                    <i class="ph ph-upload"></i> {{ store.i18n.t('importOPML') }}
                </button>
                <input type="file" ref="opmlInput" class="hidden" @change="handleImportOPML">
                <button @click="handleExportOPML" class="btn-secondary flex-1 justify-center">
                    <i class="ph ph-download"></i> {{ store.i18n.t('exportOPML') }}
                </button>
            </div>
            <div class="flex gap-3">
                <button @click="handleCleanupDatabase" class="btn-danger flex-1 justify-center">
                    <i class="ph ph-broom"></i> {{ store.i18n.t('cleanDatabase') }}
                </button>
            </div>
            <p class="text-xs text-text-secondary mt-2">
                {{ store.i18n.t('cleanDatabaseDesc') }}
            </p>
        </div>
        
        <div class="setting-group">
            <label class="block font-semibold mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2">
                <i class="ph ph-rss text-base"></i>
                {{ store.i18n.t('manageFeeds') }}
            </label>
            
            <div class="flex flex-wrap gap-2 mb-2">
                <button @click="handleAddFeed" class="btn-secondary text-sm py-1.5 px-3">
                    <i class="ph ph-plus"></i> {{ store.i18n.t('addFeed') }}
                </button>
                <button @click="handleBatchDelete" class="btn-danger text-sm py-1.5 px-3" :disabled="selectedFeeds.length === 0">
                    <i class="ph ph-trash"></i> {{ store.i18n.t('deleteSelected') }}
                </button>
                <button @click="handleBatchMove" class="btn-secondary text-sm py-1.5 px-3" :disabled="selectedFeeds.length === 0">
                    <i class="ph ph-folder"></i> {{ store.i18n.t('moveSelected') }}
                </button>
                <div class="flex-1"></div>
                <label class="flex items-center gap-2 text-sm cursor-pointer select-none">
                    <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" class="w-4 h-4 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer">
                    {{ store.i18n.t('selectAll') }}
                </label>
            </div>

            <div class="border border-border rounded-lg bg-bg-secondary overflow-y-auto max-h-96">
                <div v-for="feed in store.feeds" :key="feed.id" class="flex items-center p-2 border-b border-border last:border-0 bg-bg-primary hover:bg-bg-secondary gap-2">
                    <input type="checkbox" :value="feed.id" v-model="selectedFeeds" class="w-4 h-4 shrink-0 rounded border-border text-accent focus:ring-2 focus:ring-accent cursor-pointer">
                    <div class="truncate flex-1 min-w-0">
                        <div class="font-medium truncate text-sm">{{ feed.title }}</div>
                        <div class="text-xs text-text-secondary truncate">{{ feed.url }}</div>
                    </div>
                    <div class="flex gap-1 shrink-0">
                        <button @click="handleEditFeed(feed)" class="text-accent hover:bg-bg-tertiary p-1 rounded text-sm" :title="store.i18n.t('edit')"><i class="ph ph-pencil"></i></button>
                        <button @click="handleDeleteFeed(feed.id)" class="text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 p-1 rounded text-sm" :title="store.i18n.t('delete')"><i class="ph ph-trash"></i></button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.btn-secondary {
    @apply bg-transparent border border-border text-text-primary px-4 py-2 rounded-md cursor-pointer flex items-center gap-2 font-medium hover:bg-bg-tertiary transition-colors;
}
.btn-secondary:disabled {
    @apply opacity-50 cursor-not-allowed;
}
.btn-danger {
    @apply bg-transparent border border-red-300 text-red-600 px-4 py-2 rounded-md cursor-pointer flex items-center gap-2 font-semibold hover:bg-red-50 dark:hover:bg-red-900/20 dark:border-red-400 dark:text-red-400 transition-colors;
}
.btn-danger:disabled {
    @apply opacity-50 cursor-not-allowed;
}
</style>
