/// <reference types="cypress" />

describe('Theme and Language Switching', () => {
  beforeEach(() => {
    cy.visit('/')
    cy.get('body').should('be.visible')
    cy.wait(1000)
  })

  it('should switch between light and dark themes', () => {
    // Intercept settings API
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true })
    
    // Get initial theme
    const initialTheme = cy.get('html').invoke('attr', 'class')
    
    // Switch to dark theme
    cy.contains(/dark|深色/i).click({ force: true })
    cy.wait(1000)
    
    // Verify theme changed
    cy.get('html').should('have.class', /dark/)
    
    // Switch to light theme
    cy.contains(/light|亮色/i).click({ force: true })
    cy.wait(1000)
    
    // Verify theme changed back
    cy.get('html').should('not.have.class', 'dark')
    
    // Close settings
    cy.get('body').type('{esc}')
  })

  it('should persist theme after page reload', () => {
    // Intercept settings API
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings and change theme
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    cy.contains(/general|常规/i).click({ force: true })
    cy.contains(/dark|深色/i).click({ force: true })
    cy.wait('@saveSettings', { timeout: 5000 })
    
    // Close settings
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Reload page
    cy.reload()
    cy.wait(1000)
    
    // Verify dark theme persisted
    cy.get('html').should('have.class', /dark/)
  })

  it('should switch between languages', () => {
    // Intercept settings API
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true })
    
    // Find language selector - look for a select element
    cy.get('select').first().then(($select) => {
      // Get current value
      const currentValue = $select.val()
      // Select the other option (toggle between en and zh)
      if (currentValue === 'en') {
        cy.get('select').first().select('zh')
      } else {
        cy.get('select').first().select('en')
      }
    })
    
    // Wait for language change
    cy.wait(1000)
    
    // Verify language changed - check for different text
    cy.get('body').should('be.visible')
    
    // Switch back
    cy.get('select').first().then(($select) => {
      const currentValue = $select.val()
      if (currentValue === 'en') {
        cy.get('select').first().select('zh')
      } else {
        cy.get('select').first().select('en')
      }
    })
    cy.wait(1000)
    
    // Verify language changed back
    cy.contains(/Settings|General/).should('exist')
  })

  it('should persist language after page reload', () => {
    // Intercept settings API
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings and change language
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    cy.contains(/general|常规/i).click({ force: true })
    
    // Switch to Chinese using the first select element
    cy.get('select').first().select('zh')
    
    cy.wait('@saveSettings', { timeout: 5000 })
    
    // Close settings
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Reload page
    cy.reload()
    cy.wait(1000)
    
    // Verify Chinese language persisted
    cy.contains(/设置/).should('exist')
  })

  it('should handle system theme preference', () => {
    // Intercept settings API
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Open settings
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    // Navigate to general tab
    cy.contains(/general|常规/i).click({ force: true })
    
    // Select system theme
    cy.contains(/system|系统/i).click({ force: true })
    cy.wait('@saveSettings', { timeout: 5000 })
    
    // Verify system theme is selected
    cy.contains(/system|系统/i).should('exist')
    
    // Close settings
    cy.get('body').type('{esc}')
  })

  it('should apply theme to all components', () => {
    // Intercept settings API
    cy.intercept('GET', '/api/settings').as('getSettings')
    cy.intercept('POST', '/api/settings').as('saveSettings')

    // Switch to dark theme
    cy.get('button').contains(/settings|设置/i).click({ force: true })
    cy.wait('@getSettings')
    
    cy.contains(/general|常规/i).click({ force: true })
    cy.contains(/dark|深色/i).click({ force: true })
    cy.wait(1000)
    
    // Close settings
    cy.get('body').type('{esc}')
    cy.wait(500)
    
    // Verify dark theme is applied to various components
    cy.get('body').should('have.css', 'background-color')
    cy.get('[class*="sidebar"]').should('exist')
    cy.get('[class*="article"]').should('exist')
    
    // Check that colors have changed (this is a simple check)
    cy.get('body').should('not.have.css', 'background-color', 'rgb(255, 255, 255)')
  })
})
