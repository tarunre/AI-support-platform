package retrieval

import (
	"context"

	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
)

// Retriever defines the contract for knowledge document retrieval.
type Retriever interface {
	Retrieve(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]models.KnowledgeBase, error)
}
