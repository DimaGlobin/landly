package render

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticRenderer_RenderStatic_Success(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewStaticRenderer(tmpDir)

	projectID := uuid.New()
	schemaJSON := `{
		"pages": [
			{
				"path": "/",
				"title": "Home",
				"blocks": [
					{
						"type": "hero",
						"props": {
							"headline": "Welcome",
							"subheadline": "Test subtitle",
							"ctaText": "Start"
						}
					}
				]
			}
		]
	}`

	buildDir, err := renderer.RenderStatic(context.Background(), projectID, schemaJSON)
	require.NoError(t, err)
	assert.Contains(t, buildDir, projectID.String())

	// Verify index.html exists
	indexPath := filepath.Join(buildDir, "index.html")
	assert.FileExists(t, indexPath)

	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Welcome")
	assert.Contains(t, string(content), "Test subtitle")
	assert.Contains(t, string(content), "Start")

	// Verify static assets
	assert.FileExists(t, filepath.Join(buildDir, "styles.css"))
	assert.FileExists(t, filepath.Join(buildDir, "analytics.js"))
}

func TestStaticRenderer_RenderStatic_MultiplePages(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewStaticRenderer(tmpDir)

	projectID := uuid.New()
	schemaJSON := `{
		"pages": [
			{
				"path": "/",
				"title": "Home",
				"blocks": [
					{
						"type": "hero",
						"props": {"headline": "Home"}
					}
				]
			},
			{
				"path": "/about",
				"title": "About",
				"blocks": [
					{
						"type": "cta",
						"props": {"title": "About Us"}
					}
				]
			}
		]
	}`

	buildDir, err := renderer.RenderStatic(context.Background(), projectID, schemaJSON)
	require.NoError(t, err)

	// Verify both pages
	assert.FileExists(t, filepath.Join(buildDir, "index.html"))
	assert.FileExists(t, filepath.Join(buildDir, "about", "index.html"))

	aboutContent, err := os.ReadFile(filepath.Join(buildDir, "about", "index.html"))
	require.NoError(t, err)
	assert.Contains(t, string(aboutContent), "About Us")
}

func TestStaticRenderer_RenderStatic_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewStaticRenderer(tmpDir)

	projectID := uuid.New()
	schemaJSON := `{invalid json`

	_, err := renderer.RenderStatic(context.Background(), projectID, schemaJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse schema")
}

func TestStaticRenderer_RenderStatic_InvalidPagesStructure(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewStaticRenderer(tmpDir)

	projectID := uuid.New()
	schemaJSON := `{"pages": "not an array"}`

	_, err := renderer.RenderStatic(context.Background(), projectID, schemaJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pages structure")
}

func TestStaticRenderer_RenderBlock_Hero(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	props := map[string]interface{}{
		"headline":    "Test Headline",
		"subheadline": "Test Subheadline",
		"ctaText":     "Click Me",
	}

	html := renderer.renderBlock("hero", props, nil)
	assert.Contains(t, html, "Test Headline")
	assert.Contains(t, html, "Test Subheadline")
	assert.Contains(t, html, "Click Me")
	assert.Contains(t, html, "hero")
}

func TestStaticRenderer_RenderBlock_Features(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	props := map[string]interface{}{
		"title": "Amazing Features",
	}

	html := renderer.renderBlock("features", props, nil)
	assert.Contains(t, html, "Amazing Features")
	assert.Contains(t, html, "features")
}

func TestStaticRenderer_RenderBlock_Pricing(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	props := map[string]interface{}{
		"title": "Our Pricing",
		"plans": []interface{}{
			map[string]interface{}{
				"name":     "Basic",
				"price":    "49",
				"currency": "₽",
				"features": []interface{}{"Feature One", "Feature Two"},
			},
		},
	}

	schema := map[string]interface{}{
		"payment": map[string]interface{}{
			"url": "https://payment.example.com",
		},
	}

	html := renderer.renderBlock("pricing", props, schema)
	assert.Contains(t, html, "Our Pricing")
	assert.Contains(t, html, "Basic")
	assert.Contains(t, html, "Feature One")
	assert.Contains(t, html, "landing-button landing-button--secondary")
	assert.Contains(t, html, "https://payment.example.com")
}

func TestStaticRenderer_RenderBlock_CTA(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	props := map[string]interface{}{
		"title":       "Ready?",
		"buttonText":  "Start Now",
		"description": "Join today",
	}

	html := renderer.renderBlock("cta", props, nil)
	assert.Contains(t, html, "Ready?")
	assert.Contains(t, html, "Start Now")
	assert.Contains(t, html, "Join today")
}

func TestStaticRenderer_RenderBlock_Testimonials(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	props := map[string]interface{}{
		"title": "Customer Reviews",
		"items": []interface{}{
			map[string]interface{}{
				"text":   "Great product!",
				"author": "John Doe",
				"role":   "CEO",
				"rating": "5",
			},
		},
	}

	html := renderer.renderBlock("testimonials", props, nil)
	assert.Contains(t, html, "Customer Reviews")
	assert.Contains(t, html, "Great product!")
	assert.Contains(t, html, "John Doe")
	assert.Contains(t, html, "CEO")
	assert.Contains(t, html, "5")
}

func TestStaticRenderer_RenderBlock_FAQ(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	props := map[string]interface{}{
		"title": "FAQ Section",
		"items": []interface{}{
			map[string]interface{}{
				"question": "What is this?",
				"answer":   "This is a test",
			},
		},
	}

	html := renderer.renderBlock("faq", props, nil)
	assert.Contains(t, html, "FAQ Section")
	assert.Contains(t, html, "What is this?")
	assert.Contains(t, html, "This is a test")
}

func TestStaticRenderer_RenderBlock_Unknown(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	html := renderer.renderBlock("unknown", map[string]interface{}{}, nil)
	assert.Contains(t, html, "Блок unknown пока не поддерживается")
}

func TestGetStringProp(t *testing.T) {
	props := map[string]interface{}{
		"exists":    "value",
		"notString": 123,
	}

	assert.Equal(t, "value", getStringProp(props, "exists", "default"))
	assert.Equal(t, "default", getStringProp(props, "missing", "default"))
	assert.Equal(t, "default", getStringProp(props, "notString", "default"))
}

func TestStaticRenderer_GenerateHTML_WithBlocks(t *testing.T) {
	renderer := NewStaticRenderer("/tmp")

	blocks := []interface{}{
		map[string]interface{}{
			"type": "hero",
			"props": map[string]interface{}{
				"headline": "Test",
			},
		},
	}

	html := renderer.generateHTML("Test Title", blocks, nil)
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<title>Test Title</title>")
	assert.Contains(t, html, "landing-section--hero")
	assert.Contains(t, html, "Test")
}
