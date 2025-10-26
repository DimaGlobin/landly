package domain

import (
	"time"

	"github.com/google/uuid"
)

// BlockType тип блока
type BlockType string

const (
	BlockTypeHero         BlockType = "hero"
	BlockTypeFeatures     BlockType = "features"
	BlockTypePricing      BlockType = "pricing"
	BlockTypeTestimonials BlockType = "testimonials"
	BlockTypeFAQ          BlockType = "faq"
	BlockTypeCTA          BlockType = "cta"
	BlockTypeGallery      BlockType = "gallery"
	BlockTypeAbout        BlockType = "about"
	BlockTypeContact      BlockType = "contact"
)

// Block представляет блок (секцию) страницы
type Block struct {
	ID        uuid.UUID `db:"id" json:"id"`
	PageID    uuid.UUID `db:"page_id" json:"page_id"`
	Type      BlockType `db:"type" json:"type"`
	PropsJSON string    `db:"props_json" json:"props_json"` // Гибкие свойства в JSON
	Sort      int       `db:"sort" json:"sort"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// NewBlock создаёт новый блок
func NewBlock(pageID uuid.UUID, blockType BlockType, propsJSON string, sort int) *Block {
	return &Block{
		ID:        uuid.New(),
		PageID:    pageID,
		Type:      blockType,
		PropsJSON: propsJSON,
		Sort:      sort,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
