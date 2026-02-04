package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/output"
)

// PendingTransactionsCmd is the parent command for pending transactions.
type PendingTransactionsCmd struct {
	List PendingTransactionsListCmd `cmd:"" help:"List pending transactions"`
}

// PendingTransactionsListCmd lists pending transactions.
type PendingTransactionsListCmd struct {
	AccountID string `help:"Account ID (default: from config or auto-detect)" name:"account-id"`
}

func (c *PendingTransactionsListCmd) Run(ctx context.Context) error {
	accountID, err := ResolveAccountID(ctx, c.AccountID)
	if err != nil {
		return err
	}

	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	transactions, err := client.ListPendingTransactions(ctx, accountID)
	if err != nil {
		return fmt.Errorf("list pending transactions: %w", err)
	}

	mode := output.ModeFrom(ctx)

	return output.PendingTransactions(mode, transactions)
}
