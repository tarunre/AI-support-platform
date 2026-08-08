package replyvalidator

import (
	"fmt"
	"strings"

	replyparser "github.com/ayush/supportiq/internal/ai/reply/parser"
)

// Validate checks the parsed AI reply for completeness and validity.
func Validate(r *replyparser.RawReplyResponse) error {
	if strings.TrimSpace(r.Reply) == "" {
		return fmt.Errorf("reply field is empty")
	}
	if r.Confidence < 0 || r.Confidence > 100 {
		return fmt.Errorf("confidence %d is outside valid range [0, 100]", r.Confidence)
	}
	return nil
}
