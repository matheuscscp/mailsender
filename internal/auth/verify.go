package auth

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/matheuscscp/mailsender/internal/auth/providers/github"
	"github.com/matheuscscp/mailsender/internal/auth/providers/google"
)

func Verify(ctx context.Context, l logrus.FieldLogger, token string) (string, bool) {
	for provider, verify := range map[string]func(context.Context, string) (string, error){
		"github": github.Verify,
		"google": google.Verify,
	} {
		sub, err := verify(ctx, token)
		if err == nil {
			return sub, true
		}
		l.WithError(err).WithField("provider", provider).Info("error verifying token with provider")
	}
	return "", false
}
