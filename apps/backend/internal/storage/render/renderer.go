package render

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// StaticRenderer рендерер статических HTML-сайтов
// PLUGGABLE: можно заменить на более сложную реализацию с SSG-фреймворком
type StaticRenderer struct {
	tmpDir string
}

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
	// Простой HTML-шаблон
	tmpl := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/styles.css">
    <script src="/analytics.js" defer></script>
</head>
<body>
    {{range .Blocks}}
    <section class="block block-{{.Type}}" data-block="{{.Type}}">
        {{.HTML}}
    </section>
    {{end}}
</body>
</html>`

	type Block struct {
		Type string
		HTML template.HTML
	}

	data := struct {
		Title  string
		Blocks []Block
	}{
		Title:  title,
		Blocks: make([]Block, 0),
	}

	// Рендерим каждый блок
	for _, b := range blocks {
		block := b.(map[string]interface{})
		blockType := block["type"].(string)
		props := block["props"].(map[string]interface{})

		blockHTML := r.renderBlock(blockType, props, schema)
		data.Blocks = append(data.Blocks, Block{
			Type: blockType,
			HTML: template.HTML(blockHTML),
		})
	}

	t := template.Must(template.New("page").Parse(tmpl))
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return ""
	}

	return buf.String()
}

func (r *StaticRenderer) renderBlock(blockType string, props map[string]interface{}, schema map[string]interface{}) string {
	// Упрощённый рендеринг блоков
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
		return fmt.Sprintf("<div>Block type: %s</div>", blockType)
	}
}

func (r *StaticRenderer) renderHero(props map[string]interface{}) string {
	headline := getStringProp(props, "headline", "Welcome")
	subheadline := getStringProp(props, "subheadline", "")
	ctaText := getStringProp(props, "ctaText", "Get Started")

	return fmt.Sprintf(`
		<div class="hero">
			<h1>%s</h1>
			<p>%s</p>
			<button class="cta-button" data-track="cta_click">%s</button>
		</div>
	`, headline, subheadline, ctaText)
}

func (r *StaticRenderer) renderFeatures(props map[string]interface{}) string {
	title := getStringProp(props, "title", "Features")
	// Упрощённая реализация
	return fmt.Sprintf(`<div class="features"><h2>%s</h2></div>`, title)
}

func (r *StaticRenderer) renderPricing(props map[string]interface{}, schema map[string]interface{}) string {
	title := getStringProp(props, "title", "Pricing")
	paymentURL := ""

	if payment, ok := schema["payment"].(map[string]interface{}); ok {
		paymentURL = getStringProp(payment, "url", "")
	}

	html := fmt.Sprintf(`<div class="pricing"><h2>%s</h2>`, title)

	if paymentURL != "" {
		html += fmt.Sprintf(`<a href="%s" class="pay-button" data-track="pay_click">Оплатить</a>`, paymentURL)
	}

	html += `</div>`
	return html
}

func (r *StaticRenderer) renderCTA(props map[string]interface{}) string {
	title := getStringProp(props, "title", "Ready to start?")
	buttonText := getStringProp(props, "buttonText", "Get Started")
	description := getStringProp(props, "description", "")

	html := fmt.Sprintf(`<div class="cta"><h2>%s</h2>`, title)
	if description != "" {
		html += fmt.Sprintf(`<p>%s</p>`, description)
	}
	html += fmt.Sprintf(`<button class="cta-button" data-track="cta_click">%s</button></div>`, buttonText)

	return html
}

func (r *StaticRenderer) renderTestimonials(props map[string]interface{}) string {
	title := getStringProp(props, "title", "Отзывы клиентов")

	html := fmt.Sprintf(`<div class="testimonials"><h2>%s</h2>`, title)

	if items, ok := props["items"].([]interface{}); ok {
		html += `<div class="testimonials-grid">`
		for _, item := range items {
			if testimonial, ok := item.(map[string]interface{}); ok {
				text := getStringProp(testimonial, "text", "")
				author := getStringProp(testimonial, "author", "")
				role := getStringProp(testimonial, "role", "")
				rating := getStringProp(testimonial, "rating", "5")

				html += fmt.Sprintf(`
					<div class="testimonial">
						<p>"%s"</p>
						<div class="author">
							<strong>%s</strong>
							<span>%s</span>
							<div class="rating">⭐ %s</div>
						</div>
					</div>
				`, text, author, role, rating)
			}
		}
		html += `</div>`
	}

	html += `</div>`
	return html
}

func (r *StaticRenderer) renderFAQ(props map[string]interface{}) string {
	title := getStringProp(props, "title", "Частые вопросы")

	html := fmt.Sprintf(`<div class="faq"><h2>%s</h2>`, title)

	if items, ok := props["items"].([]interface{}); ok {
		html += `<div class="faq-list">`
		for _, item := range items {
			if faq, ok := item.(map[string]interface{}); ok {
				question := getStringProp(faq, "question", "")
				answer := getStringProp(faq, "answer", "")

				html += fmt.Sprintf(`
					<div class="faq-item">
						<h3>%s</h3>
						<p>%s</p>
					</div>
				`, question, answer)
			}
		}
		html += `</div>`
	}

	html += `</div>`
	return html
}

func (r *StaticRenderer) copyStaticAssets(buildDir string) error {
	// Создаём красивый CSS
	css := `
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    color: #1a1a1a;
    background: #ffffff;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
}

