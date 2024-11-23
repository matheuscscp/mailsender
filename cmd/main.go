package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/matheuscscp/mailsender/internal/providers/sendgrid"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	logrus.SetOutput(os.Stdout)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage:\n\n\tgo run github.com/matheuscscp/mailsender/cmd <subject> <plain-text-content> <html-content>")
		os.Exit(1)
	}
	subject := os.Args[1]
	plainTextContent := os.Args[2]
	htmlContent := os.Args[3]

	client, err := sendgrid.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating sendgrid client: %v\n", err)
		os.Exit(2)
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

	if err := client.SendEmail(ctx, subject, plainTextContent, htmlContent); err != nil {
		_ = sendgrid.LogSendErrorAndGetStatusCode(logrus.StandardLogger(), err)
		os.Exit(3)
	}
}
