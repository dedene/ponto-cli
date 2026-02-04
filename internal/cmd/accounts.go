package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/output"
)

// AccountsCmd is the parent command for accounts.
type AccountsCmd struct {
	List AccountsListCmd `cmd:"" help:"List all accounts"`
	Get  AccountsGetCmd  `cmd:"" help:"Get account details"`
	Sync AccountsSyncCmd `cmd:"" help:"Trigger account synchronization"`
}

// AccountsListCmd lists accounts.
type AccountsListCmd struct {
	Product string `help:"Filter by product type"`
}

func (c *AccountsListCmd) Run(ctx context.Context) error {
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		return fmt.Errorf("list accounts: %w", err)
	}

	mode := output.ModeFrom(ctx)

	if c.Product != "" {
		filtered := make([]api.Account, 0)

		for _, a := range accounts {
			if a.Product == c.Product {
				filtered = append(filtered, a)
			}
		}

		accounts = filtered
	}

	return output.Accounts(mode, accounts)
}

// AccountsGetCmd gets account details.
type AccountsGetCmd struct {
	ID string `arg:"" help:"Account ID (use - for stdin)"`
}

func (c *AccountsGetCmd) Run(ctx context.Context) error {
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
		account, err := client.GetAccount(ctx, id)
		if err != nil {
			return fmt.Errorf("get account %s: %w", id, err)
		}

		if err := output.Account(mode, account); err != nil {
			return err
		}
	}

	return nil
}

// AccountsSyncCmd triggers synchronization.
type AccountsSyncCmd struct {
	ID   string `arg:"" help:"Account ID"`
	Wait bool   `help:"Wait for sync to complete"`
}

func (c *AccountsSyncCmd) Run(ctx context.Context) error {
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	sync, err := client.CreateSync(ctx, c.ID, "accountTransactions")
	if err != nil {
		return fmt.Errorf("create sync: %w", err)
	}

	mode := output.ModeFrom(ctx)

	if c.Wait {
		sync, err = client.WaitForSync(ctx, sync.ID)
		if err != nil {
			return fmt.Errorf("wait for sync: %w", err)
		}
	}

	return output.Sync(mode, sync)
}
