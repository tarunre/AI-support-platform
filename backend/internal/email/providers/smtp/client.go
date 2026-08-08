// Package smtp implements the providers.Sender interface using Go's standard net/smtp.
package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/ayush/supportiq/internal/email/providers"
	"github.com/google/uuid"
)

// Client sends email via SMTP with STARTTLS (port 587) or implicit TLS (port 465).
type Client struct {
	host        string
	port        int
	username    string
	password    string
	fromAddress string  // e.g. "support@example.com"
	fromName    string  // e.g. "Support Team"
	implicitTLS bool    // true = port 465 (wrap TCP in TLS)
}

// New creates a new SMTP client.
func New(host string, port int, username, password, fromAddress, fromName string, implicitTLS bool) *Client {
	return &Client{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		fromAddress: fromAddress,
		fromName:    fromName,
		implicitTLS: implicitTLS,
	}
}

// Send transmits the outbound message.
func (c *Client) Send(_ context.Context, msg providers.OutboundMessage) error {
	body := c.buildBody(msg)
	from := c.envelopeFrom()
	to := []string{msg.To}
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	if c.implicitTLS {
		return c.sendImplicitTLS(addr, from, to, body)
	}
	// STARTTLS path — smtp.SendMail negotiates STARTTLS automatically
	auth := smtp.PlainAuth("", c.username, c.password, c.host)
	return smtp.SendMail(addr, auth, from, to, body)
}

// TestConnection verifies credentials by opening a connection and logging in.
func (c *Client) TestConnection(_ context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	if c.implicitTLS {
		conn, err := tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp", addr,
			&tls.Config{ServerName: c.host},
		)
		if err != nil {
			return fmt.Errorf("smtp: TLS dial: %w", err)
		}
		cl, err := smtp.NewClient(conn, c.host)
		if err != nil {
			conn.Close()
			return err
		}
		defer cl.Close()
		return cl.Auth(smtp.PlainAuth("", c.username, c.password, c.host))
	}

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp: dial: %w", err)
	}
	cl, err := smtp.NewClient(conn, c.host)
	if err != nil {
		conn.Close()
		return err
	}
	defer cl.Close()
	if ok, _ := cl.Extension("STARTTLS"); ok {
		if err = cl.StartTLS(&tls.Config{ServerName: c.host}); err != nil {
			return err
		}
	}
	return cl.Auth(smtp.PlainAuth("", c.username, c.password, c.host))
}

// ── Internals ─────────────────────────────────────────────────────────────────

func (c *Client) sendImplicitTLS(addr, from string, to []string, body []byte) error {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 30 * time.Second},
		"tcp", addr,
		&tls.Config{ServerName: c.host},
	)
	if err != nil {
		return fmt.Errorf("smtp: TLS dial: %w", err)
	}
	cl, err := smtp.NewClient(conn, c.host)
	if err != nil {
		conn.Close()
		return err
	}
	defer cl.Close()

	if err = cl.Auth(smtp.PlainAuth("", c.username, c.password, c.host)); err != nil {
		return fmt.Errorf("smtp: auth: %w", err)
	}
	if err = cl.Mail(from); err != nil {
		return fmt.Errorf("smtp: MAIL FROM: %w", err)
	}
	for _, rcpt := range to {
		if err = cl.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp: RCPT TO %s: %w", rcpt, err)
		}
	}
	w, err := cl.Data()
	if err != nil {
		return fmt.Errorf("smtp: DATA: %w", err)
	}
	if _, err = w.Write(body); err != nil {
		return fmt.Errorf("smtp: write body: %w", err)
	}
	return w.Close()
}

func (c *Client) envelopeFrom() string {
	return c.fromAddress
}

// buildBody constructs a valid RFC 2822 message.
// Sends multipart/alternative when HTMLBody is provided.
func (c *Client) buildBody(msg providers.OutboundMessage) []byte {
	var buf bytes.Buffer

	from := c.fromAddress
	if c.fromName != "" {
		from = fmt.Sprintf("%s <%s>", c.fromName, c.fromAddress)
	}
	if msg.From != "" {
		from = msg.From
	}

	writeHeader := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
	}

	writeHeader("From", from)
	writeHeader("To", msg.To)
	writeHeader("Subject", msg.Subject)
	writeHeader("Message-ID", fmt.Sprintf("<%s>", msg.MessageID))
	writeHeader("Date", time.Now().UTC().Format(time.RFC1123Z))
	writeHeader("MIME-Version", "1.0")
	if msg.InReplyTo != "" {
		writeHeader("In-Reply-To", fmt.Sprintf("<%s>", strings.Trim(msg.InReplyTo, "<>")))
	}
	if msg.References != "" {
		writeHeader("References", msg.References)
	}

	if msg.HTMLBody != "" {
		boundary := "alt_" + uuid.New().String()[:16]
		writeHeader("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", boundary))
		fmt.Fprintf(&buf, "\r\n")

		// Plain text part
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=\"utf-8\"\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		encodeQP(&buf, msg.TextBody)
		fmt.Fprintf(&buf, "\r\n")

		// HTML part
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/html; charset=\"utf-8\"\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		encodeQP(&buf, msg.HTMLBody)
		fmt.Fprintf(&buf, "\r\n")

		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	} else {
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=\"utf-8\"\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: quoted-printable\r\n\r\n")
		encodeQP(&buf, msg.TextBody)
	}

	return buf.Bytes()
}

func encodeQP(w *bytes.Buffer, text string) {
	qpw := quotedprintable.NewWriter(w)
	_, _ = qpw.Write([]byte(text))
	_ = qpw.Close()
}
