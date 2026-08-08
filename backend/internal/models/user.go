package models

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin        Role = "Admin"
	RoleSupportAgent Role = "SupportAgent"
	RoleSuperAdmin   Role = "SuperAdmin"
)

// User represents a registered user in the system.
// PasswordHash is never serialised to JSON (json:"-").
// SuperAdmin users have TenantID == uuid.Nil.
type User struct {
	ID           uint      `gorm:"primarykey;autoIncrement"                                           json:"id"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	Name         string    `gorm:"type:varchar(100);not null"                                         json:"name"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_email"             json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null"                                         json:"-"`
	Role         Role      `gorm:"type:varchar(20);not null;default:'SupportAgent'"                    json:"role"`
	Team         string    `gorm:"type:varchar(50)"                                                   json:"team,omitempty"`
	IsActive     bool      `gorm:"not null;default:true"                                              json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}
