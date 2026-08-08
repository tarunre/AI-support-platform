package services

import (
	"fmt"
	"math"

	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
)

// KnowledgeService handles all business logic for knowledge base management.
type KnowledgeService struct {
	repo *repositories.KnowledgeRepository
}

func NewKnowledgeService(repo *repositories.KnowledgeRepository) *KnowledgeService {
	return &KnowledgeService{repo: repo}
}

func (s *KnowledgeService) Create(tenantID uuid.UUID, req dto.CreateKnowledgeRequest) (*models.KnowledgeBase, error) {
	if !isValidKBCategory(req.Category) {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}
	doc := &models.KnowledgeBase{
		TenantID: tenantID,
		Title:    req.Title,
		Category: models.KnowledgeCategory(req.Category),
		Content:  req.Content,
		IsActive: true,
	}
	if err := s.repo.Create(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *KnowledgeService) GetByID(tenantID uuid.UUID, id uint) (*models.KnowledgeBase, error) {
	return s.repo.FindByID(tenantID, id)
}

func (s *KnowledgeService) Update(tenantID uuid.UUID, id uint, req dto.UpdateKnowledgeRequest) (*models.KnowledgeBase, error) {
	doc, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Category != "" {
		if !isValidKBCategory(req.Category) {
			return nil, fmt.Errorf("invalid category: %s", req.Category)
		}
		doc.Category = models.KnowledgeCategory(req.Category)
	}
	if req.Content != "" {
		doc.Content = req.Content
	}
	if req.IsActive != nil {
		doc.IsActive = *req.IsActive
	}

	if err := s.repo.Update(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *KnowledgeService) Delete(tenantID uuid.UUID, id uint) error {
	if _, err := s.repo.FindByID(tenantID, id); err != nil {
		return err
	}
	return s.repo.Delete(tenantID, id)
}

func (s *KnowledgeService) List(tenantID uuid.UUID, q dto.ListKnowledgeQuery) ([]models.KnowledgeBase, int64, dto.ListKnowledgeResponse, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	limit := q.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	docs, total, err := s.repo.List(tenantID, q.Search, q.Category, q.ActiveOnly, page, limit)
	if err != nil {
		return nil, 0, dto.ListKnowledgeResponse{}, err
	}

	items := make([]dto.KnowledgeResponse, len(docs))
	for i, d := range docs {
		items[i] = toKnowledgeResponse(d)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return docs, total, dto.ListKnowledgeResponse{
		Items:       items,
		TotalCount:  total,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
	}, nil
}

func toKnowledgeResponse(d models.KnowledgeBase) dto.KnowledgeResponse {
	return dto.KnowledgeResponse{
		ID:        d.ID,
		Title:     d.Title,
		Category:  string(d.Category),
		Content:   d.Content,
		IsActive:  d.IsActive,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func isValidKBCategory(cat string) bool {
	valid := map[string]struct{}{
		string(models.KBCategoryFAQ):          {},
		string(models.KBCategoryRefund):       {},
		string(models.KBCategoryShipping):     {},
		string(models.KBCategorySubscription): {},
		string(models.KBCategoryAccount):      {},
		string(models.KBCategoryPayment):      {},
		string(models.KBCategoryGeneral):      {},
	}
	_, ok := valid[cat]
	return ok
}
