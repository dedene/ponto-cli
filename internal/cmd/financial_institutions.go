package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/output"
)

// FinancialInstitutionsCmd is the parent command for financial institutions.
type FinancialInstitutionsCmd struct {
	List FinancialInstitutionsListCmd `cmd:"" help:"List financial institutions"`
	Get  FinancialInstitutionsGetCmd  `cmd:"" help:"Get financial institution details"`
}

// FinancialInstitutionsListCmd lists financial institutions.
type FinancialInstitutionsListCmd struct{}

func (c *FinancialInstitutionsListCmd) Run(ctx context.Context) error {
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	institutions, err := client.ListFinancialInstitutions(ctx)
	if err != nil {
		return fmt.Errorf("list financial institutions: %w", err)
	}

	mode := output.ModeFrom(ctx)

	return output.FinancialInstitutions(mode, institutions)
}

// FinancialInstitutionsGetCmd gets financial institution details.
type FinancialInstitutionsGetCmd struct {
	ID string `arg:"" help:"Financial institution ID"`
}

func (c *FinancialInstitutionsGetCmd) Run(ctx context.Context) error {
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	institution, err := client.GetFinancialInstitution(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("get financial institution: %w", err)
	}

	mode := output.ModeFrom(ctx)

	return output.FinancialInstitution(mode, institution)
}
