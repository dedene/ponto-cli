package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/output"
)

// OrganizationCmd is the parent command for organization.
type OrganizationCmd struct {
	Show OrganizationShowCmd `cmd:"" help:"Show organization info"`
}

// OrganizationShowCmd shows org info.
type OrganizationShowCmd struct{}

func (c *OrganizationShowCmd) Run(ctx context.Context) error {
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	org, err := client.GetOrganization(ctx)
	if err != nil {
		return fmt.Errorf("get organization: %w", err)
	}

	mode := output.ModeFrom(ctx)

	return output.Organization(mode, org)
}
