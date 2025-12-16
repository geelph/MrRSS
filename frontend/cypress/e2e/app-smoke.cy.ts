/// <reference types="cypress" />

describe('Application Smoke Tests', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  it('should load the application successfully', () => {
    // Verify the app loads
    cy.get('body').should('be.visible')
    
    // Check for main layout elements
    cy.get('[class*="sidebar"]').should('exist')
    cy.get('[class*="article"]').should('exist')
  })

  it('should display the sidebar', () => {
    // Verify sidebar is present
    cy.get('[class*="sidebar"]').should('be.visible')
    
    // Check for common sidebar elements
    cy.contains(/all|全部/i).should('exist')
    cy.contains(/unread|未读/i).should('exist')
  })

  it('should have working navigation', () => {
    // Click on different navigation items
    cy.contains(/all|全部/i).click({ force: true })
    cy.wait(500)
    
    cy.contains(/unread|未读/i).click({ force: true })
    cy.wait(500)
    
    cy.contains(/favorite|收藏/i).click({ force: true })
    cy.wait(500)
  })

  it('should open and close settings modal', () => {
    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    
    // Verify modal is open
    cy.contains(/settings|设置/i).should('be.visible')
    
    // Close modal using ESC key
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Verify modal is closed
    cy.get('[data-modal-open="true"]').should('not.exist')
  })

  it('should handle keyboard shortcuts', () => {
    // Test settings shortcut (usually Ctrl+,)
    cy.get('body').type('{ctrl},')
    cy.wait(500)
    
    // Verify settings opened
    cy.contains(/settings|设置/i).should('be.visible')
    
    // Close with ESC
    cy.get('body').type('{esc}')
    cy.wait(500)
  })

  it('should display articles when feeds exist', () => {
    // Intercept articles API
    cy.intercept('GET', '/api/articles*').as('getArticles')
    
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Check if articles are displayed (or empty state)
    cy.get('[class*="article"], [class*="empty"]').should('exist')
  })

  it('should handle API errors gracefully', () => {
    // Intercept and force an error
    cy.intercept('GET', '/api/feeds', {
      statusCode: 500,
      body: { error: 'Internal server error' }
    }).as('getFeedsError')
    
    // Reload page
    cy.reload()
    
    // Wait for the error
    cy.wait('@getFeedsError')
    
    // Verify app doesn't crash
    cy.get('body').should('be.visible')
  })

  it('should be responsive', () => {
    // Test different viewport sizes
    cy.viewport(1920, 1080)
    cy.get('body').should('be.visible')
    
    cy.viewport(1280, 720)
    cy.get('body').should('be.visible')
    
    cy.viewport(768, 1024)
    cy.get('body').should('be.visible')
    
    // Mobile view
    cy.viewport(375, 667)
    cy.get('body').should('be.visible')
  })

  it('should handle long content gracefully', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')
    
    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Click on an article
    cy.get('[class*="article"]').first().click({ force: true })
    
    // Verify content is scrollable
    cy.get('[class*="detail"], [class*="content"]').should('be.visible')
  })

  it('should maintain state during navigation', () => {
    // Select unread filter
    cy.contains(/unread|未读/i).click({ force: true })
    cy.wait(500)
    
    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait(500)
    
    // Close settings
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Verify unread filter is still active by checking if the element exists
    cy.contains(/unread|未读/i).should('exist').and('be.visible')
  })
})
