package database

import (
	"fmt"
	"time"

	"github.com/ayush/supportiq/internal/config"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect establishes a PostgreSQL connection pool via GORM and returns the DB handle.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve underlying sql.DB: %w", err)
	}

	// Connection pool tuning
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Ensure any new columns are added safely before running AutoMigrate.
	// This avoids "column already exists" errors on PostgreSQL when a previous
	// run already applied the migration but AutoMigrate tries again.
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS team varchar(50) DEFAULT ''")

	if err := db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Ticket{},
		&models.TicketCounter{},
		&models.TicketNote{},
		&models.TicketActivity{},
		&models.TicketComment{},
		&models.KnowledgeBase{},
		&models.AIReply{},
		&models.BackgroundJob{},
		&models.EmailAccount{},
		&models.EmailMessage{},
		&models.DailyTicketMetrics{},
		&models.AgentMetrics{},
		&models.AIMetrics{},
		&models.Report{},
		&models.Integration{},
		&models.IntegrationEvent{},
		&models.TicketIntegration{},
		&models.SLAPolicy{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// TicketCounter now uses TenantID as primary key — no seeding needed (created on first ticket per tenant)
	utils.Logger.Info("Database connected and migrations applied")
	return db, nil
}
