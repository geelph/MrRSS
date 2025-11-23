# Changelog

All notable changes to MrRSS will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.5] - 2025-11-23

### Added

- Content rendering toggle: Switch between viewing the original webpage and RSS content within the app
  - Toggle button in article detail view to switch between original webpage and rendered RSS content
  - Right-click menu option on articles to trigger content rendering
  - Styled RSS content display with proper typography and formatting
- Article hiding functionality:
  - Right-click menu option to hide/unhide articles
  - Hidden articles are excluded from the article list by default
  - Settings toggle to show/hide hidden articles
  - Visual indicator (eye-slash icon and dimmed styling) for hidden articles when shown
- Last article update time display in settings:
  - Shows when articles were last refreshed with relative time formatting (e.g., "2 hours ago")
  - Automatically updated after each refresh
- Translation clearing on settings change:
  - Automatically clears all translations when translation is enabled/disabled
  - Automatically clears all translations when target language is changed
  - Re-translates visible articles after clearing when translation is enabled

### Changed

- Improved UI text and image selection prevention:
  - All UI elements now prevent text selection by default
  - Images cannot be dragged
  - Input fields and textareas remain selectable for user interaction
- Settings auto-save now triggers article list refresh when "show hidden articles" toggle is changed

### Fixed

- Fixed issue where translation settings changes didn't clear existing translations

## [1.1.4] - 2025-11-23

### Added

- Auto cleanup sub-settings:
  - Max cache size setting (default 20MB) - controls maximum database size before cleanup
  - Max article age setting (default 30 days) - automatically delete articles older than specified days (except favorites)
- Download progress bar during update download
- App automatically closes after starting installer to prevent conflicts during update
- Automatic cleanup of installation packages after update installation

### Changed

- Settings now auto-save immediately when changed (no need to click save button)

### Removed

- "Save Settings" button at bottom of settings page (replaced with auto-save)

## [1.1.3] - 2025-11-22

### Added

- Automatically detects user's operating system and CPU architecture and downloads appropriate installer from GitHub releases. Then launches installer and prepares for update
- Multi-Platform Support:
  - Windows: x64 (amd64), ARM64
  - Linux: x64 (amd64), ARM64 (aarch64)
  - macOS: Universal (Intel & Apple Silicon)
- Visual feedback during update download and installation

## [1.1.2] - 2025-11-22

### Added

- Initial release preparation
- OPML import/export functionality
- Feed category organization
- Automatically detect and apply system theme preference
- Better defaults for translation settings
- Version check functionality in Settings → About tab

### Changed

- Simplified update check UI
- Improved theme switching mechanism
- Better handling of translation provider selection

### Fixed

- Various bug fixes and stability improvements
- UI refinements for better user experience
- Theme switching issues between light and dark modes
- Translation default language selection
- Update notification display

## [1.1.0] - 2025-11-20

### Added

- **Initial Public Release** of MrRSS
- **Cross-Platform Support**: Native desktop app for Windows, macOS, and Linux
- **RSS Feed Management**: Add, edit, and delete RSS feeds
- **Article Reading**: Clean, distraction-free reading interface
- **Smart Organization**: Organize feeds into categories
- **Favorites & Reading Tracking**: Save articles and track read/unread status
- **Modern UI**: Clean, responsive interface with dark mode support
- **Auto-Translation**: Translate article titles using Google Translate or DeepL API
- **OPML Support**: Import and export feed subscriptions
- **Auto-Update**: Configurable interval for fetching new articles
- **Database Cleanup**: Automatic removal of old articles
- **Multi-Language Support**: English and Chinese interface
- **Theme Support**: Light, dark, and auto (system) themes

---

## Release Notes

### Version Numbering

MrRSS follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Download

Downloads for all platforms are available on the [GitHub Releases](https://github.com/WCY-dt/MrRSS/releases) page.

### Upgrade Notes

When upgrading from a previous version:

1. Your data (feeds, articles, settings) is preserved in platform-specific directories
2. Database migrations are applied automatically on first launch
3. For major version upgrades, please review the changelog for breaking changes

### Support

- Report bugs: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- Feature requests: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- Documentation: [README](README.md)
