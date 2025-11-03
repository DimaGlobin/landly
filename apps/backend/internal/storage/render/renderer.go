package render

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/google/uuid"
)

// StaticRenderer рендерер статических HTML-сайтов
// PLUGGABLE: можно заменить на более сложную реализацию с SSG-фреймворком
type StaticRenderer struct {
	tmpDir string
}

//go:embed assets/landing.css
var landingCSS string

const analyticsJS = `// Analytics tracking
const projectId = window.location.pathname.split('/')[2];

function track(eventType) {
    fetch('/api/track', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            project_id: projectId,
            event_type: eventType,
            path: window.location.pathname,
            referrer: document.referrer
        })
    });
}

// Track pageview
track('pageview');

// Track button clicks
document.addEventListener('click', function(e) {
    const trackType = e.target.dataset.track;
    if (trackType) {
        track(trackType);
    }
});
`

// NewStaticRenderer создаёт новый статический рендерер
func NewStaticRenderer(tmpDir string) *StaticRenderer {
	return &StaticRenderer{
		tmpDir: tmpDir,
	}
}

// RenderStatic рендерит статический сайт из JSON-схемы
func (r *StaticRenderer) RenderStatic(ctx context.Context, projectID uuid.UUID, schemaJSON string) (string, error) {
	// Парсим схему
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return "", fmt.Errorf("failed to parse schema: %w", err)
	}

	// Создаём временную директорию для проекта
	buildDir := filepath.Join(r.tmpDir, projectID.String())
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}

	// Рендерим страницы
	pages, ok := schema["pages"].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid pages structure in schema")
	}

	for _, p := range pages {
		page, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		if err := r.renderPage(buildDir, page, schema); err != nil {
			return "", fmt.Errorf("failed to render page: %w", err)
		}
	}

	// Копируем статические ресурсы (CSS, JS)
	if err := r.copyStaticAssets(buildDir); err != nil {
		return "", fmt.Errorf("failed to copy static assets: %w", err)
	}

	return buildDir, nil
}

func (r *StaticRenderer) renderPage(buildDir string, page map[string]interface{}, schema map[string]interface{}) error {
	path := page["path"].(string)
	title := page["title"].(string)
	blocks := page["blocks"].([]interface{})

	// Генерируем HTML
	html := r.generateHTML(title, blocks, schema)

	// Определяем путь к файлу
	var filename string
	if path == "/" {
		filename = filepath.Join(buildDir, "index.html")
	} else {
		dir := filepath.Join(buildDir, path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create page directory: %w", err)
		}
		filename = filepath.Join(dir, "index.html")
	}

	// Записываем файл
	return os.WriteFile(filename, []byte(html), 0644)
}

