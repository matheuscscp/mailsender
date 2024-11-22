package cloudfunctions

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sirupsen/logrus"

	"github.com/matheuscscp/mailsender/internal/sendgrid"
)

const (
	audience = "mailsender"
	email    = "client@mail-sender-442416.iam.gserviceaccount.com"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{DisableTimestamp: true})

	client := sendgrid.New()

	functions.HTTP("SendEmail", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var l logrus.FieldLogger = logrus.StandardLogger()

		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			const msg = "missing or invalid Authorization header"
			l.Error(msg)
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")

		provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			const msg = "error creating google oidc provider"
			l.WithError(err).Error(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		oidcToken, err := provider.VerifierContext(ctx, &oidc.Config{ClientID: audience}).Verify(ctx, token)
		if err != nil {
			const msg = "error verifying token as google oidc token"
			l.WithError(err).Error(msg)
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}

		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := oidcToken.Claims(&claims); err != nil {
			const msg = "error parsing google oidc token claims"
			l.WithError(err).Error(msg)
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}

		if !claims.EmailVerified || claims.Email != email {
			const msg = "forbidden"
			l.Error(msg)
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		var req struct {
			Subject          string `json:"subject"`
			PlainTextContent string `json:"plainTextContent"`
			HTMLContent      string `json:"htmlContent"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			const msg = "error decoding request as json"
			logrus.WithError(err).Error(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		l = l.WithField("request", req)

		if err := client.SendEmail(ctx, req.Subject, req.PlainTextContent, req.HTMLContent); err != nil {
			statusCode := sendgrid.LogSendErrorAndGetStatusCode(l, err)
			http.Error(w, "error sending email", statusCode)
			return
		}

		l.Info("email sent")
	})
}
