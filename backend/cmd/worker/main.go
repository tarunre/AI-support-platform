package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayush/supportiq/internal/ai/gemini"
	"github.com/ayush/supportiq/internal/ai/groq"
	"github.com/ayush/supportiq/internal/ai/provider"
	replyprovider "github.com/ayush/supportiq/internal/ai/reply/provider"
	"github.com/ayush/supportiq/internal/config"
	"github.com/ayush/supportiq/internal/database"
	emailattachments "github.com/ayush/supportiq/internal/email/attachments"
	"github.com/ayush/supportiq/internal/email/threading"
	"github.com/ayush/supportiq/internal/knowledge/retrieval"
	"github.com/ayush/supportiq/internal/models"
	orderspkg "github.com/ayush/supportiq/internal/orders"
	"github.com/ayush/supportiq/internal/queue/redisqueue"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/services"
	"github.com/ayush/supportiq/internal/utils"
	workerhandlers "github.com/ayush/supportiq/worker/handlers"
	"github.com/ayush/supportiq/worker/processor"
)

func main() {
	// ─── Config ─────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		utils.Logger.Fatalf("Worker: Failed to load config: %v", err)
	}

	if cfg.RedisURL == "" {
		utils.Logger.Fatal("Worker: REDIS_URL is required to run the worker")
	}

	utils.Logger.WithField("workers", cfg.WorkerCount).Info("Worker: Starting up")

	// ─── Database ────────────────────────────────────────────────────────────
	db, err := database.Connect(cfg)
	if err != nil {
		utils.Logger.Fatalf("Worker: Database connection failed: %v", err)
	}

	// ─── Redis queue ─────────────────────────────────────────────────────────
	redisQ, err := redisqueue.New(cfg.RedisURL, cfg.QueueName)
	if err != nil {
		utils.Logger.Fatalf("Worker: Redis connection failed: %v", err)
	}
	defer redisQ.Close()

	// ─── Repositories ────────────────────────────────────────────────────────
	ticketRepo := repositories.NewTicketRepository(db)
	activityRepo := repositories.NewActivityRepository(db)
	userRepo := repositories.NewUserRepository(db)
	knowledgeRepo := repositories.NewKnowledgeRepository(db)
	replyRepo := repositories.NewReplyRepository(db)
	jobRepo := repositories.NewJobRepository(db)

	// ─── AI providers — priority: Groq (free) > Gemini > Noop ───────────────
	var aiProv provider.Provider
	var replyProv replyprovider.ReplyProvider
	activeModel := cfg.GeminiModel
	if cfg.GroqAPIKey != "" {
		groqClient := groq.NewClientWithReplyConfig(
			cfg.GroqAPIKey, cfg.GroqModel,
			time.Duration(cfg.AITimeout)*time.Second,
			cfg.AIMaxRetries, cfg.MaxReplyTokens, cfg.ReplyTemperature,
		)
		aiProv = groqClient
		replyProv = groqClient
		activeModel = cfg.GroqModel
		utils.Logger.WithField("model", cfg.GroqModel).Info("Worker: Groq provider initialised")
	} else if cfg.GeminiAPIKey != "" {
		geminiClient := gemini.NewClientWithReplyConfig(
			cfg.GeminiAPIKey, cfg.GeminiModel,
			time.Duration(cfg.AITimeout)*time.Second,
			cfg.AIMaxRetries, cfg.MaxReplyTokens, cfg.ReplyTemperature,
		)
		aiProv = geminiClient
		replyProv = geminiClient
		utils.Logger.WithField("model", cfg.GeminiModel).Info("Worker: Gemini provider initialised")
	} else {
		aiProv = &provider.NoopProvider{}
		replyProv = &replyprovider.NoopReplyProvider{}
		utils.Logger.Warn("Worker: No API key set — AI jobs will fail")
	}

	// ─── Services ────────────────────────────────────────────────────────────
	knowledgeRetriever := retrieval.NewPostgresRetriever(knowledgeRepo)
	replySvc := services.NewReplyService(replyProv, knowledgeRetriever, ticketRepo, replyRepo, activityRepo, activeModel)

	// Wire email service so auto-approved replies get queued for outbound delivery.
	// The API server's outbound worker will pick them up and send via SMTP.
	emailAccountRepo := repositories.NewEmailAccountRepository(db)
	emailMessageRepo := repositories.NewEmailMessageRepository(db)
	emailAccountSvc := services.NewEmailAccountService(emailAccountRepo, cfg.JWTAccessSecret)
	threadDetector := threading.NewDetector(emailMessageRepo)
	attachStorage := emailattachments.NewLocalStorage(cfg.AttachmentPath)
	// Minimal AIService for the email service dependency (TriggerAIForTicket is never
	// called from within the worker, so a noop provider is safe here).
	minimalAISvc := services.NewAIService(&provider.NoopProvider{}, ticketRepo, activityRepo)
	emailSvc := services.NewEmailService(
		emailAccountRepo, emailMessageRepo,
		ticketRepo, activityRepo,
		emailAccountSvc, threadDetector,
		attachStorage, minimalAISvc, db,
	)
	// Include portal magic-link in outbound reply emails
	emailSvc.SetPortalConfig(cfg.AppURL, cfg.JWTAccessSecret)
	replySvc.SetEmailService(emailSvc)

	// Wire order lookup so AI replies include real order status
	orderLoader := orderspkg.NewLoader("storage/orders.json")
	replySvc.SetOrderLoader(orderLoader)

	// ─── Job handlers ────────────────────────────────────────────────────────
	aiHandler := workerhandlers.NewAIAnalysisHandler(ticketRepo, activityRepo, aiProv)
	aiHandler.SetUserRepo(userRepo)
	aiHandler.SetReplyGenerator(replySvc)
	replyHandler := workerhandlers.NewGenerateReplyHandler(replySvc)

	// ─── Processor ───────────────────────────────────────────────────────────
	proc := processor.New(redisQ, redisQ, jobRepo, cfg.WorkerCount, cfg.MaxRetries, cfg.RetryDelay)
	proc.SetTicketRepo(ticketRepo)
	proc.RegisterHandler(string(models.JobTypeAIAnalysis), aiHandler)
	proc.RegisterHandler(string(models.JobTypeGenerateReply), replyHandler)
	proc.RegisterHandler(string(models.JobTypeRegenerateReply), replyHandler)
	proc.RegisterHandler(string(models.JobTypeRetryAI), aiHandler)
	proc.RegisterHandler(string(models.JobTypeRetryReply), replyHandler)

	// ─── Graceful shutdown ───────────────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-quit
		utils.Logger.WithField("signal", sig.String()).Info("Worker: Shutdown signal received")
		cancel()
	}()

	utils.Logger.WithField("queue", cfg.QueueName).
		WithField("workers", cfg.WorkerCount).
		Info("Worker: Ready — listening for jobs")

	proc.Start(ctx) // blocks until shutdown
	utils.Logger.Info("Worker: Exited cleanly")
}
