/// <reference types="cypress" />

describe('Article Operations', () => {
  beforeEach(() => {
    cy.visit('/')
    cy.get('body').should('be.visible')
    cy.wait(1000)
  })

  it('should mark article as read', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')
    cy.intercept('PUT', '/api/articles/*').as('updateArticle')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Click on an article to mark as read
    cy.get('[class*="article"]').first().click({ force: true })
    
    // Wait for update
    cy.wait('@updateArticle', { timeout: 5000 }).then((interception) => {
      // Verify the request updated is_read status
      expect(interception.request.body).to.have.property('is_read')
    })
  })

  it('should mark article as favorite', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')
    cy.intercept('PUT', '/api/articles/*').as('updateArticle')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Right-click on an article to open context menu
    cy.get('[class*="article"]').first().rightclick({ force: true })
    
    // Click favorite option
    cy.contains(/favorite|收藏|star/i).click({ force: true })
    
    // Wait for update
    cy.wait('@updateArticle', { timeout: 5000 })
  })

  it('should filter articles by read status', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Look for filter buttons
    cy.contains(/unread|未读/i).click({ force: true })
    
    // Wait for filtered results
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Verify articles are displayed
    cy.get('[class*="article"]').should('exist')
  })

  it('should filter articles by favorites', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Click favorites filter
    cy.contains(/favorite|收藏/i).click({ force: true })
    
    // Wait for filtered results
    cy.wait('@getArticles', { timeout: 10000 })
  })

  it('should mark all articles as read', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')
    cy.intercept('PUT', '/api/articles/mark-all-read').as('markAllRead')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Look for mark all as read button - could be in a menu
    cy.get('button').contains(/mark.*all|全部标记/i).click({ force: true })
    
    // Wait for confirmation if needed
    cy.contains(/confirm|确认/i).click({ force: true })
    
    // Wait for update
    cy.wait('@markAllRead', { timeout: 10000 })
  })

  it('should open article detail view', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Click on an article
    cy.get('[class*="article"]').first().click({ force: true })
    
    // Verify detail view is shown
    cy.wait(500)
    cy.get('[class*="detail"], [class*="content"]').should('be.visible')
  })

  it('should search articles', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Find search input
    cy.get('input[type="search"], input[placeholder*="search"], input[placeholder*="搜索"]')
      .last()
      .type('test{enter}')
    
    // Wait for search results
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Verify search results
    cy.get('[class*="article"]').should('exist')
  })

  it('should open article in external browser', () => {
    // Intercept API calls
    cy.intercept('GET', '/api/articles*').as('getArticles')

    // Wait for articles to load
    cy.wait('@getArticles', { timeout: 10000 })
    
    // Right-click on article
    cy.get('[class*="article"]').first().rightclick({ force: true })
    
    // Look for "Open in browser" option
    cy.contains(/open.*browser|在浏览器中打开/i).should('exist')
  })
})
