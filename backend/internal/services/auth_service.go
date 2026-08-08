package services

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ayush/supportiq/internal/config"
	"github.com/ayush/supportiq/internal/dto"
	jwtpkg "github.com/ayush/supportiq/internal/jwt"
	"github.com/ayush/supportiq/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Sentinel errors used by handlers to determine response codes.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTenantSlugTaken    = errors.New("company slug already in use")
)

// AuthService contains all business logic for authentication.
type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAuthService constructs an AuthService via dependency injection.
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

// RegisterWithTenant creates a new tenant and the first admin user.
func (s *AuthService) RegisterWithTenant(req *dto.RegisterWithTenantRequest) (*dto.AuthResponse, int, error) {
	// Derive slug from company name if not provided
	slug := req.CompanySlug
	if slug == "" {
		slug = slugify(req.CompanyName)
	}

	// Check slug uniqueness
	var existing models.Tenant
	if err := s.db.Where("slug = ?", slug).First(&existing).Error; err == nil {
		return nil, http.StatusConflict, ErrTenantSlugTaken
	}

	// Check email uniqueness (global for now since slugs are unique)
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, http.StatusConflict, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Create tenant
	tenant := models.Tenant{
		Name:             req.CompanyName,
		Slug:             slug,
		SubscriptionPlan: models.PlanFree,
		Status:           models.TenantStatusActive,
		Timezone:         "UTC",
		BrandColor:       "#3B82F6",
	}

	// Create user as Admin within a transaction
	var user models.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}
		user = models.User{
			TenantID:     tenant.ID,
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(hash),
			Role:         models.RoleAdmin,
			IsActive:     true,
		}
		return tx.Create(&user).Error
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return s.buildAuthResponse(&user, tenant.ID)
}

// Login verifies credentials and returns a token pair.
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, int, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, http.StatusUnauthorized, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, http.StatusUnauthorized, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, http.StatusForbidden, errors.New("account is disabled")
	}

	return s.buildAuthResponse(&user, user.TenantID)
}

// RefreshToken validates a refresh token and returns a new token pair.
func (s *AuthService) RefreshToken(refreshToken string) (*dto.AuthResponse, int, error) {
	claims, err := jwtpkg.ValidateToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid or expired refresh token")
	}

	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, http.StatusUnauthorized, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, http.StatusForbidden, errors.New("account is disabled")
	}

	return s.buildAuthResponse(&user, user.TenantID)
}

// AgentJoin registers a new SupportAgent under an existing tenant identified
// by company_slug. No invite code required — the slug acts as the join key.
func (s *AuthService) AgentJoin(req *dto.AgentJoinRequest) (*dto.AuthResponse, int, error) {
	// Find the tenant
	var tenant models.Tenant
	if err := s.db.Where("slug = ? AND status = ?", req.CompanySlug, models.TenantStatusActive).First(&tenant).Error; err != nil {
		return nil, http.StatusNotFound, errors.New("company not found — check the company slug")
	}

	// Check email uniqueness within this tenant
	var existing models.User
	if err := s.db.Where("email = ? AND tenant_id = ?", req.Email, tenant.ID).First(&existing).Error; err == nil {
		return nil, http.StatusConflict, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	user := models.User{
		TenantID:     tenant.ID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.RoleSupportAgent,
		Team:         req.Team,
		IsActive:     true,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return s.buildAuthResponse(&user, tenant.ID)
}

// GetUserByID fetches a user record by primary key and returns a safe DTO.
func (s *AuthService) GetUserByID(id uint) (*dto.UserResponse, int, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, http.StatusNotFound, ErrUserNotFound
	}
	resp := toUserResponse(&user)
	return &resp, http.StatusOK, nil
}

// buildAuthResponse generates a JWT pair and assembles the AuthResponse DTO.
func (s *AuthService) buildAuthResponse(user *models.User, tenantID uuid.UUID) (*dto.AuthResponse, int, error) {
	pair, err := jwtpkg.GenerateTokenPair(
		user.ID,
		tenantID,
		user.Email,
		string(user.Role),
		s.cfg.JWTAccessSecret,
		s.cfg.JWTRefreshSecret,
	)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &dto.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toUserResponse(user),
	}, http.StatusOK, nil
}

// toUserResponse maps a User model to the safe public DTO.
func toUserResponse(u *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Role:     string(u.Role),
		Team:     u.Team,
		IsActive: u.IsActive,
	}
}

// slugify creates a URL-safe slug from a company name.
func slugify(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Keep only alphanumeric and dashes
	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "company"
	}
	return result
}
