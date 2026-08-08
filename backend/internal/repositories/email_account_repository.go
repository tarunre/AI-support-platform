package repositories

import (
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailAccountRepository handles all database access for email_accounts.
type EmailAccountRepository struct {
	db *gorm.DB
}

func NewEmailAccountRepository(db *gorm.DB) *EmailAccountRepository {
	return &EmailAccountRepository{db: db}
}

func (r *EmailAccountRepository) Create(acc *models.EmailAccount) error {
	return r.db.Create(acc).Error
}

func (r *EmailAccountRepository) FindByID(tenantID uuid.UUID, id uint) (*models.EmailAccount, error) {
	var acc models.EmailAccount
	if err := r.db.Where("tenant_id = ? AND id = ?", tenantID, id).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *EmailAccountRepository) FindByAddress(tenantID uuid.UUID, address string) (*models.EmailAccount, error) {
	var acc models.EmailAccount
	if err := r.db.Where("tenant_id = ? AND email_address = ?", tenantID, address).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *EmailAccountRepository) Update(acc *models.EmailAccount) error {
	return r.db.Save(acc).Error
}

func (r *EmailAccountRepository) Delete(tenantID uuid.UUID, id uint) error {
	return r.db.Where("tenant_id = ?", tenantID).Delete(&models.EmailAccount{}, id).Error
}

func (r *EmailAccountRepository) ListByTenant(tenantID uuid.UUID) ([]models.EmailAccount, error) {
	var accounts []models.EmailAccount
	if err := r.db.Where("tenant_id = ? AND is_active = true", tenantID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// ListAllActive returns all active email accounts across all tenants (used by background workers).
func (r *EmailAccountRepository) ListAllActive() ([]models.EmailAccount, error) {
	var accounts []models.EmailAccount
	if err := r.db.Where("is_active = true").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// Count returns the total number of email accounts (all tenants).
func (r *EmailAccountRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.EmailAccount{}).Count(&count).Error
	return count, err
}

// CountActive returns the number of active email accounts (all tenants).
func (r *EmailAccountRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&models.EmailAccount{}).Where("is_active = true").Count(&count).Error
	return count, err
}
// FindActive returns all active email accounts across all tenants (alias for inbound worker).
func (r *EmailAccountRepository) FindActive() ([]models.EmailAccount, error) {
        return r.ListAllActive()
}