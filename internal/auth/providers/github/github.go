package github

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/matheuscscp/mailsender/internal/auth/common"
)

func Verify(ctx context.Context, token string) (string, error) {
	provider, err := oidc.NewProvider(ctx, "https://token.actions.githubusercontent.com")
	if err != nil {
		return "", fmt.Errorf("error creating github actions provider: %w", err)
	}

	oidcToken, err := provider.VerifierContext(ctx, &oidc.Config{ClientID: common.Audience}).Verify(ctx, token)
	if err != nil {
		return "", fmt.Errorf("error verifying token as github actions oidc token: %w", err)
	}

	var claims struct {
		Subject         string `json:"sub"`
		RepositoryOwner string `json:"repository_owner"`
	}
	if err := oidcToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("error extracting claims from github actions oidc token: %w", err)
	}

	if claims.RepositoryOwner != "matheuscscp" {
		return "", fmt.Errorf("invalid repository owner: %s", claims.RepositoryOwner)
	}

	return claims.Subject, nil
}
