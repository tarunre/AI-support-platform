package replyprompt

import (
	"fmt"
	"strings"

	replyprovider "github.com/ayush/supportiq/internal/ai/reply/provider"
)

// CurrentVersion is the prompt template version. Bump when the prompt changes.
const CurrentVersion = "v1"
const PortalPromptVersion = "portal-v1"

// BuildReplyPrompt constructs the RAG-enhanced Gemini prompt for reply generation.
// It embeds relevant knowledge base documents so the AI answers only from
// verified company policies — never from its own pre-trained knowledge.
// If req.CustomPrompt is set it is returned as-is, bypassing the standard template.
func BuildReplyPrompt(req replyprovider.ReplyRequest) string {
	if req.CustomPrompt != "" {
		return req.CustomPrompt
	}
	var sb strings.Builder

	sb.WriteString("You are a professional customer support AI assistant. " +
		"Your task is to draft a reply to a customer support ticket.\n\n")
	sb.WriteString("Return ONLY a valid JSON object. No markdown. No code blocks. " +
		"No explanations. Just the raw JSON.\n\n")

	sb.WriteString("--- TICKET DETAILS ---\n")
	sb.WriteString(fmt.Sprintf("Subject: %s\n", req.Subject))
	sb.WriteString(fmt.Sprintf("Description: %s\n", req.Description))
	sb.WriteString(fmt.Sprintf("Category: %s\n", req.Category))
	sb.WriteString(fmt.Sprintf("Priority: %s\n", req.Priority))
	sb.WriteString(fmt.Sprintf("Customer Sentiment: %s\n\n", req.Sentiment))

	// Inject real order status if found — AI must use this, not guess
	if req.OrderContext != "" {
		sb.WriteString("--- REAL ORDER DATA (use this EXACTLY in your reply) ---\n")
		sb.WriteString(req.OrderContext)
		sb.WriteString("--- END ORDER DATA ---\n\n")
	}

	if len(req.Documents) > 0 {
		sb.WriteString("--- RELEVANT KNOWLEDGE BASE DOCUMENTS ---\n\n")
		for i, doc := range req.Documents {
			sb.WriteString(fmt.Sprintf("[Document %d] %s (%s)\n", i+1, doc.Title, doc.Category))
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
		sb.WriteString("--- END OF KNOWLEDGE BASE DOCUMENTS ---\n\n")
		sb.WriteString(`Instructions:
- Write a professional, empathetic customer support reply.
- Be concise and solution-focused.
- Prefer information from the knowledge base documents above when available.
- If documents do not fully address the issue, use your general knowledge to help.
- Do not include subject lines or greetings like "Dear Customer" — go straight to the response body.
`)
	} else {
		sb.WriteString("--- NO KNOWLEDGE BASE DOCUMENTS AVAILABLE ---\n\n")
		sb.WriteString("Instructions:\n- Write a professional, empathetic customer support reply using your general knowledge.\n- Be concise and solution-focused.\n- Acknowledge the customer's concern and provide helpful, actionable guidance.\n- Do not include subject lines or greetings like \"Dear Customer\" — go straight to the response body.\n")
	}

	sb.WriteString(`
Required JSON structure (use EXACTLY these field names):
{
  "reply": "<the complete ready-to-send reply text>",
  "confidence": <integer between 0 and 100>
}

Rules:
- Output ONLY the JSON object. Nothing before or after it.
- confidence must be an integer reflecting how well the response covers the issue (100 = perfect).
- reply must be a complete, professional customer support response.`)

	return sb.String()
}

// BuildPortalReplyPrompt builds a focused, concise prompt for real-time portal chat replies.
// The reply must directly answer the customer's latest message in 1-2 sentences max.
func BuildPortalReplyPrompt(subject, latestMessage, category, sentiment string, docs []replyprovider.RelevantDocument) string {
	var sb strings.Builder

	sb.WriteString("You are a friendly, helpful customer support agent chatting in real-time with a customer.\n\n")
	sb.WriteString("Return ONLY a valid JSON object. No markdown. No code blocks. Just raw JSON.\n\n")

	sb.WriteString(fmt.Sprintf("Ticket subject: %s\n", subject))
	if category != "" {
		sb.WriteString(fmt.Sprintf("Category: %s | Sentiment: %s\n", category, sentiment))
	}

	if len(docs) > 0 {
		sb.WriteString("\nRelevant knowledge:\n")
		for _, d := range docs {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", d.Title, d.Content))
		}
	}

	sb.WriteString(fmt.Sprintf("\nCustomer's latest message: \"%s\"\n\n", latestMessage))

	sb.WriteString(`STRICT RULES:
- Answer ONLY the customer's latest message above. Do not re-summarize the ticket.
- Keep reply to 1-2 short sentences MAXIMUM. Be direct and specific.
- Conversational tone, not formal. No "Dear Customer", no lengthy greetings.
- If you need more info, ask ONE specific question only.

Required JSON (EXACTLY these fields):
{
  "reply": "<1-2 sentence direct answer>",
  "confidence": <0-100>
}`)

	return sb.String()
}
