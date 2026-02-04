package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/output"
)

// SyncCmd is the parent command for synchronization.
type SyncCmd struct {
	Create SyncCreateCmd `cmd:"" help:"Create a new synchronization"`
	Get    SyncGetCmd    `cmd:"" help:"Get synchronization status"`
	List   SyncListCmd   `cmd:"" help:"List synchronizations"`
}

// SyncCreateCmd creates a sync.
type SyncCreateCmd struct {
	AccountID string `help:"Account ID (default: from config or auto-detect)" name:"account-id"`
	Subtype   string `required:"" help:"Sync subtype (accountDetails, accountTransactions)" enum:"accountDetails,accountTransactions"`
	Wait      bool   `help:"Wait for sync to complete"`
}

func (c *SyncCreateCmd) Run(ctx context.Context) error {
	accountID, err := ResolveAccountID(ctx, c.AccountID)
	if err != nil {
		return err
	}

	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	sync, err := client.CreateSync(ctx, accountID, c.Subtype)
	if err != nil {
		return fmt.Errorf("create sync: %w", err)
	}

	if c.Wait {
		sync, err = client.WaitForSync(ctx, sync.ID)
		if err != nil {
			return fmt.Errorf("wait for sync: %w", err)
		}
	}

	mode := output.ModeFrom(ctx)

	return output.Sync(mode, sync)
}

// SyncGetCmd gets sync status.
type SyncGetCmd struct {
	ID string `arg:"" help:"Synchronization ID"`
}

func (c *SyncGetCmd) Run(ctx context.Context) error {
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	sync, err := client.GetSync(ctx, c.ID)
	if err != nil {
		return fmt.Errorf("get sync: %w", err)
	}

	mode := output.ModeFrom(ctx)

	return output.Sync(mode, sync)
}

// SyncListCmd lists syncs.
type SyncListCmd struct {
	AccountID string `help:"Account ID (default: from config or auto-detect)" name:"account-id"`
	Limit     int    `help:"Maximum number of syncs" default:"10"`
}

func (c *SyncListCmd) Run(ctx context.Context) error {
	accountID, err := ResolveAccountID(ctx, c.AccountID)
	if err != nil {
		return err
	}

	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	syncs, err := client.ListSyncs(ctx, accountID, c.Limit)
	if err != nil {
		return fmt.Errorf("list syncs: %w", err)
	}

	mode := output.ModeFrom(ctx)

	return output.Syncs(mode, syncs)
}
