package testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
	"github.com/landly/backend/internal/repositories"
)

// SetupTestDB establishes a clean PostgreSQL connection for integration tests
func SetupTestDB(t *testing.T) *query.Builder {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		host := envOrDefault("TEST_POSTGRES_HOST", "")
		if host == "" {
			t.Skip("TEST_POSTGRES_DSN not set, skipping integration test")
		}

		port := envOrDefault("TEST_POSTGRES_PORT", "5432")
		user := envOrDefault("TEST_POSTGRES_USER", "postgres")
		password := envOrDefault("TEST_POSTGRES_PASSWORD", "postgres")
		dbname := envOrDefault("TEST_POSTGRES_DB", "landly_test")

		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "failed to open test database")

	require.NoError(t, db.Ping(), "failed to ping test database")
	require.NoError(t, createTestSchema(db), "failed to ensure test schema")

	qb := query.NewBuilder(query.PostgreSQL, db)

	t.Cleanup(func() {
		cleanupTestDB(t, db)
		db.Close()
	})

	return qb
}

// SetupTestRedis establishes a Redis client for integration tests
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	addr := envOrDefault("TEST_REDIS_ADDR", "")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set, skipping integration test")
	}

	client := redis.NewClient(&redis.Options{Addr: addr, DB: 0})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.Ping(ctx).Err(), "failed to connect to test redis")

	t.Cleanup(func() {
		client.FlushDB(context.Background())
		client.Close()
	})

	return client
}

// CreateTestUser inserts a user with a random email and password. The plaintext password is returned for convenience.
func CreateTestUser(t *testing.T, qb *query.Builder, email, password string) (*domain.User, string) {
	t.Helper()

	if email == "" {
		email = fmt.Sprintf("test-user-%s@example.com", uuid.New().String()[:8])
	}
	if password == "" {
		password = fmt.Sprintf("Password-%s", uuid.New().String()[:8])
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password for test user")

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	userRepo := repositories.NewUserRepository(qb)
	require.NoError(t, userRepo.Create(context.Background(), user), "failed to create test user")

	return user, password
}

// CreateTestProject inserts a project tied to the provided user ID
func CreateTestProject(t *testing.T, qb *query.Builder, userID uuid.UUID, name, niche string) *domain.Project {
	t.Helper()

	if name == "" {
		name = fmt.Sprintf("Test Project %s", uuid.New().String()[:8])
	}
	if niche == "" {
		niche = "SaaS"
	}

	project := &domain.Project{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       name,
		Niche:      niche,
		Status:     "draft",
		SchemaJSON: "",
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	projectRepo := repositories.NewProjectRepository(qb)
	require.NoError(t, projectRepo.Create(context.Background(), project), "failed to create test project")

	return project
}

// CreateTestIntegration inserts an integration for the given project
func CreateTestIntegration(t *testing.T, qb *query.Builder, projectID uuid.UUID, integrationType domain.IntegrationType, config string) *domain.Integration {
	t.Helper()

	if integrationType == "" {
		integrationType = domain.IntegrationTypeStripe
	}
	if config == "" {
		config = fmt.Sprintf("{\"apiKey\":\"test-%s\"}", uuid.New().String()[:8])
	}

	integration := &domain.Integration{
		ID:        uuid.New(),
		ProjectID: projectID,
		Type:      string(integrationType),
		Config:    config,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	integrationRepo := repositories.NewIntegrationRepository(qb)
	require.NoError(t, integrationRepo.Create(context.Background(), integration), "failed to create test integration")

	return integration
}

func createTestSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS projects (
		id UUID PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		niche VARCHAR(255) NOT NULL,
		schema_json TEXT,
		status VARCHAR(50) NOT NULL DEFAULT 'draft',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS generation_sessions (
		id UUID PRIMARY KEY,
		project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		prompt TEXT NOT NULL,
		model VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		schema_json TEXT,
		completed_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS integrations (
		id UUID PRIMARY KEY,
		project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		type VARCHAR(50) NOT NULL,
		config TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(project_id, type)
	);

	CREATE TABLE IF NOT EXISTS publish_targets (
		id UUID PRIMARY KEY,
		project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		subdomain VARCHAR(255) UNIQUE NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'draft',
		last_published_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS analytics_events (
		id UUID PRIMARY KEY,
		project_id UUID NOT NULL,
		event_type VARCHAR(100) NOT NULL,
		path VARCHAR(500),
		referrer VARCHAR(500),
		user_agent VARCHAR(500),
		ip_address VARCHAR(50),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := db.Exec(schema)
	return err
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"analytics_events",
		"publish_targets",
		"integrations",
		"generation_sessions",
		"projects",
		"users",
	}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Logf("warning: failed to truncate table %s: %v", table, err)
		}
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
