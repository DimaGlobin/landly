package query

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

// Dialect представляет диалект базы данных
type Dialect string

const (
	PostgreSQL Dialect = "postgres"
	MySQL      Dialect = "mysql"
	SQLite     Dialect = "sqlite"
)

// Builder представляет query builder для конкретного диалекта
type Builder struct {
	dialect Dialect
	db      *sql.DB
}

// NewBuilder создает новый query builder
func NewBuilder(dialect Dialect, db *sql.DB) *Builder {
	return &Builder{
		dialect: dialect,
		db:      db,
	}
}

// GetDialect возвращает диалект
func (b *Builder) GetDialect() Dialect {
	return b.dialect
}

// GetDB возвращает подключение к БД
func (b *Builder) GetDB() *sql.DB {
	return b.db
}

// Select создает SELECT запрос
func (b *Builder) Select(columns ...string) squirrel.SelectBuilder {
	return squirrel.Select(columns...).PlaceholderFormat(b.getPlaceholder())
}

// Insert создает INSERT запрос
func (b *Builder) Insert(table string) squirrel.InsertBuilder {
	return squirrel.Insert(table).PlaceholderFormat(b.getPlaceholder())
}

// Update создает UPDATE запрос
func (b *Builder) Update(table string) squirrel.UpdateBuilder {
	return squirrel.Update(table).PlaceholderFormat(b.getPlaceholder())
}

// Delete создает DELETE запрос
func (b *Builder) Delete(table string) squirrel.DeleteBuilder {
	return squirrel.Delete(table).PlaceholderFormat(b.getPlaceholder())
}

// getPlaceholder возвращает placeholder для диалекта
func (b *Builder) getPlaceholder() squirrel.PlaceholderFormat {
	switch b.dialect {
	case PostgreSQL:
		return squirrel.Dollar
	case MySQL:
		return squirrel.Question
	case SQLite:
		return squirrel.Question
	default:
		return squirrel.Question
	}
}

// GetPlaceholderFormat возвращает формат placeholder для внешнего использования
func (b *Builder) GetPlaceholderFormat() squirrel.PlaceholderFormat {
	return b.getPlaceholder()
}

// BuildQuery собирает SQL запрос с параметрами
func (b *Builder) BuildQuery(query squirrel.Sqlizer) (string, []interface{}, error) {
	return query.ToSql()
}

// Execute выполняет запрос
func (b *Builder) Execute(query squirrel.Sqlizer) (sql.Result, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return b.db.Exec(sql, args...)
}

// Query выполняет SELECT запрос
func (b *Builder) Query(query squirrel.Sqlizer) (*sql.Rows, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	return b.db.Query(sql, args...)
}

// QueryRow выполняет SELECT запрос для одной строки
func (b *Builder) QueryRow(query squirrel.Sqlizer) *sql.Row {
	sql, args, err := query.ToSql()
	if err != nil {
		// Возвращаем row с ошибкой
		return b.db.QueryRow("SELECT 1 WHERE 1=0")
	}

	return b.db.QueryRow(sql, args...)
}
