// Package emailparser parses raw RFC 2822 email bytes using Go's standard library.
package emailparser

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"github.com/ayush/supportiq/internal/email/providers"
)

// Parse parses a raw RFC 2822 email message and returns a ParsedEmail.
func Parse(raw []byte) (*providers.ParsedEmail, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("emailparser: read message: %w", err)
	}

	p := &providers.ParsedEmail{}

	// ── Headers ────────────────────────────────────────────────────────────────
	p.MessageID = cleanAngleBrackets(msg.Header.Get("Message-ID"))
	p.InReplyTo = cleanAngleBrackets(msg.Header.Get("In-Reply-To"))
	p.References = msg.Header.Get("References")
	p.ThreadID = msg.Header.Get("X-Thread-ID")
	if p.ThreadID == "" {
		p.ThreadID = msg.Header.Get("X-GM-THRID")
	}

	p.From = msg.Header.Get("From")
	p.FromAddress, p.FromName = parseAddress(p.From)

	p.To = msg.Header.Get("To")
	p.ToAddress, _ = parseAddress(p.To)

	p.Subject = decodeRFC2047(msg.Header.Get("Subject"))

	if dateStr := msg.Header.Get("Date"); dateStr != "" {
		if t, err := mail.ParseDate(dateStr); err == nil {
			p.Date = t
		}
	}
	if p.Date.IsZero() {
		p.Date = time.Now().UTC()
	}

	// Store a compact header snapshot (for debugging / threading fallback)
	var hb strings.Builder
	for k, vs := range msg.Header {
		for _, v := range vs {
			fmt.Fprintf(&hb, "%s: %s\n", k, v)
		}
	}
	p.RawHeaders = hb.String()

	// ── Body ───────────────────────────────────────────────────────────────────
	ct := msg.Header.Get("Content-Type")
	cte := msg.Header.Get("Content-Transfer-Encoding")

	if ct == "" {
		body, _ := io.ReadAll(msg.Body)
		p.TextBody = string(decodeTransfer(body, cte))
		return p, nil
	}

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		body, _ := io.ReadAll(msg.Body)
		p.TextBody = string(decodeTransfer(body, cte))
		return p, nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		parseMultipart(msg.Body, params["boundary"], p)
	} else {
		body, _ := io.ReadAll(msg.Body)
		decoded := decodeTransfer(body, cte)
		switch mediaType {
		case "text/html":
			p.HTMLBody = string(decoded)
		default:
			p.TextBody = string(decoded)
		}
	}

	return p, nil
}

// parseMultipart recursively walks multipart MIME parts.
func parseMultipart(r io.Reader, boundary string, p *providers.ParsedEmail) {
	mr := multipart.NewReader(r, boundary)
	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		ct := part.Header.Get("Content-Type")
		cte := part.Header.Get("Content-Transfer-Encoding")
		cd := part.Header.Get("Content-Disposition")

		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil {
			continue
		}

		// Nested multipart (e.g. multipart/alternative inside multipart/mixed)
		if strings.HasPrefix(mediaType, "multipart/") {
			parseMultipart(part, params["boundary"], p)
			continue
		}

		// Attachment detection: Content-Disposition = attachment, or non-text types
		if isAttachment(cd, mediaType) {
			data, err := io.ReadAll(part)
			if err != nil {
				continue
			}
			data = decodeTransfer(data, cte)
			filename := decodeRFC2047(params["name"])
			if filename == "" {
				filename = part.FileName()
			}
			if filename == "" {
				filename = "attachment"
			}
			p.Attachments = append(p.Attachments, providers.ParsedAttachment{
				Filename:    sanitiseFilename(filename),
				ContentType: mediaType,
				Data:        data,
				Size:        int64(len(data)),
			})
			continue
		}

		body, err := io.ReadAll(part)
		if err != nil {
			continue
		}
		decoded := decodeTransfer(body, cte)
		switch mediaType {
		case "text/html":
			if p.HTMLBody == "" {
				p.HTMLBody = string(decoded)
			}
		default:
			if p.TextBody == "" {
				p.TextBody = string(decoded)
			}
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func decodeTransfer(data []byte, encoding string) []byte {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		// Remove whitespace that some servers insert
		stripped := bytes.ReplaceAll(data, []byte("\r\n"), nil)
		stripped = bytes.ReplaceAll(stripped, []byte("\n"), nil)
		out := make([]byte, base64.StdEncoding.DecodedLen(len(stripped)))
		n, err := base64.StdEncoding.Decode(out, stripped)
		if err != nil {
			return data
		}
		return out[:n]
	case "quoted-printable":
		r := quotedprintable.NewReader(bytes.NewReader(data))
		out, err := io.ReadAll(r)
		if err != nil {
			return data
		}
		return out
	default:
		return data
	}
}

func decodeRFC2047(s string) string {
	dec := new(mime.WordDecoder)
	out, err := dec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return out
}

// parseAddress extracts the email address and display name from a header value.
func parseAddress(raw string) (addr, name string) {
	if raw == "" {
		return "", ""
	}
	a, err := mail.ParseAddress(raw)
	if err != nil {
		// might be a bare address
		return strings.TrimSpace(raw), ""
	}
	return a.Address, a.Name
}

// cleanAngleBrackets strips < > wrappers from Message-ID / In-Reply-To values.
func cleanAngleBrackets(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(s, ">")
	return strings.TrimSpace(s)
}

func isAttachment(cd, mediaType string) bool {
	if cd != "" {
		disp, _, _ := mime.ParseMediaType(cd)
		if strings.EqualFold(disp, "attachment") {
			return true
		}
	}
	switch {
	case strings.HasPrefix(mediaType, "text/plain"),
		strings.HasPrefix(mediaType, "text/html"):
		return false
	case strings.HasPrefix(mediaType, "application/"),
		strings.HasPrefix(mediaType, "image/"),
		strings.HasPrefix(mediaType, "audio/"),
		strings.HasPrefix(mediaType, "video/"):
		return true
	}
	return false
}

// sanitiseFilename removes path separators to prevent path traversal.
func sanitiseFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "..", "_")
	if name == "" {
		name = "attachment"
	}
	return name
}
