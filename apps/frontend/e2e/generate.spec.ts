import { test, expect } from '@playwright/test'

test.describe('Landing Page Generation', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    const email = `test-${Date.now()}@example.com`
    const password = 'SecurePassword123'

    await page.goto('/auth/signup')
    await page.fill('input[name="email"]', email)
    await page.fill('input[name="password"]', password)
    await page.fill('input[name="confirmPassword"]', password)
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL(/\/app\/projects/)
  })

  test('should generate a simple landing page', async ({ page }) => {
    await page.goto('/app/projects')

    // Click "Create Project" or "Simple Generate"
    await page.click('text=/create.*project|new.*project/i')

    // Fill out simple form
    await page.fill('input[name="businessName"]', 'My SaaS Product')
    await page.fill('textarea[name="description"]', 'Cloud-based collaboration tool')

    // Submit
    await page.click('button:has-text("Generate")')

    // Should see loading state
    await expect(page.locator('text=/generating|processing/i')).toBeVisible({ timeout: 2000 })

    // Wait for generation to complete
    await expect(page.locator('text=/preview|generated/i')).toBeVisible({ timeout: 30000 })

    // Should see preview
    await expect(page.locator('.landing-preview')).toBeVisible()
  })

  test('should allow editing generated content', async ({ page }) => {
    // Generate a page first
    await page.goto('/app/projects')
    await page.click('text=/create.*project/i')
    await page.fill('input[name="businessName"]', 'Test Product')
    await page.fill('textarea[name="description"]', 'Test description')
    await page.click('button:has-text("Generate")')

    await expect(page.locator('text=/preview/i')).toBeVisible({ timeout: 30000 })

    // Switch to edit mode
    await page.click('button:has-text("Edit")')

    // Modify content
    await page.fill('input[name="title"]', 'Updated Title')
    await page.click('button:has-text("Save")')

    // Verify changes
    await expect(page.locator('text=Updated Title')).toBeVisible()
  })

  test('should publish generated landing page', async ({ page }) => {
    // Generate a page
    await page.goto('/app/projects')
    await page.click('text=/create.*project/i')
    await page.fill('input[name="businessName"]', 'Publish Test')
    await page.fill('textarea[name="description"]', 'Testing publish flow')
    await page.click('button:has-text("Generate")')

    await expect(page.locator('text=/preview/i')).toBeVisible({ timeout: 30000 })

    // Click publish
    await page.click('button:has-text("Publish")')

    // Should see success message with URL
    await expect(page.locator('text=/published|live/i')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('text=https://landly.com/')).toBeVisible()

    // Should have a "View Live" button
    const viewLiveButton = page.locator('a:has-text("View Live")')
    await expect(viewLiveButton).toBeVisible()
    await expect(viewLiveButton).toHaveAttribute('href', /https:\/\/landly\.com\//)
  })

  test('should handle generation errors gracefully', async ({ page }) => {
    await page.goto('/app/projects')
    await page.click('text=/create.*project/i')

    // Submit with empty fields
    await page.click('button:has-text("Generate")')

    // Should show validation errors
    await expect(page.locator('text=/required|fill.*field/i')).toBeVisible()
  })

  test('should persist project after page refresh', async ({ page }) => {
    // Generate a project
    await page.goto('/app/projects')
    await page.click('text=/create.*project/i')
    await page.fill('input[name="businessName"]', 'Persistence Test')
    await page.fill('textarea[name="description"]', 'Testing persistence')
    await page.click('button:has-text("Generate")')

    await expect(page.locator('text=/preview/i')).toBeVisible({ timeout: 30000 })

    // Get project URL
    const projectUrl = page.url()

    // Refresh page
    await page.reload()

    // Should still see the preview
    await expect(page.locator('.landing-preview')).toBeVisible()
    await expect(page.locator('text=Persistence Test')).toBeVisible()
  })
})

