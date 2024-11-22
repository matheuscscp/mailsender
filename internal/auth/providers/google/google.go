package google

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/matheuscscp/mailsender/internal/auth/common"
)

func Verify(ctx context.Context, token string) (string, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return "", fmt.Errorf("error creating google oidc provider: %w", err)
	}

	oidcToken, err := provider.VerifierContext(ctx, &oidc.Config{ClientID: common.Audience}).Verify(ctx, token)
	if err != nil {
		return "", fmt.Errorf("error verifying token as google oidc token: %w", err)
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := oidcToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("error extracting claims from google oidc token: %w", err)
	}

	if !claims.EmailVerified || claims.Email != "matheuscscp@gmail.com" {
		return "", fmt.Errorf("invalid email: %s (verified: %v)", claims.Email, claims.EmailVerified)
	}

	return claims.Email, nil
}
