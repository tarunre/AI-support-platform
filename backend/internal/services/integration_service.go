package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	emailcrypto "github.com/ayush/supportiq/internal/email/crypto"
	"github.com/ayush/supportiq/internal/dto"
	"github.com/ayush/supportiq/internal/integrations"
	"github.com/ayush/supportiq/internal/integrations/provider"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IntegrationService handles business logic for integration management.
type IntegrationService struct {
	repo          *repositories.IntegrationRepository
	ticketRepo    *repositories.TicketRepository
	registry      *integrations.Registry
	encryptionKey string
}

func NewIntegrationService(
	repo *repositories.IntegrationRepository,
	ticketRepo *repositories.TicketRepository,
	registry *integrations.Registry,
	encryptionKey string,
) *IntegrationService {
	return &IntegrationService{repo: repo, ticketRepo: ticketRepo, registry: registry, encryptionKey: encryptionKey}
}

func (s *IntegrationService) List(tenantID uuid.UUID) ([]dto.IntegrationResponse, error) {
	integrationList, err := s.repo.FindAll(tenantID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.IntegrationResponse, 0, len(integrationList))
	for _, intg := range integrationList {
		resp = append(resp, toIntegrationResponse(intg))
	}
	return resp, nil
}

func (s *IntegrationService) Create(tenantID uuid.UUID, req dto.CreateIntegrationRequest, userID uint) (*dto.IntegrationResponse, error) {
	prov, err := s.registry.Build(req.Provider, req.Config)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	_ = prov

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	encrypted, err := emailcrypto.Encrypt(s.encryptionKey, string(configJSON))
	if err != nil {
		return nil, fmt.Errorf("encrypt config: %w", err)
	}

	intg := &models.Integration{
		TenantID:      tenantID,
		Provider:      models.IntegrationProvider(req.Provider),
		Name:          req.Name,
		Configuration: encrypted,
		Status:        models.IntegrationStatusInactive,
		Enabled:       req.Enabled,
		CreatedBy:     userID,
	}
	if err := s.repo.Create(intg); err != nil {
		return nil, err
	}
	resp := toIntegrationResponse(*intg)
	return &resp, nil
}

func (s *IntegrationService) Update(tenantID uuid.UUID, id uint, req dto.UpdateIntegrationRequest) (*dto.IntegrationResponse, error) {
	intg, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		intg.Name = *req.Name
	}
	if req.Enabled != nil {
		intg.Enabled = *req.Enabled
	}
	if req.Config != nil {
		if _, err := s.registry.Build(string(intg.Provider), req.Config); err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}
		configJSON, _ := json.Marshal(req.Config)
		encrypted, err := emailcrypto.Encrypt(s.encryptionKey, string(configJSON))
		if err != nil {
			return nil, fmt.Errorf("encrypt config: %w", err)
		}
		intg.Configuration = encrypted
		intg.Status = models.IntegrationStatusInactive
	}

	if err := s.repo.Update(intg); err != nil {
		return nil, err
	}
	resp := toIntegrationResponse(*intg)
	return &resp, nil
}

func (s *IntegrationService) Delete(tenantID uuid.UUID, id uint) error {
	return s.repo.Delete(tenantID, id)
}

func (s *IntegrationService) TestConnection(ctx context.Context, tenantID uuid.UUID, id uint) error {
	intg, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return err
	}

	prov, err := s.buildProvider(*intg)
	if err != nil {
		s.setError(intg, err.Error())
		return err
	}

	if err := prov.TestConnection(ctx); err != nil {
		s.setError(intg, err.Error())
		return err
	}

	now := time.Now()
	intg.Status = models.IntegrationStatusActive
	intg.ErrorMessage = ""
	intg.LastSyncAt = &now
	_ = s.repo.Update(intg)
	return nil
}

func (s *IntegrationService) GetTicketIntegrations(tenantID uuid.UUID, ticketID uuid.UUID) ([]dto.TicketIntegrationResponse, error) {
	items, err := s.repo.FindTicketIntegrations(tenantID, ticketID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.TicketIntegrationResponse, 0, len(items))
	for _, item := range items {
		r := dto.TicketIntegrationResponse{
			ID:            item.ID,
			IntegrationID: item.IntegrationID,
			ExternalKey:   item.ExternalKey,
			ExternalURL:   item.ExternalURL,
			SyncedAt:      item.SyncedAt,
			CreatedAt:     item.CreatedAt,
		}
		if item.Integration != nil {
			r.Provider = string(item.Integration.Provider)
			r.Name = item.Integration.Name
		}
		resp = append(resp, r)
	}
	return resp, nil
}

