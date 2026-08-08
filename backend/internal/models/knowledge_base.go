package models

import (
	"time"

	"github.com/google/uuid"
)

// KnowledgeCategory represents the type of knowledge document.
type KnowledgeCategory string

const (
	KBCategoryFAQ          KnowledgeCategory = "FAQ"
	KBCategoryRefund       KnowledgeCategory = "Refund Policy"
	KBCategoryShipping     KnowledgeCategory = "Shipping Policy"
	KBCategorySubscription KnowledgeCategory = "Subscription Policy"
	KBCategoryAccount      KnowledgeCategory = "Account Policy"
	KBCategoryPayment      KnowledgeCategory = "Payment Policy"
	KBCategoryGeneral      KnowledgeCategory = "General Documentation"
	KBCategoryEngineering  KnowledgeCategory = "Engineering Team"
)

// KnowledgeBase is a company knowledge document used to ground AI replies (RAG).
type KnowledgeBase struct {
	ID        uint              `gorm:"primarykey;autoIncrement"   json:"id"`
	TenantID  uuid.UUID         `gorm:"type:uuid;not null;default:'00000000-0000-0000-0000-000000000000';index" json:"tenant_id"`
	Title     string            `gorm:"type:varchar(255);not null" json:"title"`
	Category  KnowledgeCategory `gorm:"type:varchar(50);not null"  json:"category"`
	Content   string            `gorm:"type:text;not null"         json:"content"`
	IsActive  bool              `gorm:"not null;default:true"      json:"is_active"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
