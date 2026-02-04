package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/output"
)

// TransactionsCmd is the parent command for transactions.
type TransactionsCmd struct {
	List   TransactionsListCmd   `cmd:"" help:"List transactions"`
	Get    TransactionsGetCmd    `cmd:"" help:"Get transaction details"`
	Export TransactionsExportCmd `cmd:"" help:"Export transactions"`
}

// TransactionsListCmd lists transactions.
type TransactionsListCmd struct {
	AccountID string `help:"Account ID (default: from config or auto-detect)" name:"account-id"`
	Since     string `help:"Start date (ISO 8601 or relative like -30d)"`
	Until     string `help:"End date (ISO 8601 or relative like -1d)"`
	Limit     int    `help:"Maximum number of transactions" default:"100"`
	Type      string `help:"Filter by type: income, expense, or all" enum:"income,expense,all" default:"all"`
}

func (c *TransactionsListCmd) Run(ctx context.Context) error {
	accountID, err := ResolveAccountID(ctx, c.AccountID)
	if err != nil {
		return err
	}

	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	opts := api.TransactionListOptions{
		Since: c.Since,
		Until: c.Until,
		Limit: c.Limit,
	}

	transactions, err := client.ListTransactions(ctx, accountID, opts)
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}

	transactions = filterTransactionsByType(transactions, c.Type)
	mode := output.ModeFrom(ctx)

	return output.Transactions(mode, transactions)
}

// TransactionsGetCmd gets transaction details.
type TransactionsGetCmd struct {
	AccountID string `help:"Account ID (default: from config or auto-detect)" name:"account-id"`
	ID        string `arg:"" help:"Transaction ID (use - for stdin)"`
}

func (c *TransactionsGetCmd) Run(ctx context.Context) error {
	accountID, err := ResolveAccountID(ctx, c.AccountID)
	if err != nil {
		return err
	}

	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	mode := output.ModeFrom(ctx)

	// Handle stdin batching
	ids, err := ReadStdinIDs(c.ID)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	if ids == nil {
		ids = []string{c.ID}
	}

	for _, id := range ids {
		transaction, err := client.GetTransaction(ctx, accountID, id)
		if err != nil {
			return fmt.Errorf("get transaction %s: %w", id, err)
		}

		if err := output.Transaction(mode, transaction); err != nil {
			return err
		}
	}

	return nil
}

// TransactionsExportCmd exports transactions.
type TransactionsExportCmd struct {
	AccountID string `help:"Account ID (default: from config or auto-detect)" name:"account-id"`
	Since     string `help:"Start date (ISO 8601 or relative like -30d)"`
	Until     string `help:"End date (ISO 8601 or relative like -1d)"`
	Format    string `help:"Output format (csv, json)" default:"csv" enum:"csv,json"`
	Type      string `help:"Filter by type: income, expense, or all" enum:"income,expense,all" default:"all"`
}

func (c *TransactionsExportCmd) Run(ctx context.Context) error {
	accountID, err := ResolveAccountID(ctx, c.AccountID)
	if err != nil {
		return err
	}

	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	opts := api.TransactionListOptions{
		Since: c.Since,
		Until: c.Until,
		Limit: 0, // no limit for export
	}

	transactions, err := client.ListTransactions(ctx, accountID, opts)
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}

	transactions = filterTransactionsByType(transactions, c.Type)

	mode := output.ModeCSV
	if c.Format == "json" {
		mode = output.ModeJSON
	}

	return output.Transactions(mode, transactions)
}

// filterTransactionsByType filters transactions by income/expense type.
func filterTransactionsByType(txs []api.Transaction, typ string) []api.Transaction {
	if typ == "all" {
		return txs
	}
	filtered := make([]api.Transaction, 0, len(txs))
	for _, tx := range txs {
		if typ == "income" && tx.Amount > 0 {
			filtered = append(filtered, tx)
		} else if typ == "expense" && tx.Amount < 0 {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}