func (r *StaticRenderer) generateHTML(title string, blocks []interface{}, schema map[string]interface{}) string {
	tmpl := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>{{.InlineCSS}}</style>
    <link rel="stylesheet" href="styles.css">
    <script src="analytics.js" defer></script>
</head>
<body class="landing-body">
    <main class="landing" style="{{.ThemeStyle}}">
        {{range .Sections}}
            {{.}}
        {{end}}
    </main>
</body>
</html>`

	palette := extractPalette(schema)
	themeStyle := buildThemeStyle(palette)

	sections := make([]template.HTML, 0, len(blocks))
	for _, rawBlock := range blocks {
		block, ok := rawBlock.(map[string]interface{})
		if !ok {
			continue
		}

		blockType, _ := block["type"].(string)
		props, _ := block["props"].(map[string]interface{})
		sectionHTML := r.renderBlock(blockType, props, schema)
		if sectionHTML == "" {
			continue
		}
		sections = append(sections, template.HTML(sectionHTML))
	}

	if len(sections) == 0 {
		sections = append(sections, template.HTML(`<section class="landing-section"><div class="landing-container"><div class="landing-empty-state">Контент появится после первой генерации</div></div></section>`))
	}

	data := struct {
		Title      string
		ThemeStyle string
		InlineCSS  template.CSS
		Sections   []template.HTML
	}{
		Title:      title,
		ThemeStyle: themeStyle,
		InlineCSS:  template.CSS(landingCSS),
		Sections:   sections,
	}

	t := template.Must(template.New("page").Parse(tmpl))
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return ""
	}

	return buf.String()
}

func (r *StaticRenderer) renderBlock(blockType string, props map[string]interface{}, schema map[string]interface{}) string {
	switch blockType {
	case "hero":
		return r.renderHero(props)
	case "features":
		return r.renderFeatures(props)
	case "pricing":
		return r.renderPricing(props, schema)
	case "cta":
		return r.renderCTA(props)
	case "testimonials":
		return r.renderTestimonials(props)
	case "faq":
		return r.renderFAQ(props)
	default:
		return fmt.Sprintf(`<section class="landing-section" data-block="%s"><div class="landing-container"><div class="landing-empty-state">Блок %s пока не поддерживается</div></div></section>`, html.EscapeString(blockType), html.EscapeString(blockType))
	}
}

func (r *StaticRenderer) renderHero(props map[string]interface{}) string {
	headline := html.EscapeString(getStringProp(props, "headline", "Заголовок лендинга"))
	subheadline := html.EscapeString(getStringProp(props, "subheadline", ""))
	ctaText := html.EscapeString(getStringProp(props, "ctaText", ""))
	ctaURL := html.EscapeString(getStringProp(props, "ctaUrl", "#"))
	secondaryText := html.EscapeString(getStringProp(props, "secondaryCtaText", "Подробнее"))
	secondaryURL := html.EscapeString(getStringProp(props, "secondaryCtaUrl", "#"))
	eyebrow := html.EscapeString(getStringProp(props, "eyebrow", "Инновационная платформа"))
	brand := html.EscapeString(getStringProp(props, "brand", "Landly"))
	navActionText := html.EscapeString(getStringProp(props, "navActionText", "Войти"))
	navActionURL := html.EscapeString(getStringProp(props, "navActionUrl", "#"))
	heroImage := html.EscapeString(getStringProp(props, "image", ""))
	imageAlt := html.EscapeString(getStringProp(props, "imageAlt", headline))

	navItems := toStringSlice(props["navItems"])
	if len(navItems) == 0 {
		navItems = []string{"Возможности", "Цены", "Отзывы", "Контакты"}
	}

	var sb strings.Builder
	sb.WriteString(`<section class="landing-section landing-section--hero" data-block="hero"><div class="landing-hero-overlay"></div><div class="landing-container">`)

	sb.WriteString(`<div class="landing-topbar"><span class="landing-brand">`)
	sb.WriteString(brand)
	sb.WriteString(`</span><nav class="landing-nav">`)
	for _, item := range navItems {
		sb.WriteString(`<a href="#">`)
		sb.WriteString(html.EscapeString(item))
		sb.WriteString(`</a>`)
	}
	sb.WriteString(`</nav>`)
	if navActionText != "" {
		sb.WriteString(`<a class="landing-nav-action" href="`)
		sb.WriteString(navActionURL)
		sb.WriteString(`" target="_blank" rel="noopener noreferrer">`)
		sb.WriteString(navActionText)
		sb.WriteString(`</a>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`<div class="landing-hero-grid"><div class="landing-hero-content">`)
	if eyebrow != "" {
		sb.WriteString(`<span class="landing-eyebrow">`)
		sb.WriteString(eyebrow)
		sb.WriteString(`</span>`)
	}
	sb.WriteString(`<h1>`)
	sb.WriteString(headline)
	sb.WriteString(`</h1>`)
	if subheadline != "" {
		sb.WriteString(`<p>`)
		sb.WriteString(subheadline)
		sb.WriteString(`</p>`)
	}
	if ctaText != "" || secondaryText != "" {
		sb.WriteString(`<div class="landing-actions landing-actions--hero">`)
		if ctaText != "" {
			sb.WriteString(`<a class="landing-button landing-button--primary" data-track="cta_click" href="`)
			sb.WriteString(ctaURL)
			sb.WriteString(`">`)
			sb.WriteString(ctaText)
			sb.WriteString(`</a>`)
		}
		if secondaryText != "" {
			sb.WriteString(`<a class="landing-button landing-button--ghost" data-track="cta_secondary" href="`)
			sb.WriteString(secondaryURL)
			sb.WriteString(`">`)
			sb.WriteString(secondaryText)
			sb.WriteString(`</a>`)
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)

	if heroImage != "" {
		sb.WriteString(`<div class="landing-hero-media"><div class="landing-hero-media-card"><img src="`)
		sb.WriteString(heroImage)
		sb.WriteString(`" alt="`)
		sb.WriteString(imageAlt)
		sb.WriteString(`" /></div></div>`)
	}

	sb.WriteString(`</div></div></section>`)
	return sb.String()
}