.block {
    margin: 0;
}

/* Hero Section */
.hero {
    text-align: center;
    padding: 120px 0;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    position: relative;
    overflow: hidden;
}

.hero::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><pattern id="grain" width="100" height="100" patternUnits="userSpaceOnUse"><circle cx="25" cy="25" r="1" fill="white" opacity="0.1"/><circle cx="75" cy="75" r="1" fill="white" opacity="0.1"/><circle cx="50" cy="10" r="0.5" fill="white" opacity="0.1"/><circle cx="10" cy="60" r="0.5" fill="white" opacity="0.1"/><circle cx="90" cy="40" r="0.5" fill="white" opacity="0.1"/></pattern></defs><rect width="100" height="100" fill="url(%23grain)"/></svg>');
    opacity: 0.3;
}

.hero h1 {
    font-size: 4rem;
    margin-bottom: 24px;
    font-weight: 800;
    letter-spacing: -0.02em;
    position: relative;
    z-index: 1;
}

.hero p {
    font-size: 1.4rem;
    margin-bottom: 40px;
    opacity: 0.95;
    font-weight: 400;
    position: relative;
    z-index: 1;
}

.cta-button, .pay-button {
    background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
    color: white;
    border: none;
    padding: 18px 36px;
    font-size: 1.1rem;
    font-weight: 600;
    border-radius: 50px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 8px 25px rgba(255, 107, 107, 0.3);
    position: relative;
    z-index: 1;
    text-decoration: none;
    display: inline-block;
}

.cta-button:hover, .pay-button:hover {
    background: linear-gradient(135deg, #ff5252 0%, #d63031 100%);
    transform: translateY(-3px);
    box-shadow: 0 15px 35px rgba(255, 107, 107, 0.4);
}

/* Features Section */
.features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: 40px;
    padding: 100px 0;
    background: #fafbfc;
}

.feature {
    text-align: center;
    padding: 40px 30px;
    border-radius: 20px;
    background: white;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 20px rgba(0,0,0,0.08);
    border: 1px solid rgba(0,0,0,0.05);
}

.feature:hover {
    transform: translateY(-8px);
    box-shadow: 0 20px 40px rgba(0,0,0,0.12);
}

.feature h3 {
    color: #667eea;
    margin-bottom: 20px;
    font-size: 1.4rem;
    font-weight: 700;
}

.feature p {
    color: #666;
    font-size: 1rem;
    line-height: 1.7;
}

/* Pricing Section */
.pricing {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 30px;
    padding: 100px 0;
    background: white;
}

.pricing-card {
    background: white;
    border: 2px solid #e9ecef;
    border-radius: 20px;
    padding: 50px 40px;
    text-align: center;
    position: relative;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 0 4px 20px rgba(0,0,0,0.08);
}

.pricing-card:hover {
    border-color: #667eea;
    transform: translateY(-8px);
    box-shadow: 0 25px 50px rgba(102, 126, 234, 0.15);
}

.pricing-card.featured {
    border-color: #667eea;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    transform: scale(1.05);
}

.pricing-card.featured:hover {
    transform: scale(1.05) translateY(-8px);
}

.price {
    font-size: 3.5rem;
    font-weight: 800;
    color: #667eea;
    margin: 20px 0;
    letter-spacing: -0.02em;
}

