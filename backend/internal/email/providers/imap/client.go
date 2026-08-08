// Package imap implements providers.Receiver using go-imap v1.
package imap

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	goImap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"

	emailparser "github.com/ayush/supportiq/internal/email/parser"
	"github.com/ayush/supportiq/internal/email/providers"
)

// Client polls an IMAP mailbox and returns unread messages.
type Client struct {
	host      string
	port      int
	user      string
	pass      string
	useTLS    bool
	sinceTime *time.Time // if set, only fetch emails received at or after this time
}

// New creates an IMAP client.
func New(host string, port int, username, password string, useTLS bool) *Client {
	return &Client{host: host, port: port, user: username, pass: password, useTLS: useTLS}
}

// SetSince restricts IMAP SEARCH to emails received at or after the given time.
// Call this with account.LastSyncAt to avoid re-processing old inbox emails.
func (c *Client) SetSince(t time.Time) { c.sinceTime = &t }

// FetchUnread connects, fetches unseen messages received since sinceTime
// (or since today if sinceTime is nil), and returns the parsed results.
func (c *Client) FetchUnread(_ context.Context) ([]providers.ParsedEmail, error) {
	cl, err := c.dial()
	if err != nil {
		return nil, err
	}
	defer cl.Logout() //nolint:errcheck

	if _, err = cl.Select("INBOX", false); err != nil {
		return nil, fmt.Errorf("imap: SELECT INBOX: %w", err)
	}

	// Search for unseen messages received since sinceTime (ignores old inbox emails)
	criteria := goImap.NewSearchCriteria()
	criteria.WithoutFlags = []string{goImap.SeenFlag}
	if c.sinceTime != nil {
		criteria.Since = *c.sinceTime
	} else {
		criteria.Since = time.Now().UTC().Truncate(24 * time.Hour)
	}
	uids, err := cl.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("imap: UID SEARCH: %w", err)
	}
	if len(uids) == 0 {
		return nil, nil
	}

	seqset := new(goImap.SeqSet)
	seqset.AddNum(uids...)

	// BODY.PEEK[] fetches the full RFC 2822 message without marking as seen
	var bodySec goImap.BodySectionName
	bodySec.Peek = true

	items := []goImap.FetchItem{
		goImap.FetchUid,
		goImap.FetchFlags,
		bodySec.FetchItem(),
	}

	msgCh := make(chan *goImap.Message, 20)
	fetchErr := make(chan error, 1)
	go func() {
		fetchErr <- cl.UidFetch(seqset, items, msgCh)
	}()

	var results []providers.ParsedEmail
	for imapMsg := range msgCh {
		// Grab the raw body bytes (iterate the body map — key matching can vary)
		var raw []byte
		for _, literal := range imapMsg.Body {
			b, err := io.ReadAll(literal)
			if err == nil && len(b) > 0 {
				raw = b
				break
			}
		}
		if len(raw) == 0 {
			continue
		}

		parsed, err := emailparser.Parse(raw)
		if err != nil {
			continue
		}
		parsed.UID = imapMsg.Uid
		results = append(results, *parsed)
	}

	if err := <-fetchErr; err != nil {
		return nil, fmt.Errorf("imap: UidFetch: %w", err)
	}

	return results, nil
}

// MarkSeen adds the \Seen flag to the message with the given UID.
func (c *Client) MarkSeen(_ context.Context, uid uint32) error {
	cl, err := c.dial()
	if err != nil {
		return err
	}
	defer cl.Logout() //nolint:errcheck

	if _, err = cl.Select("INBOX", false); err != nil {
		return fmt.Errorf("imap: SELECT INBOX: %w", err)
	}

	seqset := new(goImap.SeqSet)
	seqset.AddNum(uid)

	item := goImap.FormatFlagsOp(goImap.AddFlags, true)
	return cl.UidStore(seqset, item, []interface{}{goImap.SeenFlag}, nil)
}

// TestConnection authenticates and immediately logs out to verify credentials.
func (c *Client) TestConnection(_ context.Context) error {
	cl, err := c.dial()
	if err != nil {
		return err
	}
	return cl.Logout()
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (c *Client) dial() (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	var cl *client.Client
	var err error
	if c.useTLS {
		tlsCfg := &tls.Config{ServerName: c.host}
		cl, err = client.DialTLS(addr, tlsCfg)
	} else {
		conn, dialErr := net.DialTimeout("tcp", addr, 20*time.Second)
		if dialErr != nil {
			return nil, fmt.Errorf("imap: dial %s: %w", addr, dialErr)
		}
		cl, err = client.New(conn)
	}
	if err != nil {
		return nil, fmt.Errorf("imap: connect %s: %w", addr, err)
	}

	if err = cl.Login(c.user, c.pass); err != nil {
		_ = cl.Logout()
		return nil, fmt.Errorf("imap: login: %w", err)
	}
	return cl, nil
}
