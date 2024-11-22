package auth

import (
	"context"
	"errors"

	"github.com/matheuscscp/mailsender/internal/auth/providers/github"
)

func Verify(ctx context.Context, token string) (string, error) {
	for _, verify := range []func(context.Context, string) (string, error){
		github.Verify,
	} {
		if sub, err := verify(ctx, token); err == nil {
			return sub, nil
		}
	}
	return "", errors.New("invalid token")
}
