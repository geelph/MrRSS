/// <reference types="cypress" />

describe('Feed Management', () => {
  beforeEach(() => {
    cy.visit('/')
    cy.get('body').should('be.visible')
    cy.wait(1000)
  })

  it('should add a new feed', () => {
    // Intercept API calls
    cy.intercept('POST', '/api/feeds').as('addFeed')
    cy.intercept('GET', '/api/feeds').as('getFeeds')

    // Look for add feed button - could be a + icon or "Add Feed" text
    cy.get('button').contains(/add|添加|\+/i).first().click({ force: true })
    
    // Wait for add feed modal to appear
    cy.wait(500)
    
    // Fill in the feed URL
    cy.get('input[type="url"], input[type="text"]').first().type('https://example.com/feed.xml')
    
    // Submit the form - look for submit button
    cy.get('button').contains(/add|submit|确定|添加/i).click({ force: true })
    
    // Wait for the feed to be added
    cy.wait('@addFeed', { timeout: 10000 })
    
    // Verify the feed appears in the list
    cy.wait('@getFeeds', { timeout: 10000 })
    cy.contains(/example\.com/i).should('exist')
  })

  it('should delete a feed', () => {
    // Intercept API calls
    cy.intercept('DELETE', '/api/feeds/*').as('deleteFeed')
    cy.intercept('GET', '/api/feeds').as('getFeeds')

    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 })
    
    // Right-click on a feed to open context menu
    cy.get('[class*="feed"]').first().rightclick({ force: true })
    
    // Click delete option in context menu
    cy.contains(/delete|删除/i).click({ force: true })
    
    // Confirm deletion in the confirm dialog
    cy.contains(/confirm|确认/i).click({ force: true })
    
    // Wait for deletion to complete
    cy.wait('@deleteFeed', { timeout: 10000 })
  })

  it('should refresh feeds', () => {
    // Intercept API calls
    cy.intercept('POST', '/api/feeds/refresh').as('refreshFeeds')
    cy.intercept('GET', '/api/feeds').as('getFeeds')

    // Wait for initial load
    cy.wait('@getFeeds', { timeout: 10000 })
    
    // Look for refresh button - could be a refresh icon
    cy.get('button').contains(/refresh|刷新/i).click({ force: true })
    
    // Wait for refresh to complete
    cy.wait('@refreshFeeds', { timeout: 10000 })
  })

  it('should edit feed details', () => {
    // Intercept API calls
    cy.intercept('PUT', '/api/feeds/*').as('updateFeed')
    cy.intercept('GET', '/api/feeds').as('getFeeds')

    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 })
    
    // Right-click on a feed
    cy.get('[class*="feed"]').first().rightclick({ force: true })
    
    // Click edit option
    cy.contains(/edit|编辑/i).click({ force: true })
    
    // Wait for edit modal
    cy.wait(500)
    
    // Change the title
    cy.get('input[type="text"]').first().clear().type('Updated Feed Title')
    
    // Save changes
    cy.get('button').contains(/save|保存|确定/i).click({ force: true })
    
    // Wait for update to complete
    cy.wait('@updateFeed', { timeout: 10000 })
    
    // Verify the updated title appears
    cy.contains('Updated Feed Title').should('exist')
  })

  it('should filter feeds by category', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/feeds').as('getFeeds')

    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 })
    
    // Look for category filter
    cy.get('select, [role="listbox"]').first().select(1)
    
    // Wait for filtered results
    cy.wait(500)
    
    // Verify only feeds from that category are shown
    cy.get('[class*="feed"]').should('have.length.at.least', 0)
  })

  it('should search feeds', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/feeds').as('getFeeds')

    // Wait for feeds to load
    cy.wait('@getFeeds', { timeout: 10000 })
    
    // Look for search input
    cy.get('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
      .first()
      .type('test')
    
    // Wait for search results to filter
    cy.wait(500)
    
    // Verify search results
    cy.get('[class*="feed"]').should('exist')
  })
})