func (r *StaticRenderer) renderFeatures(props map[string]interface{}) string {
	title := html.EscapeString(getStringProp(props, "title", "Наши преимущества"))
	items := toSlice(props["items"])

	var sb strings.Builder
	sb.WriteString(`<section class="landing-section landing-section--features" data-block="features"><div class="landing-container">`)
	sb.WriteString(fmt.Sprintf(`<div class="landing-section-header"><h2 class="landing-section-title">%s</h2></div>`, title))
	sb.WriteString(`<div class="landing-features__grid">`)

	if len(items) == 0 {
		sb.WriteString(`<div class="landing-card landing-feature-card"><p class="landing-empty-state">Добавьте преимущества, чтобы показать их здесь</p></div>`)
	} else {
		for _, item := range items {
			icon := html.EscapeString(getStringProp(item, "icon", ""))
			itemTitle := html.EscapeString(getStringProp(item, "title", ""))
			description := html.EscapeString(getStringProp(item, "description", ""))

			sb.WriteString(`<div class="landing-card landing-feature-card">`)
			if icon != "" {
				sb.WriteString(`<div class="landing-feature-icon"><span>`)
				sb.WriteString(icon)
				sb.WriteString(`</span></div>`)
			}
			sb.WriteString(fmt.Sprintf(`<h3>%s</h3>`, itemTitle))
			if description != "" {
				sb.WriteString(fmt.Sprintf(`<p>%s</p>`, description))
			}
			sb.WriteString(`</div>`)
		}
	}

	sb.WriteString(`</div></div></section>`)
	return sb.String()
}

func (r *StaticRenderer) renderPricing(props map[string]interface{}, schema map[string]interface{}) string {
	title := html.EscapeString(getStringProp(props, "title", "Тарифы"))
	plans := toSlice(props["plans"])
	paymentMap, _ := schema["payment"].(map[string]interface{})
	defaultButtonText := html.EscapeString(getStringProp(paymentMap, "buttonText", "Выбрать тариф"))
	defaultURL := html.EscapeString(getStringProp(paymentMap, "url", ""))

	var sb strings.Builder
	sb.WriteString(`<section class="landing-section landing-section--pricing" data-block="pricing"><div class="landing-container">`)
	sb.WriteString(fmt.Sprintf(`<div class="landing-section-header"><h2 class="landing-section-title">%s</h2></div>`, title))
	sb.WriteString(`<div class="landing-pricing__grid">`)

	if len(plans) == 0 {
		sb.WriteString(`<div class="landing-card"><div class="landing-empty-state">Добавьте тарифы в описании проекта</div></div>`)
	} else {
		for _, plan := range plans {
			name := html.EscapeString(getStringProp(plan, "name", ""))
			price := html.EscapeString(getStringProp(plan, "price", ""))
			currency := html.EscapeString(getStringProp(plan, "currency", ""))
			period := html.EscapeString(getStringProp(plan, "period", ""))
			features := toStringSlice(plan["features"])
			featured := getBoolProp(plan, "featured")
			buttonText := html.EscapeString(getStringProp(plan, "buttonText", defaultButtonText))
			buttonURL := html.EscapeString(getStringProp(plan, "url", defaultURL))

			classes := "pricing-card"
			if featured {
				classes += " pricing-card--featured"
			}

			sb.WriteString(fmt.Sprintf(`<div class="%s" data-featured="%t">`, classes, featured))
			sb.WriteString(fmt.Sprintf(`<div class="pricing-name">%s</div>`, name))
			sb.WriteString(`<div class="pricing-price">`)
			sb.WriteString(fmt.Sprintf(`<span class="pricing-price__value">%s</span>`, price))
			sb.WriteString(`<span class="pricing-price__period">`)
			sb.WriteString(currency)
			if period != "" {
				sb.WriteString(fmt.Sprintf(` / %s`, period))
			}
			sb.WriteString(`</span></div>`)

			sb.WriteString(`<ul class="pricing-features">`)
			for _, feature := range features {
				sb.WriteString(`<li class="pricing-feature"><span class="pricing-feature-icon">✓</span><span>`)
				sb.WriteString(html.EscapeString(feature))
				sb.WriteString(`</span></li>`)
			}
			sb.WriteString(`</ul>`)

			sb.WriteString(`<div class="pricing-action">`)
			if buttonURL != "" {
				sb.WriteString(fmt.Sprintf(`<a class="landing-button landing-button--secondary" data-track="pay_click" href="%s" target="_blank" rel="noopener">%s</a>`, buttonURL, buttonText))
			} else {
				sb.WriteString(fmt.Sprintf(`<button type="button" class="landing-button landing-button--secondary" data-track="pay_click">%s</button>`, buttonText))
			}
			sb.WriteString(`</div>`)
			sb.WriteString(`</div>`)
		}
	}

	sb.WriteString(`</div></div></section>`)
	return sb.String()
}