.pricing-card.featured .price {
    color: white;
}

.pricing-card h3 {
    font-size: 1.5rem;
    margin-bottom: 20px;
    font-weight: 700;
}

.pricing-card ul {
    list-style: none;
    padding: 0;
    margin: 30px 0;
}

.pricing-card li {
    padding: 12px 0;
    border-bottom: 1px solid #e9ecef;
    font-size: 1rem;
}

.pricing-card.featured li {
    border-bottom-color: rgba(255,255,255,0.2);
}

.pricing-card .cta-button {
    width: 100%;
    margin-top: 20px;
}

/* CTA Section */
.cta {
    text-align: center;
    padding: 120px 0;
    background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
}

.cta h2 {
    font-size: 3rem;
    margin-bottom: 24px;
    color: #1a1a1a;
    font-weight: 800;
    letter-spacing: -0.02em;
}

.cta p {
    font-size: 1.3rem;
    margin-bottom: 40px;
    color: #666;
    font-weight: 400;
}

/* Testimonials Section */
.testimonials {
    padding: 100px 0;
    background: #fafbfc;
}

.testimonials h2 {
    text-align: center;
    font-size: 2.5rem;
    margin-bottom: 60px;
    color: #1a1a1a;
    font-weight: 800;
}

.testimonials-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
    gap: 30px;
}

.testimonial {
    background: white;
    padding: 40px;
    border-radius: 20px;
    box-shadow: 0 8px 30px rgba(0,0,0,0.1);
    transition: all 0.3s ease;
    border: 1px solid rgba(0,0,0,0.05);
}

.testimonial:hover {
    transform: translateY(-5px);
    box-shadow: 0 15px 40px rgba(0,0,0,0.15);
}

.testimonial p {
    font-size: 1.1rem;
    line-height: 1.7;
    color: #333;
    font-style: italic;
    margin-bottom: 20px;
}

.author {
    font-weight: 700;
    color: #667eea;
    margin-top: 20px;
    font-size: 1rem;
}

.author strong {
    display: block;
    font-weight: bold;
}

.author span {
    color: #666;
    font-size: 0.9em;
}

.rating {
    color: #ffc107;
    margin-top: 10px;
    font-size: 1.2rem;
}

/* FAQ Section */
.faq {
    padding: 100px 0;
    background: white;
}

.faq h2 {
    text-align: center;
    font-size: 2.5rem;
    margin-bottom: 60px;
    color: #1a1a1a;
    font-weight: 800;
}

.faq-list {
    max-width: 800px;
    margin: 0 auto;
}

.faq-item {
    background: white;
    margin-bottom: 20px;
    border-radius: 15px;
    overflow: hidden;
    box-shadow: 0 4px 20px rgba(0,0,0,0.08);
    border: 1px solid rgba(0,0,0,0.05);
    transition: all 0.3s ease;
}

.faq-item:hover {
    box-shadow: 0 8px 30px rgba(0,0,0,0.12);
}

.faq-item h3 {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 25px 30px;
    font-weight: 700;
    cursor: pointer;
    font-size: 1.1rem;
    transition: all 0.3s ease;
    margin-bottom: 0;
}

.faq-item h3:hover {
    background: linear-gradient(135deg, #5a6fd8 0%, #6a4190 100%);
}

.faq-item p {
    padding: 30px;
    background: white;
    font-size: 1rem;
    line-height: 1.7;
    color: #555;
    margin-bottom: 0;
}

/* Responsive Design */
@media (max-width: 768px) {
    .hero h1 {
        font-size: 2.5rem;
    }
    
    .hero p {
        font-size: 1.2rem;
    }
    
    .features {
        grid-template-columns: 1fr;
        padding: 60px 0;
    }
    
    .pricing {
        grid-template-columns: 1fr;
        padding: 60px 0;
    }
    
    .pricing-card.featured {
        transform: none;
    }
    
    .cta h2 {
        font-size: 2rem;
    }
    
    .testimonials-grid {
        grid-template-columns: 1fr;
    }
}
`

	// Создаём базовый JS для аналитики
	js := `
// Analytics tracking
const projectId = window.location.hostname.split('.')[0];

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

	if err := os.WriteFile(filepath.Join(buildDir, "styles.css"), []byte(css), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(buildDir, "analytics.js"), []byte(js), 0644); err != nil {
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
