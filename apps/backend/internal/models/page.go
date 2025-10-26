package domain

import (
	"time"

	"github.com/google/uuid"
)

// Page представляет страницу лендинга
type Page struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	Path      string    `db:"path" json:"path"`
	Title     string    `db:"title" json:"title"`
	MetaJSON  string    `db:"meta_json" json:"meta_json,omitempty"` // SEO мета-данные
	Sort      int       `db:"sort" json:"sort"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// NewPage создаёт новую страницу
func NewPage(projectID uuid.UUID, path, title string, sort int) *Page {
	return &Page{
		ID:        uuid.New(),
		ProjectID: projectID,
		Path:      path,
		Title:     title,
		Sort:      sort,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