func (r *StaticRenderer) renderTestimonials(props map[string]interface{}) string {
	title := html.EscapeString(getStringProp(props, "title", "Отзывы клиентов"))
	items := toSlice(props["items"])

	var sb strings.Builder
	sb.WriteString(`<section class="landing-section landing-section--testimonials" data-block="testimonials"><div class="landing-container">`)
	sb.WriteString(fmt.Sprintf(`<div class="landing-section-header"><h2 class="landing-section-title">%s</h2></div>`, title))
	sb.WriteString(`<div class="landing-testimonials__grid">`)

	if len(items) == 0 {
		sb.WriteString(`<div class="landing-card landing-testimonial-card"><p class="landing-empty-state">Добавьте отзывы, чтобы повысить доверие</p></div>`)
	} else {
		for _, item := range items {
			text := html.EscapeString(getStringProp(item, "text", ""))
			author := html.EscapeString(getStringProp(item, "author", ""))
			role := html.EscapeString(getStringProp(item, "role", ""))
			rating := html.EscapeString(getStringProp(item, "rating", ""))

			sb.WriteString(`<div class="landing-card landing-testimonial-card">`)
			if text != "" {
				sb.WriteString(fmt.Sprintf(`<p class="landing-testimonial-quote">“%s”</p>`, text))
			}
			sb.WriteString(`<div class="landing-testimonial-author">`)
			if author != "" {
				sb.WriteString(fmt.Sprintf(`<strong>%s</strong>`, author))
			}
			if role != "" {
				sb.WriteString(fmt.Sprintf(`<span>%s</span>`, role))
			}
			if rating != "" {
				sb.WriteString(fmt.Sprintf(`<span class="landing-testimonial-rating">⭐ %s</span>`, rating))
			}
			sb.WriteString(`</div></div>`)
		}
	}

	sb.WriteString(`</div></div></section>`)
	return sb.String()
}

func (r *StaticRenderer) renderFAQ(props map[string]interface{}) string {
	title := html.EscapeString(getStringProp(props, "title", "Частые вопросы"))
	items := toSlice(props["items"])

	var sb strings.Builder
	sb.WriteString(`<section class="landing-section landing-section--faq" data-block="faq"><div class="landing-container">`)
	sb.WriteString(fmt.Sprintf(`<div class="landing-section-header"><h2 class="landing-section-title">%s</h2></div>`, title))
	sb.WriteString(`<div class="landing-faq__list">`)

	if len(items) == 0 {
		sb.WriteString(`<div class="faq-item"><div class="landing-empty-state">Добавьте вопросы и ответы, которые волнуют клиентов</div></div>`)
	} else {
		for _, item := range items {
			question := html.EscapeString(getStringProp(item, "question", ""))
			answer := html.EscapeString(getStringProp(item, "answer", ""))
			sb.WriteString(`<div class="faq-item">`)
			sb.WriteString(fmt.Sprintf(`<div class="faq-question">%s</div>`, question))
			if answer != "" {
				sb.WriteString(fmt.Sprintf(`<div class="faq-answer">%s</div>`, answer))
			}
			sb.WriteString(`</div>`)
		}
	}

	sb.WriteString(`</div></div></section>`)
	return sb.String()
}

