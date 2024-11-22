package sendgrid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/sendgrid/rest"
	sg "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

type Client struct {
	*sg.Client
}

type APIError struct {
	*rest.Response
}

const myName = "Matheus Pimenta"

var (
	from = mail.NewEmail(myName, "no-reply@matheuspimenta.dev")
	to   = mail.NewEmail(myName, "matheuscscp@gmail.com")
)

func New() (*Client, error) {
	if b, err := os.ReadFile("key.txt"); err == nil {
		return &Client{sg.NewSendClient(string(b))}, nil
	}
	if key := os.Getenv("SENDGRID_API_KEY"); key != "" {
		return &Client{sg.NewSendClient(key)}, nil
	}
	return nil, errors.New("sendgrid api key not found")
}

func (c *Client) SendEmail(ctx context.Context, subject, plainTextContent, htmlContent string) error {
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	resp, err := c.SendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending with context: %w", err)
	}

	if c := resp.StatusCode; c < 200 || 300 <= c {
		return APIError{resp}
	}

	return nil
}

func (a APIError) Error() string {
	b, err := json.Marshal(a.Response)
	if err != nil {
		return fmt.Sprintf("error marshaling response: %v (%+v)", err, a.Response)
	}
	return string(b)
}

func LogSendErrorAndGetStatusCode(l logrus.FieldLogger, err error) (statusCode int) {
	statusCode = http.StatusInternalServerError

	var el logrus.FieldLogger = l.WithError(err)

	// enrich log with sendgrid response if error is of that type
	if apiErr := (APIError{}); errors.As(err, &apiErr) {
		statusCode = apiErr.Response.StatusCode
		var body any = apiErr.Response.Body

		// try to unmarshal response body if it's json
		if ct, ok := apiErr.Response.Headers["Content-Type"]; ok && len(ct) > 0 && ct[0] == "application/json" {
			if err := json.Unmarshal([]byte(apiErr.Response.Body), &body); err != nil {
				l.WithError(err).Error("error unmarshaling sendgrid error response body as json")
				body = apiErr.Response.Body
			}
		}

		el = l.WithField("sendGridResponse", logrus.Fields{
			"statusCode": apiErr.Response.StatusCode,
			"headers":    apiErr.Response.Headers,
			"body":       body,
		})
	}

	el.Error("error sending email")

	return
}
