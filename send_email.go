package cloudfunctions

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/sirupsen/logrus"

	"github.com/matheuscscp/mailsender/internal/providers/sendgrid"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{DisableTimestamp: true})

	client := sendgrid.New()

	functions.HTTP("SendEmail", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var l logrus.FieldLogger = logrus.StandardLogger()

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
