package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/ym/btc/internal/config"
)

const smtpTimeout = 30 * time.Second

type Notifier interface {
	Send(ctx context.Context, subject, body string)
	Enabled() bool
}

type LogNotifier struct{}

func (LogNotifier) Enabled() bool { return false }

func (LogNotifier) Send(ctx context.Context, subject, body string) {
	_ = ctx
	slog.Info("notification", "subject", subject, "body", strings.TrimSpace(body))
}

type EmailNotifier struct {
	cfg config.EmailConfig
}

func New(cfg config.EmailConfig) Notifier {
	if cfg.Enabled() {
		return &EmailNotifier{cfg: cfg}
	}
	return LogNotifier{}
}

func (e *EmailNotifier) Enabled() bool { return true }

func (e *EmailNotifier) Send(ctx context.Context, subject, body string) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, smtpTimeout)
	defer cancel()

	if err := e.send(ctx, subject, body); err != nil {
		slog.Error("email send failed", "subject", subject, "error", err)
		return
	}
	slog.Info("email sent", "subject", subject, "to", e.cfg.AlertTo)
}

func (e *EmailNotifier) send(ctx context.Context, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	msg := strings.Join([]string{
		fmt.Sprintf("To: %s", e.cfg.AlertTo),
		fmt.Sprintf("From: %s", e.cfg.Username),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.SMTPHost)
	return sendMail(ctx, addr, auth, e.cfg.Username, []string{e.cfg.AlertTo}, []byte(msg))
}

func sendMail(ctx context.Context, addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
