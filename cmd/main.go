package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/matheuscscp/mailsender/internal/sendgrid"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	logrus.SetOutput(os.Stdout)
}

func main() {
	client, err := sendgrid.New()
	if err != nil {
		logrus.WithError(err).Fatal("error creating client")
	}

	// create context with cancelation on interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := client.SendEmail(ctx, "Test", "This is a test email.", ""); err != nil {
		_ = sendgrid.LogSendErrorAndGetStatusCode(logrus.StandardLogger(), err)
		os.Exit(1)
	}
}
