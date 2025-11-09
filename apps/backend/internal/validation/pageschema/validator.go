package pageschema

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

const schemaResourceID = "landly://page-schema"

// schemaDocument is embedded copy of docs/page_schema.json (keep in sync).
//
//go:embed schema.json
var schemaDocument []byte

var compiledSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7
	if err := compiler.AddResource(schemaResourceID, bytes.NewReader(schemaDocument)); err != nil {
		panic(fmt.Sprintf("pageschema: failed to load schema resource: %v", err))
	}

	schema, err := compiler.Compile(schemaResourceID)
	if err != nil {
		panic(fmt.Sprintf("pageschema: failed to compile schema: %v", err))
	}

	compiledSchema = schema
}

// Result описывает итоги проверки страницы.
type Result struct {
	Normalized []byte
	AutoFixes  []string
}

// Validate проверяет JSON схему и применяет простые авто-фиксы.
func Validate(raw []byte) (*Result, error) {
	var doc interface{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("schema is not a valid JSON object: %w", err)
	}

	if err := compiledSchema.Validate(doc); err != nil {
		return nil, fmt.Errorf("schema does not match specification: %w", err)
	}

	autoFixes := applyBusinessRules(doc)

	normalized, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to encode schema: %w", err)
	}

	return &Result{
		Normalized: normalized,
		AutoFixes:  uniqueStrings(autoFixes),
	}, nil
}

func applyBusinessRules(doc interface{}) []string {
	root, ok := doc.(map[string]interface{})
	if !ok {
		return nil
	}

	var fixes []string
	paymentMap, _ := root["payment"].(map[string]interface{})
	paymentURL := strings.TrimSpace(getString(paymentMap, "url"))
	if paymentURL == "" {
		paymentURL = "#"
	}
	paymentButton := getString(paymentMap, "buttonText")
	if paymentButton == "" {
		paymentButton = "Связаться"
	}

	pages := toSlice(root["pages"])
	if len(pages) == 0 {
		defaultPage := map[string]interface{}{
			"path":   "/",
			"title":  "Лендинг",
			"blocks": []interface{}{},
		}
		root["pages"] = []interface{}{defaultPage}
		pages = toSlice(root["pages"])
		fixes = append(fixes, "default_page_added")
	}

	for i, raw := range pages {
		page, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if fix := normalizePagePath(page, i); fix != "" {
			fixes = append(fixes, fix)
		}
		if fix := truncateField(page, "title", 90, fmt.Sprintf("page_%d_title_truncated", i)); fix != "" {
			fixes = append(fixes, fix)
		}
		if fix := truncateField(page, "description", 160, fmt.Sprintf("page_%d_description_truncated", i)); fix != "" {
			fixes = append(fixes, fix)
		}

		pageFixes := normalizeBlocks(page, paymentURL, paymentButton)
		fixes = append(fixes, pageFixes...)

		pages[i] = page
	}

	root["pages"] = pages
	return fixes
}

func normalizeBlocks(page map[string]interface{}, paymentURL, paymentButton string) []string {
	blocks := toSlice(page["blocks"])
	if len(blocks) == 0 {
		return nil
	}

	var fixes []string
	for i, raw := range blocks {
		block, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		blockType := strings.ToLower(strings.TrimSpace(getString(block, "type")))
		props, _ := block["props"].(map[string]interface{})
		if props == nil {
			props = map[string]interface{}{}
		}

		switch blockType {
		case "hero":
			fixes = append(fixes, ensureHeroCTA(props, paymentURL, paymentButton)...)
			fixes = append(fixes, truncateField(props, "headline", 90, fmt.Sprintf("block_%d_hero_headline_truncated", i)))
			fixes = append(fixes, truncateField(props, "subheadline", 160, fmt.Sprintf("block_%d_hero_subheadline_truncated", i)))
		case "features":
			fixes = append(fixes, ensureFeatureItems(props)...)
		case "cta":
			fixes = append(fixes, ensureCTABlock(props, paymentURL, paymentButton)...)
		case "pricing":
			fixes = append(fixes, ensurePricingButtons(props, paymentURL, paymentButton)...)
		}

		block["props"] = props
		blocks[i] = block
	}

	page["blocks"] = blocks
	return fixes
}