func (r *StaticRenderer) renderCTA(props map[string]interface{}) string {
	title := html.EscapeString(getStringProp(props, "title", "Готовы начать?"))
	description := html.EscapeString(getStringProp(props, "description", ""))
	buttonText := html.EscapeString(getStringProp(props, "buttonText", "Связаться"))
	buttonURL := html.EscapeString(getStringProp(props, "buttonUrl", "#"))
	secondaryText := html.EscapeString(getStringProp(props, "secondaryButtonText", ""))
	secondaryURL := html.EscapeString(getStringProp(props, "secondaryButtonUrl", "#"))

	var sb strings.Builder
	sb.WriteString(`<section class="landing-section landing-section--cta" data-block="cta"><div class="landing-container">`)
	sb.WriteString(`<div class="landing-section-header">`)
	sb.WriteString(fmt.Sprintf(`<h2 class="landing-section-title">%s</h2>`, title))
	if description != "" {
		sb.WriteString(fmt.Sprintf(`<p>%s</p>`, description))
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<div class="landing-actions landing-actions--center">`)
	sb.WriteString(`<a class="landing-button landing-button--primary" data-track="cta_click" href="`)
	sb.WriteString(buttonURL)
	sb.WriteString(`">`)
	sb.WriteString(buttonText)
	sb.WriteString(`</a>`)
	if secondaryText != "" {
		sb.WriteString(`<a class="landing-button landing-button--ghost" data-track="cta_secondary" href="`)
		sb.WriteString(secondaryURL)
		sb.WriteString(`">`)
		sb.WriteString(secondaryText)
		sb.WriteString(`</a>`)
	}
	sb.WriteString(`</div></div></section>`)
	return sb.String()
}

func (r *StaticRenderer) copyStaticAssets(buildDir string) error {
	if err := os.WriteFile(filepath.Join(buildDir, "styles.css"), []byte(landingCSS), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(buildDir, "analytics.js"), []byte(analyticsJS), 0644); err != nil {
		return err
	}

	return nil
}

func getStringProp(props map[string]interface{}, key, defaultValue string) string {
	if val, ok := props[key].(string); ok {
		return val
	}
	return defaultValue
}

type paletteValues struct {
	Primary    string
	Secondary  string
	Accent     string
	Background string
	Text       string
}

func extractPalette(schema map[string]interface{}) paletteValues {
	palette := paletteValues{
		Primary:    "#2563EB",
		Secondary:  "#7C3AED",
		Accent:     "#F97316",
		Background: "#FFFFFF",
		Text:       "#1F2937",
	}

	if theme, ok := schema["theme"].(map[string]interface{}); ok {
		if rawPalette, ok := theme["palette"].(map[string]interface{}); ok {
			if val := getStringProp(rawPalette, "primary", palette.Primary); val != "" {
				palette.Primary = val
			}
			if val := getStringProp(rawPalette, "secondary", palette.Secondary); val != "" {
				palette.Secondary = val
			}
			if val := getStringProp(rawPalette, "accent", palette.Accent); val != "" {
				palette.Accent = val
			}
			if val := getStringProp(rawPalette, "background", palette.Background); val != "" {
				palette.Background = val
			}
			if val := getStringProp(rawPalette, "text", palette.Text); val != "" {
				palette.Text = val
			}
		}
	}

	return palette
}

func buildThemeStyle(p paletteValues) string {
	return fmt.Sprintf("--landing-primary:%s;--landing-secondary:%s;--landing-accent:%s;--landing-background:%s;--landing-text:%s;",
		html.EscapeString(p.Primary),
		html.EscapeString(p.Secondary),
		html.EscapeString(p.Accent),
		html.EscapeString(p.Background),
		html.EscapeString(p.Text),
	)
}

func toSlice(value interface{}) []map[string]interface{} {
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

func toStringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

func getBoolProp(props map[string]interface{}, key string) bool {
	if val, ok := props[key].(bool); ok {
		return val
	}
	if val, ok := props[key].(string); ok {
		return strings.EqualFold(val, "true")
	}
	return false
}
