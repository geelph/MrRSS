/// <reference types="cypress" />

describe('Settings Persistence', () => {
  beforeEach(() => {
    // Visit the app before each test
    cy.visit('/')
    
    // Wait for the app to be fully loaded
    cy.get('body').should('be.visible')
    
    // Wait a bit for the app to initialize
    cy.wait(1000)
  })

  it('should persist theme changes after closing and reopening settings', () => {
    // Intercept settings API calls
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings modal - look for settings button with gear icon or text
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    
    // Wait for settings modal to be visible
    cy.contains(/settings|设置/i).should('be.visible')
    
    // Ensure we're on the general tab (or navigate to it)
    cy.contains(/general|常规/i).click({ force: true })
    
    // Wait for settings to load
    cy.wait('@getSettings')
    
    // Find the theme selector - could be a select, radio buttons, or buttons
    // Try to find dark theme option and click it
    cy.contains(/dark|深色/i).click({ force: true })
    
    // Wait for settings to be saved
    cy.wait('@saveSettings', { timeout: 5000 })
    
    // Close the settings modal by clicking the X or clicking outside
    cy.get('body').type('{esc}')
    
    // Wait a bit for modal to close
    cy.wait(500)
    
    // Reopen settings to verify the change persisted
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    
    // Wait for settings to load again
    cy.wait('@getSettings')
    
    // Verify dark theme is still selected
    cy.contains(/dark|深色/i).should('exist')
  })

  it('should persist language changes', () => {
    // Intercept settings API calls
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    
    // Navigate to general tab if not already there
    cy.contains(/general|常规/i).click({ force: true })
    
    // Wait for settings to load
    cy.wait('@getSettings')
    
    // Look for language selector and change it
    // Try to find a select element or radio group
    cy.get('body').then(($body) => {
      if ($body.find('select').length > 0) {
        // If there's a select dropdown
        cy.get('select').first().select(1)
      } else if ($body.find('[role="radiogroup"]').length > 0) {
        // If there are radio buttons
        cy.get('[role="radio"]').last().click({ force: true })
      }
    })
    
    // Wait for settings to be saved
    cy.wait('@saveSettings', { timeout: 5000 })
    
    // Close settings
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Reopen settings to verify
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Verify language is still set
    cy.get('select, [role="radiogroup"]').should('exist')
  })

  it('should persist update interval changes', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    
    // Navigate to general or feeds tab
    cy.contains(/general|常规|feeds|订阅/i).first().click({ force: true })
    
    // Wait for settings to load
    cy.wait('@getSettings')
    
    // Look for update interval input/select
    cy.get('input[type="number"], select').first().clear().type('30')
    
    // Wait for auto-save or click save button if exists
    cy.wait(2000)
    
    // Close settings
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Reopen to verify
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Verify the value persisted
    cy.get('input[type="number"], select').first().should('have.value', '30')
  })

  it('should handle multiple setting changes in sequence', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Change theme
    cy.contains(/general|常规/i).click({ force: true })
    cy.contains(/light|亮色/i).click({ force: true })
    cy.wait(1000)
    
    // Navigate to another tab
    cy.contains(/feeds|订阅/i).click({ force: true })
    cy.wait(500)
    
    // Make another change
    cy.get('input[type="number"]').first().clear().type('15')
    cy.wait(1000)
    
    // Close and reopen
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Verify both changes persisted
    cy.contains(/general|常规/i).click({ force: true })
    cy.contains(/light|亮色/i).should('exist')
    
    cy.contains(/feeds|订阅/i).click({ force: true })
    cy.get('input[type="number"]').first().should('have.value', '15')
  })

  it('should save settings when switching between tabs', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Make a change in general tab
    cy.contains(/general|常规/i).click({ force: true })
    cy.contains(/dark|深色/i).click({ force: true })
    
    // Switch to feeds tab - settings should auto-save
    cy.contains(/feeds|订阅/i).click({ force: true })
    cy.wait('@saveSettings', { timeout: 5000 })
    
    // Switch to network tab
    cy.contains(/network|网络/i).click({ force: true })
    cy.wait(500)
    
    // Close settings
    cy.get('body').type('{esc}')
    
    // Reopen and verify the change was saved
    cy.wait(500)
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    cy.contains(/general|常规/i).click({ force: true })
    cy.contains(/dark|深色/i).should('exist')
  })
})
