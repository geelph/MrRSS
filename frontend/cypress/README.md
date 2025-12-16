# Cypress E2E Tests for MrRSS

This directory contains end-to-end (E2E) tests for the MrRSS frontend application using Cypress.

## Test Structure

```
cypress/
├── e2e/                      # E2E test files
│   ├── app-smoke.cy.ts      # Basic smoke tests
│   ├── settings-persistence.cy.ts  # Settings persistence tests
│   ├── feed-management.cy.ts       # Feed management tests
│   ├── article-operations.cy.ts    # Article operations tests
│   └── theme-and-language.cy.ts    # Theme and language switching tests
├── fixtures/                # Test data
│   └── example.json
├── support/                 # Support files and custom commands
│   ├── commands.ts         # Custom Cypress commands
│   ├── e2e.ts             # E2E test setup
│   └── component.ts       # Component test setup
└── README.md              # This file
```

## Running Tests

### Prerequisites

1. Install dependencies:
   ```bash
   cd frontend
   npm install
   ```

2. Start the MrRSS backend server:
   ```bash
   # From the project root
   wails dev
   # or
   go run main.go
   ```

### Running Tests Locally

**Interactive Mode (with Cypress GUI):**
```bash
cd frontend
npm run cypress
# or
npm run test:e2e:headed
```

**Headless Mode (for CI/CD):**
```bash
cd frontend
npm run cypress:headless
# or
npm run test:e2e
```

### Running from Project Root

Using Make:
```bash
make test-frontend-e2e
```

## Test Coverage

The E2E tests cover the following key user flows:

### 1. Settings Persistence (`settings-persistence.cy.ts`)
- Theme changes persist after closing and reopening settings
- Language changes persist across sessions
- Update interval changes are saved
- Multiple settings can be changed in sequence
- Settings save when switching between tabs

### 2. Feed Management (`feed-management.cy.ts`)
- Adding new feeds
- Deleting feeds
- Refreshing feeds
- Editing feed details
- Filtering feeds by category
- Searching feeds

### 3. Article Operations (`article-operations.cy.ts`)
- Marking articles as read
- Marking articles as favorite
- Filtering articles by read status
- Filtering articles by favorites
- Marking all articles as read
- Opening article detail view
- Searching articles
- Opening articles in external browser

### 4. Theme and Language (`theme-and-language.cy.ts`)
- Switching between light and dark themes
- Theme persistence after page reload
- Switching between languages (English/Chinese)
- Language persistence after page reload
- System theme preference
- Theme applied to all components

### 5. Smoke Tests (`app-smoke.cy.ts`)
- Application loads successfully
- Sidebar displays correctly
- Navigation works
- Settings modal opens and closes
- Keyboard shortcuts work
- Articles display when feeds exist
- API errors handled gracefully
- Responsive design works
- Long content handled gracefully
- State maintained during navigation

## Custom Commands

The tests use custom Cypress commands defined in `support/commands.ts`:

- `cy.openSettings()` - Opens the settings modal
- `cy.closeModal()` - Closes the current modal
- `cy.waitForApi(endpoint, alias)` - Waits for an API response
- `cy.mockApi(endpoint, response)` - Mocks an API response

## Configuration

The Cypress configuration is in `cypress.config.ts`:

- **Base URL**: `http://localhost:34115` (default Wails dev server port)
- **Viewport**: 1280x720
- **Timeout**: 10 seconds for commands, requests, and responses
- **Screenshots**: Enabled on failure
- **Videos**: Disabled by default

## CI/CD Integration

The E2E tests are integrated into the GitHub Actions workflow:

1. The backend server is started before running tests
2. Tests run in headless mode
3. Screenshots and videos are uploaded as artifacts on failure
4. The backend server is stopped after tests complete

See `.github/workflows/test.yml` for the complete CI/CD configuration.

## Writing New Tests

When adding new E2E tests:

1. Create a new file in `cypress/e2e/` with a descriptive name
2. Use the `.cy.ts` extension
3. Follow the existing test patterns
4. Use custom commands where appropriate
5. Add proper `beforeEach` hooks to ensure clean state
6. Intercept API calls to verify backend communication
7. Use meaningful test descriptions
8. Add assertions to verify expected behavior

Example test structure:

```typescript
/// <reference types="cypress" />

describe('Feature Name', () => {
  beforeEach(() => {
    cy.visit('/')
    cy.get('body').should('be.visible')
    cy.wait(1000)
  })

  it('should do something', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/endpoint').as('getEndpoint')
    
    // Perform actions
    cy.get('button').click()
    
    // Wait for API
    cy.wait('@getEndpoint')
    
    // Assert expected behavior
    cy.contains('Expected Result').should('exist')
  })
})
```

## Troubleshooting

### Tests Fail with "Connection Refused"

Make sure the backend server is running on port 34115:
```bash
go run main.go
```

### Tests Timeout

Increase the timeout in `cypress.config.ts` or use `cy.wait()` to give more time for slow operations.

### Selectors Not Found

The tests use flexible selectors that work with both English and Chinese text. If the UI structure changes, you may need to update the selectors in the test files.

### CI/CD Failures

Check the uploaded artifacts (screenshots and videos) in the GitHub Actions workflow run to see what failed.

## Future Improvements

- Add component tests for individual Vue components
- Add visual regression testing
- Add accessibility (a11y) testing
- Add performance testing
- Add network condition testing (slow 3G, offline, etc.)
- Add more comprehensive error handling tests
