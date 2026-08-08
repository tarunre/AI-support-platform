package retrieval

import (
	"context"

	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// PostgresRetriever retrieves relevant knowledge documents using SQL keyword search.
type PostgresRetriever struct {
	repo *repositories.KnowledgeRepository
}

func NewPostgresRetriever(repo *repositories.KnowledgeRepository) *PostgresRetriever {
	return &PostgresRetriever{repo: repo}
}

func (r *PostgresRetriever) Retrieve(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]models.KnowledgeBase, error) {
	return r.repo.Search(ctx, tenantID, query, limit)
}
