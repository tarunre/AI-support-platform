package jwt

import (
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the JWT payload carried inside every access/refresh token.
type Claims struct {
	UserID   uint      `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	gojwt.RegisteredClaims
}

// PortalClaims is the payload for customer magic-link portal tokens.
// No user account is required — the token itself grants access to one ticket.
type PortalClaims struct {
	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	gojwt.RegisteredClaims
}

// TokenPair holds both the short-lived access token and the long-lived refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// GenerateTokenPair creates a signed access token (1 day) and refresh token (7 days).
func GenerateTokenPair(userID uint, tenantID uuid.UUID, email, role, accessSecret, refreshSecret string) (*TokenPair, error) {
	access, err := newSignedToken(userID, tenantID, email, role, 24*time.Hour, accessSecret)
	if err != nil {
		return nil, err
	}
	refresh, err := newSignedToken(userID, tenantID, email, role, 7*24*time.Hour, refreshSecret)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

// ValidateToken parses and verifies an access/refresh token string.
func ValidateToken(tokenStr, secret string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// GeneratePortalToken creates a 30-day signed token for customer portal access.
// The token encodes ticketID + customerEmail — no login required.
func GeneratePortalToken(ticketID, customerEmail, secret string) (string, error) {
	claims := &PortalClaims{
		TicketID:      ticketID,
		CustomerEmail: customerEmail,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret + "-portal"))
}

// ValidatePortalToken parses and verifies a customer portal magic-link token.
func ValidatePortalToken(tokenStr, secret string) (*PortalClaims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &PortalClaims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret + "-portal"), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*PortalClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid portal token claims")
	}
	return claims, nil
}

func newSignedToken(userID uint, tenantID uuid.UUID, email, role string, ttl time.Duration, secret string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
		},
	}
	return gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