func (s *IntegrationService) CreateJiraIssue(ctx context.Context, tenantID uuid.UUID, ticketID uuid.UUID) (*dto.TicketIntegrationResponse, error) {
	return s.createIssue(ctx, tenantID, ticketID, "jira")
}

func (s *IntegrationService) CreateLinearIssue(ctx context.Context, tenantID uuid.UUID, ticketID uuid.UUID) (*dto.TicketIntegrationResponse, error) {
	return s.createIssue(ctx, tenantID, ticketID, "linear")
}

func (s *IntegrationService) CreateGitHubIssue(ctx context.Context, tenantID uuid.UUID, ticketID uuid.UUID) (*dto.TicketIntegrationResponse, error) {
	return s.createIssue(ctx, tenantID, ticketID, "github")
}

func (s *IntegrationService) createIssue(ctx context.Context, tenantID uuid.UUID, ticketID uuid.UUID, providerType string) (*dto.TicketIntegrationResponse, error) {
	ticket, err := s.ticketRepo.FindByID(tenantID, ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	providerIntgs, err := s.repo.FindByProvider(tenantID, models.IntegrationProvider(providerType))
	if err != nil || len(providerIntgs) == 0 {
		return nil, fmt.Errorf("%s integration not configured", providerType)
	}
	intg := providerIntgs[0]

	prov, err := s.buildProvider(intg)
	if err != nil {
		return nil, err
	}

	issueProv, ok := prov.(provider.IssueProvider)
	if !ok {
		return nil, fmt.Errorf("%s does not support issue creation", providerType)
	}

	ref, err := issueProv.CreateIssue(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("create issue: %w", err)
	}

	now := time.Now()
	ti := &models.TicketIntegration{
		TenantID:      tenantID,
		TicketID:      ticket.ID,
		IntegrationID: intg.ID,
		ExternalID:    ref.ExternalID,
		ExternalKey:   ref.ExternalKey,
		ExternalURL:   ref.ExternalURL,
		SyncedAt:      &now,
	}
	if err := s.repo.CreateTicketIntegration(ti); err != nil {
		_ = err
	}

	return &dto.TicketIntegrationResponse{
		ID:            ti.ID,
		IntegrationID: intg.ID,
		Provider:      string(intg.Provider),
		Name:          intg.Name,
		ExternalKey:   ref.ExternalKey,
		ExternalURL:   ref.ExternalURL,
		SyncedAt:      &now,
		CreatedAt:     ti.CreatedAt,
	}, nil
}

func (s *IntegrationService) ListEvents(tenantID uuid.UUID, integrationID uint) ([]dto.IntegrationEventResponse, error) {
	if _, err := s.repo.FindByID(tenantID, integrationID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("integration not found")
		}
		return nil, err
	}
	events, err := s.repo.FindPendingEvents(tenantID, 100)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.IntegrationEventResponse, 0, len(events))
	for _, evt := range events {
		if evt.IntegrationID != integrationID {
			continue
		}
		resp = append(resp, dto.IntegrationEventResponse{
			ID:            evt.ID,
			IntegrationID: evt.IntegrationID,
			EventType:     evt.EventType,
			Status:        string(evt.Status),
			RetryCount:    evt.RetryCount,
			ErrorMessage:  evt.ErrorMessage,
			CreatedAt:     evt.CreatedAt,
			ProcessedAt:   evt.ProcessedAt,
		})
	}
	return resp, nil
}

func (s *IntegrationService) buildProvider(intg models.Integration) (provider.Provider, error) {
	plaintext, err := emailcrypto.Decrypt(s.encryptionKey, intg.Configuration)
	if err != nil {
		return nil, fmt.Errorf("decrypt config: %w", err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(plaintext), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return s.registry.Build(string(intg.Provider), cfg)
}

func (s *IntegrationService) setError(intg *models.Integration, msg string) {
	intg.Status = models.IntegrationStatusError
	intg.ErrorMessage = msg
	_ = s.repo.Update(intg)
}

func toIntegrationResponse(intg models.Integration) dto.IntegrationResponse {
	return dto.IntegrationResponse{
		ID:           intg.ID,
		Provider:     string(intg.Provider),
		Name:         intg.Name,
		Status:       string(intg.Status),
		Enabled:      intg.Enabled,
		CreatedBy:    intg.CreatedBy,
		LastSyncAt:   intg.LastSyncAt,
		ErrorMessage: intg.ErrorMessage,
		CreatedAt:    intg.CreatedAt,
		UpdatedAt:    intg.UpdatedAt,
	}
}
