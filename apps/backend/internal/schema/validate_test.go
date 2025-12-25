package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Field: "test.field", Message: "something wrong"}
	assert.Equal(t, "validation error in field 'test.field': something wrong", err.Error())
}

func TestValidateSchema_ValidSchema(t *testing.T) {
	validSchema := `{
		"pages": [{
			"path": "/",
			"title": "Home",
			"blocks": [{
				"id": "block-1",
				"type": "hero",
				"order": 1,
				"props": {
					"title": "Welcome",
					"subtitle": "to our site"
				}
			}]
		}]
	}`

	err := ValidateSchema(validSchema)
	assert.NoError(t, err)
}

func TestValidateSchema_InvalidJSON(t *testing.T) {
	err := ValidateSchema("not valid json")
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "schema_json", validationErr.Field)
	assert.Contains(t, validationErr.Message, "invalid JSON")
}

func TestValidateSchema_MissingPages(t *testing.T) {
	err := ValidateSchema(`{}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages", validationErr.Field)
	assert.Contains(t, validationErr.Message, "missing")
}

func TestValidateSchema_PagesNotArray(t *testing.T) {
	err := ValidateSchema(`{"pages": "not an array"}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages", validationErr.Field)
	assert.Contains(t, validationErr.Message, "array")
}

func TestValidateSchema_EmptyPages(t *testing.T) {
	err := ValidateSchema(`{"pages": []}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages", validationErr.Field)
	assert.Contains(t, validationErr.Message, "at least one")
}

func TestValidateSchema_PageNotObject(t *testing.T) {
	err := ValidateSchema(`{"pages": ["not an object"]}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0]", validationErr.Field)
	assert.Contains(t, validationErr.Message, "object")
}

func TestValidateSchema_MissingPath(t *testing.T) {
	err := ValidateSchema(`{"pages": [{"title": "Home", "blocks": []}]}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].path", validationErr.Field)
}

func TestValidateSchema_MissingTitle(t *testing.T) {
	err := ValidateSchema(`{"pages": [{"path": "/", "blocks": []}]}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].title", validationErr.Field)
}

func TestValidateSchema_MissingBlocks(t *testing.T) {
	err := ValidateSchema(`{"pages": [{"path": "/", "title": "Home"}]}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks", validationErr.Field)
}

func TestValidateSchema_BlocksNotArray(t *testing.T) {
	err := ValidateSchema(`{"pages": [{"path": "/", "title": "Home", "blocks": "not array"}]}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks", validationErr.Field)
}

func TestValidateSchema_BlockNotObject(t *testing.T) {
	err := ValidateSchema(`{"pages": [{"path": "/", "title": "Home", "blocks": ["not object"]}]}`)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks[0]", validationErr.Field)
}

func TestValidateSchema_MissingBlockType(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"order": 1, "props": {}}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks[0].type", validationErr.Field)
}

func TestValidateSchema_BlockTypeNotString(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": 123, "order": 1, "props": {}}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks[0].type", validationErr.Field)
	assert.Contains(t, validationErr.Message, "string")
}

func TestValidateSchema_InvalidBlockType(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": "invalid_type", "order": 1, "props": {}}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Contains(t, validationErr.Message, "not allowed")
}

func TestValidateSchema_AllAllowedBlockTypes(t *testing.T) {
	blockTypes := []string{"hero", "features", "pricing", "testimonials", "faq", "cta", "gallery", "about", "contact"}

	for _, blockType := range blockTypes {
		t.Run(blockType, func(t *testing.T) {
			schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": "` + blockType + `", "order": 1, "props": {}}]}]}`
			err := ValidateSchema(schema)
			assert.NoError(t, err, "block type %s should be allowed", blockType)
		})
	}
}

func TestValidateSchema_MissingOrder(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": "hero", "props": {}}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Contains(t, validationErr.Message, "order")
}

func TestValidateSchema_SortInsteadOfOrder(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": "hero", "sort": 1, "props": {}}]}]}`
	err := ValidateSchema(schema)
	assert.NoError(t, err)
}

func TestValidateSchema_MissingProps(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": "hero", "order": 1}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks[0].props", validationErr.Field)
}

func TestValidateSchema_PropsNotObject(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"type": "hero", "order": 1, "props": "not object"}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Contains(t, validationErr.Message, "object")
}

func TestValidateSchema_InvalidBlockId(t *testing.T) {
	schema := `{"pages": [{"path": "/", "title": "Home", "blocks": [{"id": 123, "type": "hero", "order": 1, "props": {}}]}]}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Contains(t, validationErr.Message, "id")
	assert.Contains(t, validationErr.Message, "string")
}

func TestValidateSchema_MultiplePages(t *testing.T) {
	schema := `{
		"pages": [
			{"path": "/", "title": "Home", "blocks": [{"type": "hero", "order": 1, "props": {}}]},
			{"path": "/about", "title": "About", "blocks": [{"type": "about", "order": 1, "props": {}}]}
		]
	}`
	err := ValidateSchema(schema)
	assert.NoError(t, err)
}

func TestValidateSchema_MultipleBlocks(t *testing.T) {
	schema := `{
		"pages": [{
			"path": "/",
			"title": "Home",
			"blocks": [
				{"type": "hero", "order": 1, "props": {"title": "Welcome"}},
				{"type": "features", "order": 2, "props": {}},
				{"type": "cta", "order": 3, "props": {}}
			]
		}]
	}`
	err := ValidateSchema(schema)
	assert.NoError(t, err)
}

func TestValidateSchema_ErrorOnSecondPage(t *testing.T) {
	schema := `{
		"pages": [
			{"path": "/", "title": "Home", "blocks": [{"type": "hero", "order": 1, "props": {}}]},
			{"path": "/about", "blocks": [{"type": "about", "order": 1, "props": {}}]}
		]
	}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[1].title", validationErr.Field)
}

func TestValidateSchema_ErrorOnSecondBlock(t *testing.T) {
	schema := `{
		"pages": [{
			"path": "/",
			"title": "Home",
			"blocks": [
				{"type": "hero", "order": 1, "props": {}},
				{"type": "invalid", "order": 2, "props": {}}
			]
		}]
	}`
	err := ValidateSchema(schema)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok)
	assert.Equal(t, "pages[0].blocks[1].type", validationErr.Field)
}

