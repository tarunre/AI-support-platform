// Package integrations provides the provider registry and event worker.
package integrations

import (
	"fmt"

	"github.com/ayush/supportiq/internal/integrations/provider"
	"github.com/ayush/supportiq/internal/integrations/providers/discord"
	"github.com/ayush/supportiq/internal/integrations/providers/gcal"
	"github.com/ayush/supportiq/internal/integrations/providers/github"
	"github.com/ayush/supportiq/internal/integrations/providers/hubspot"
	"github.com/ayush/supportiq/internal/integrations/providers/jira"
	"github.com/ayush/supportiq/internal/integrations/providers/linear"
	"github.com/ayush/supportiq/internal/integrations/providers/salesforce"
	"github.com/ayush/supportiq/internal/integrations/providers/slack"
	"github.com/ayush/supportiq/internal/integrations/providers/teams"
	"github.com/ayush/supportiq/internal/integrations/providers/webhook"
)

// Registry creates configured provider instances by type ID.
type Registry struct{}

// NewRegistry returns a new provider Registry.
func NewRegistry() *Registry { return &Registry{} }

// Build instantiates and configures the provider identified by providerType.
func (r *Registry) Build(providerType string, cfg map[string]interface{}) (provider.Provider, error) {
	var p provider.Provider
	switch providerType {
	case "slack":
		p = &slack.Provider{}
	case "teams":
		p = &teams.Provider{}
	case "discord":
		p = &discord.Provider{}
	case "jira":
		p = &jira.Provider{}
	case "linear":
		p = &linear.Provider{}
	case "github":
		p = &github.Provider{}
	case "webhook":
		p = &webhook.Provider{}
	case "salesforce":
		p = &salesforce.Provider{}
	case "hubspot":
		p = &hubspot.Provider{}
	case "gcal":
		p = &gcal.Provider{}
	default:
		return nil, fmt.Errorf("registry: unknown provider %q", providerType)
	}
	if err := p.Configure(cfg); err != nil {
		return nil, fmt.Errorf("registry: configure %s: %w", providerType, err)
	}
	return p, nil
}
