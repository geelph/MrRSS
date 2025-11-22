import { reactive, computed } from 'vue'
import { i18n } from './i18n.js'

export const store = reactive({
    articles: [],
    feeds: [],
    currentFilter: 'all', // 'all', 'unread', 'favorites'
    currentFeedId: null,
    currentCategory: null,
    currentArticleId: null,
    isLoading: false,
    page: 1,
    hasMore: true,
    searchQuery: '',
    themePreference: localStorage.getItem('themePreference') || 'auto', // 'light', 'dark', 'auto'
    theme: 'light', // actual applied theme
    i18n: i18n, // Make i18n available in store
    
    // Actions
    setFilter(filter) {
        this.currentFilter = filter;
        this.currentFeedId = null;
        this.currentCategory = null;
        this.page = 1;
        this.articles = [];
        this.hasMore = true;
        this.fetchArticles();
    },
    
    setFeed(feedId) {
        this.currentFilter = '';
        this.currentFeedId = feedId;
        this.currentCategory = null;
        this.page = 1;
        this.articles = [];
        this.hasMore = true;
        this.fetchArticles();
    },
    
    setCategory(category) {
        this.currentFilter = '';
        this.currentFeedId = null;
        this.currentCategory = category;
        this.page = 1;
        this.articles = [];
        this.hasMore = true;
        this.fetchArticles();
    },

    async fetchArticles(append = false) {
        if (this.isLoading) return;
        if (!append && !this.hasMore) this.hasMore = true; // Reset if new search
        
        this.isLoading = true;
        const limit = 50;
        
        let url = `/api/articles?page=${this.page}&limit=${limit}`;
        if (this.currentFilter) url += `&filter=${this.currentFilter}`;
        if (this.currentFeedId) url += `&feed_id=${this.currentFeedId}`;
        if (this.currentCategory) url += `&category=${encodeURIComponent(this.currentCategory)}`;
        
        try {
            const res = await fetch(url);
            const data = await res.json();
            const newArticles = data || [];
            
            if (newArticles.length < limit) {
                this.hasMore = false;
            }
            
            if (append) {
                this.articles = [...this.articles, ...newArticles];
            } else {
                this.articles = newArticles;
            }
        } catch (e) {
            console.error(e);
        } finally {
            this.isLoading = false;
        }
    },

    async loadMore() {
        if (this.hasMore && !this.isLoading) {
            this.page++;
            await this.fetchArticles(true);
        }
    },

    async fetchFeeds() {
        try {
            const res = await fetch('/api/feeds');
            const data = await res.json();
            this.feeds = data || [];
        } catch (e) {
            console.error(e);
            this.feeds = [];
        }
    },

    toggleTheme() {
        // Cycle through: light -> dark -> auto -> light
        if (this.themePreference === 'light') {
            this.themePreference = 'dark';
        } else if (this.themePreference === 'dark') {
            this.themePreference = 'auto';
        } else {
            this.themePreference = 'light';
        }
        localStorage.setItem('themePreference', this.themePreference);
        this.applyTheme();
    },

    setTheme(preference) {
        this.themePreference = preference;
        localStorage.setItem('themePreference', this.themePreference);
        this.applyTheme();
    },

    applyTheme() {
        let actualTheme = this.themePreference;
        
        // If auto, detect system preference
        if (this.themePreference === 'auto') {
            actualTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
        }
        
        this.theme = actualTheme;
        
        if (actualTheme === 'dark') {
            document.body.classList.add('dark-mode');
        } else {
            document.body.classList.remove('dark-mode');
        }
    },

    initTheme() {
        // Listen for system theme changes
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.addEventListener('change', () => {
            if (this.themePreference === 'auto') {
                this.applyTheme();
            }
        });
        
        // Apply initial theme
        this.applyTheme();
    },

    // Auto Refresh
    refreshInterval: null,
    refreshProgress: { current: 0, total: 0, isRunning: false },

    async refreshFeeds() {
        this.refreshProgress.isRunning = true;
        try {
            await fetch('/api/refresh', { method: 'POST' });
            this.pollProgress();
        } catch (e) {
            console.error(e);
            this.refreshProgress.isRunning = false;
        }
    },

    pollProgress() {
        let lastCurrent = 0;
        const interval = setInterval(async () => {
            try {
                const res = await fetch('/api/progress');
                const data = await res.json();
                this.refreshProgress = {
                    current: data.current,
                    total: data.total,
                    isRunning: data.is_running
                };

                // Progressive refresh: update articles whenever progress advances
                if (data.current > lastCurrent) {
                    lastCurrent = data.current;
                    this.fetchArticles();
                }

                if (!data.is_running) {
                    clearInterval(interval);
                    this.fetchFeeds();
                    this.fetchArticles();
                    
                    // Check for app updates after initial refresh completes
                    this.checkForAppUpdates();
                }
            } catch (e) {
                clearInterval(interval);
                this.refreshProgress.isRunning = false;
            }
        }, 500);
    },

    async checkForAppUpdates() {
        try {
            const res = await fetch('/api/check-updates');
            if (res.ok) {
                const data = await res.json();
                
                // Only proceed if there's an update available and a download URL
                if (data.has_update && data.download_url) {
                    console.log(`Update available: ${data.latest_version}`);
                    
                    // Show notification to user
                    window.showToast(
                        `${this.i18n.t('updateAvailable')}: v${data.latest_version}`,
                        'info',
                        5000
                    );
                    
                    // Optionally: Auto-download and install
                    // For now, just notify - user can manually update from Settings
                }
            }
        } catch (e) {
            console.error('Auto-update check failed:', e);
            // Silently fail - don't disrupt user experience
        }
    },

    startAutoRefresh(minutes) {
        if (this.refreshInterval) clearInterval(this.refreshInterval);
        if (minutes > 0) {
            this.refreshInterval = setInterval(() => {
                this.refreshFeeds();
            }, minutes * 60 * 1000);
        }
    }
});
