package repositories

import (
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository encapsulates all database access for users.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID returns a user by primary key (unscoped — used by auth middleware).
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListByRole returns active users with the given role within a tenant.
func (r *UserRepository) ListByRole(tenantID uuid.UUID, role models.Role) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("tenant_id = ? AND role = ? AND is_active = true", tenantID, role).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ListByTeam returns active users belonging to a specific team within a tenant.
// SupportAgents are returned before Admins so proper agents are preferred.
func (r *UserRepository) ListByTeam(tenantID uuid.UUID, team string) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("tenant_id = ? AND team = ? AND is_active = true", tenantID, team).
		Order("CASE WHEN role = 'SupportAgent' THEN 0 ELSE 1 END, id").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindByEmail looks up a user by email within a tenant.
func (r *UserRepository) FindByEmail(tenantID uuid.UUID, email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("tenant_id = ? AND email = ?", tenantID, email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