func ensureHeroCTA(props map[string]interface{}, paymentURL, paymentButton string) []string {
	var fixes []string
	ctaText := strings.TrimSpace(getString(props, "ctaText"))
	if ctaText == "" {
		props["ctaText"] = paymentButton
		fixes = append(fixes, "hero_cta_text_defaulted")
	}

	ctaURL := strings.TrimSpace(getString(props, "ctaUrl"))
	if ctaURL == "" {
		props["ctaUrl"] = paymentURL
		fixes = append(fixes, "hero_cta_url_defaulted")
	}

	return fixes
}

func ensureCTABlock(props map[string]interface{}, paymentURL, paymentButton string) []string {
	var fixes []string

	buttonText := strings.TrimSpace(getString(props, "buttonText"))
	if buttonText == "" {
		props["buttonText"] = paymentButton
		fixes = append(fixes, "cta_button_text_defaulted")
	}

	buttonURL := strings.TrimSpace(getString(props, "buttonUrl"))
	if buttonURL == "" {
		props["buttonUrl"] = paymentURL
		fixes = append(fixes, "cta_button_url_defaulted")
	}

	return fixes
}

func ensurePricingButtons(props map[string]interface{}, paymentURL, paymentButton string) []string {
	plans := toSlice(props["plans"])
	if len(plans) == 0 {
		return nil
	}

	var fixes []string
	for _, raw := range plans {
		plan, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if strings.TrimSpace(getString(plan, "buttonText")) == "" {
			plan["buttonText"] = paymentButton
			fixes = append(fixes, "pricing_button_text_defaulted")
		}
		if strings.TrimSpace(getString(plan, "url")) == "" {
			plan["url"] = paymentURL
			fixes = append(fixes, "pricing_button_url_defaulted")
		}
	}

	props["plans"] = plans
	return fixes
}

func ensureFeatureItems(props map[string]interface{}) []string {
	items := toSlice(props["items"])
	if len(items) >= 3 {
		return nil
	}

	defaults := []map[string]string{
		{"icon": "⚡", "title": "Быстрый запуск", "description": "Лендинг готов за минуты"},
		{"icon": "🎯", "title": "Убедительные тексты", "description": "AI подстраивается под вашу нишу"},
		{"icon": "🚀", "title": "Рост конверсии", "description": "Современный дизайн и CTA"},
	}

	for len(items) < 3 {
		template := defaults[len(items)%len(defaults)]
		items = append(items, map[string]interface{}{
			"icon":        template["icon"],
			"title":       template["title"],
			"description": template["description"],
		})
	}

	props["items"] = items
	return []string{"features_autofilled"}
}

func normalizePagePath(page map[string]interface{}, idx int) string {
	path := strings.TrimSpace(getString(page, "path"))
	if path == "" {
		if idx == 0 {
			path = "/"
		} else {
			path = fmt.Sprintf("/page-%d", idx+1)
		}
		page["path"] = path
		return fmt.Sprintf("page_%d_path_defaulted", idx)
	}
	if !strings.HasPrefix(path, "/") {
		page["path"] = "/" + path
		return fmt.Sprintf("page_%d_path_normalized", idx)
	}
	return ""
}

func truncateField(m map[string]interface{}, key string, limit int, reason string) string {
	value := getString(m, key)
	if value == "" {
		return ""
	}
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 0 {
		return ""
	}

	if utf8.RuneCountInString(trimmed) > limit {
		m[key] = truncateRunes(trimmed, limit)
		return reason
	}

	if trimmed != value {
		m[key] = trimmed
		return reason + "_trimmed"
	}

	return ""
}

func truncateRunes(input string, limit int) string {
	if limit <= 0 {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(input))
	count := 0
	for _, r := range input {
		builder.WriteRune(r)
		count++
		if count >= limit {
			break
		}
	}
	return builder.String()
}

func toSlice(value interface{}) []interface{} {
	if value == nil {
		return nil
	}
	if slice, ok := value.([]interface{}); ok {
		return slice
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key]; ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
