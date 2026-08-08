package services

import (
	"fmt"
	"net/http"

	"github.com/ayush/supportiq/internal/dto"
	emailcrypto "github.com/ayush/supportiq/internal/email/crypto"
	emailproviders "github.com/ayush/supportiq/internal/email/providers"
	imapprovider "github.com/ayush/supportiq/internal/email/providers/imap"
	smtpprovider "github.com/ayush/supportiq/internal/email/providers/smtp"
	"github.com/ayush/supportiq/internal/models"
	"github.com/ayush/supportiq/internal/repositories"
	"github.com/ayush/supportiq/internal/utils"
	"github.com/google/uuid"
)

// EmailAccountService manages CRUD and connection testing for email accounts.
type EmailAccountService struct {
	repo          *repositories.EmailAccountRepository
	encryptionKey string
}

func NewEmailAccountService(repo *repositories.EmailAccountRepository, encryptionKey string) *EmailAccountService {
	return &EmailAccountService{repo: repo, encryptionKey: encryptionKey}
}

// List returns all active accounts for a tenant.
func (s *EmailAccountService) List(tenantID uuid.UUID) ([]dto.EmailAccountResponse, int, error) {
	accounts, err := s.repo.ListByTenant(tenantID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to list email accounts")
	}
	resp := make([]dto.EmailAccountResponse, len(accounts))
	for i := range accounts {
		resp[i] = dto.ToEmailAccountResponse(&accounts[i])
	}
	return resp, http.StatusOK, nil
}

// Create persists a new email account after encrypting the password.
func (s *EmailAccountService) Create(tenantID uuid.UUID, req *dto.CreateEmailAccountRequest) (*dto.EmailAccountResponse, int, error) {
	encrypted, err := emailcrypto.Encrypt(s.encryptionKey, req.Password)
	if err != nil {
		utils.Logger.WithError(err).Error("EmailAccount: password encryption failed")
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to secure credentials")
	}

	imapPort := req.IMAPPort
	if imapPort == 0 {
		imapPort = 993
	}
	smtpPort := req.SMTPPort
	if smtpPort == 0 {
		smtpPort = 587
	}

	account := &models.EmailAccount{
		TenantID:          tenantID,
		Provider:          models.EmailProvider(req.Provider),
		EmailAddress:      req.EmailAddress,
		DisplayName:       req.DisplayName,
		IMAPHost:          req.IMAPHost,
		IMAPPort:          imapPort,
		IMAPUseTLS:        req.IMAPUseTLS,
		SMTPHost:          req.SMTPHost,
		SMTPPort:          smtpPort,
		SMTPImplicitTLS:   req.SMTPImplicitTLS,
		Username:          req.Username,
		EncryptedPassword: encrypted,
		IsActive:          true,
	}

	if err := s.repo.Create(account); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create email account")
	}

	resp := dto.ToEmailAccountResponse(account)
	return &resp, http.StatusCreated, nil
}

// Update applies partial updates to an existing account.
func (s *EmailAccountService) Update(tenantID uuid.UUID, id uint, req *dto.UpdateEmailAccountRequest) (*dto.EmailAccountResponse, int, error) {
	account, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("email account not found")
	}

	if req.DisplayName != nil {
		account.DisplayName = *req.DisplayName
	}
	if req.IMAPHost != nil {
		account.IMAPHost = *req.IMAPHost
	}
	if req.IMAPPort != nil {
		account.IMAPPort = *req.IMAPPort
	}
	if req.IMAPUseTLS != nil {
		account.IMAPUseTLS = *req.IMAPUseTLS
	}
	if req.SMTPHost != nil {
		account.SMTPHost = *req.SMTPHost
	}
	if req.SMTPPort != nil {
		account.SMTPPort = *req.SMTPPort
	}
	if req.SMTPImplicitTLS != nil {
		account.SMTPImplicitTLS = *req.SMTPImplicitTLS
	}
	if req.Username != nil {
		account.Username = *req.Username
	}
	if req.Password != nil && *req.Password != "" {
		encrypted, err := emailcrypto.Encrypt(s.encryptionKey, *req.Password)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to secure credentials")
		}
		account.EncryptedPassword = encrypted
	}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	if err := s.repo.Update(account); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to update email account")
	}

	resp := dto.ToEmailAccountResponse(account)
	return &resp, http.StatusOK, nil
}

// Delete removes an email account by ID.
func (s *EmailAccountService) Delete(tenantID uuid.UUID, id uint) (int, error) {
	if _, err := s.repo.FindByID(tenantID, id); err != nil {
		return http.StatusNotFound, fmt.Errorf("email account not found")
	}
	if err := s.repo.Delete(tenantID, id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to delete email account")
	}
	return http.StatusOK, nil
}

// TestSMTP tests the SMTP connection for an account.
func (s *EmailAccountService) TestSMTP(tenantID uuid.UUID, id uint) error {
	account, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return fmt.Errorf("email account not found")
	}
	if account.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	pass, err := emailcrypto.Decrypt(s.encryptionKey, account.EncryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials")
	}
	client := smtpprovider.New(account.SMTPHost, account.SMTPPort, account.Username, pass, account.EmailAddress, account.DisplayName, account.SMTPImplicitTLS)
	return client.TestConnection(nil)
}

// TestIMAP tests the IMAP connection for an account.
func (s *EmailAccountService) TestIMAP(tenantID uuid.UUID, id uint) error {
	account, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return fmt.Errorf("email account not found")
	}
	if account.IMAPHost == "" {
		return fmt.Errorf("IMAP host not configured")
	}
	pass, err := emailcrypto.Decrypt(s.encryptionKey, account.EncryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt credentials")
	}
	client := imapprovider.New(account.IMAPHost, account.IMAPPort, account.Username, pass, account.IMAPUseTLS)
	return client.TestConnection(nil)
}

// BuildSender decrypts credentials and returns a ready-to-use Sender.
func (s *EmailAccountService) BuildSender(account *models.EmailAccount) (emailproviders.Sender, error) {
	pass, err := emailcrypto.Decrypt(s.encryptionKey, account.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SMTP credentials")
	}
	return smtpprovider.New(account.SMTPHost, account.SMTPPort, account.Username, pass, account.EmailAddress, account.DisplayName, account.SMTPImplicitTLS), nil
}

// BuildReceiver decrypts credentials and returns a ready-to-use IMAP Receiver.
// If account.LastSyncAt is set, the receiver will only fetch emails received
// after that timestamp — preventing re-processing of existing inbox emails.
func (s *EmailAccountService) BuildReceiver(account *models.EmailAccount) (emailproviders.Receiver, error) {
	pass, err := emailcrypto.Decrypt(s.encryptionKey, account.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt IMAP credentials")
	}
	cl := imapprovider.New(account.IMAPHost, account.IMAPPort, account.Username, pass, account.IMAPUseTLS)
	if account.LastSyncAt != nil {
		cl.SetSince(*account.LastSyncAt)
	}
	return cl, nil
}
