// ***********************************************
// This example commands.ts shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************

/// <reference types="cypress" />

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Custom command to open settings modal
       * @example cy.openSettings()
       */
      openSettings(): Chainable<void>

      /**
       * Custom command to close modal
       * @example cy.closeModal()
       */
      closeModal(): Chainable<void>

      /**
       * Custom command to wait for API response
       * @example cy.waitForApi('/api/settings')
       */
      waitForApi(endpoint: string, alias?: string): Chainable<void>

      /**
       * Custom command to mock API response
       * @example cy.mockApi('/api/settings', { theme: 'dark' })
       */
      mockApi(endpoint: string, response: any): Chainable<void>
    }
  }
}

// Custom command to open settings
Cypress.Commands.add('openSettings', () => {
  // Click the settings button (adjust selector based on your app)
  cy.get('[data-testid="settings-button"]').should('be.visible').click()
  // Wait for modal to be visible
  cy.get('[data-testid="settings-modal"]').should('be.visible')
})

// Custom command to close modal
Cypress.Commands.add('closeModal', () => {
  // Click the close button or press ESC
  cy.get('[data-testid="close-modal"]').click()
  // Wait for modal to disappear
  cy.get('[data-testid="settings-modal"]').should('not.exist')
})

// Custom command to wait for API response
Cypress.Commands.add('waitForApi', (endpoint: string, alias?: string) => {
  const aliasName = alias || endpoint.replace(/\//g, '-')
  cy.intercept('GET', endpoint).as(aliasName)
  cy.wait(`@${aliasName}`)
})

// Custom command to mock API response
Cypress.Commands.add('mockApi', (endpoint: string, response: any) => {
  cy.intercept('GET', endpoint, {
    statusCode: 200,
    body: response,
  }).as(endpoint.replace(/\//g, '-'))
})

export {}
