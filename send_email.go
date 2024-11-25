package cloudfunctions

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/sirupsen/logrus"

	"github.com/matheuscscp/mailsender/internal/auth"
	"github.com/matheuscscp/mailsender/internal/providers/sendgrid"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{DisableTimestamp: true})

	client, err := sendgrid.New()
	if err != nil {
		logrus.WithError(err).Fatal("error creating sendgrid client")
	}

	functions.HTTP("SendEmail", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var l logrus.FieldLogger = logrus.WithField("caller", r.Header.Get("X-Caller"))

		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			const msg = "missing or invalid authorization header"
			l.Error(msg)
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")

		sub, ok := auth.Verify(ctx, l, token)
		if !ok {
			const msg = "error verifying token"
			l.Error(msg)
			http.Error(w, msg, http.StatusUnauthorized)
			return
		}
		l = l.WithField("subject", sub)

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
