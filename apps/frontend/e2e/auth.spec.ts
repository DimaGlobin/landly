import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Reset state before each test
    await page.goto('/')
  })

  test('should sign up a new user', async ({ page }) => {
    await page.goto('/auth/signup')

    // Fill out signup form
    await page.fill('input[name="email"]', `test-${Date.now()}@example.com`)
    await page.fill('input[name="password"]', 'SecurePassword123')
    await page.fill('input[name="confirmPassword"]', 'SecurePassword123')

    // Submit form
    await page.click('button[type="submit"]')

    // Should redirect to projects page
    await expect(page).toHaveURL(/\/app\/projects/)
    await expect(page.locator('h1')).toContainText('Projects')
  })

  test('should login with existing credentials', async ({ page }) => {
    // First, create a user
    const email = `test-${Date.now()}@example.com`
    const password = 'SecurePassword123'

    await page.goto('/auth/signup')
    await page.fill('input[name="email"]', email)
    await page.fill('input[name="password"]', password)
    await page.fill('input[name="confirmPassword"]', password)
    await page.click('button[type="submit"]')

    // Logout
    await page.click('button[aria-label="User menu"]')
    await page.click('text=Logout')

    // Login again
    await page.goto('/auth/login')
    await page.fill('input[name="email"]', email)
    await page.fill('input[name="password"]', password)
    await page.click('button[type="submit"]')

    // Should be logged in
    await expect(page).toHaveURL(/\/app\/projects/)
  })

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/auth/login')

    await page.fill('input[name="email"]', 'invalid@example.com')
    await page.fill('input[name="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')

    // Should show error message
    await expect(page.locator('text=/invalid credentials/i')).toBeVisible()
    await expect(page).toHaveURL('/auth/login')
  })

  test('should validate email format', async ({ page }) => {
    await page.goto('/auth/signup')

    await page.fill('input[name="email"]', 'invalid-email')
    await page.fill('input[name="password"]', 'password123')
    await page.blur('input[name="email"]')

    await expect(page.locator('text=/invalid email/i')).toBeVisible()
  })

  test('should require password confirmation to match', async ({ page }) => {
    await page.goto('/auth/signup')

    await page.fill('input[name="email"]', 'test@example.com')
    await page.fill('input[name="password"]', 'password123')
    await page.fill('input[name="confirmPassword"]', 'different123')
    await page.click('button[type="submit"]')

    await expect(page.locator('text=/passwords.*match/i')).toBeVisible()
  })
})

