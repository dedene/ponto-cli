package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/api"
	"github.com/dedene/ponto-cli/internal/config"
	pontoCtx "github.com/dedene/ponto-cli/internal/ctx"
)

// ResolveAccountID resolves an account ID from flag, config, or auto-detection.
// Priority: flag > config > single account auto-detect.
func ResolveAccountID(ctx context.Context, flagValue string) (string, error) {
	// 1. Flag takes precedence
	if flagValue != "" {
		return flagValue, nil
	}

	// 2. Check config
	profile := pontoCtx.ProfileFrom(ctx)

	cfg, err := config.ReadConfig()
	if err == nil {
		if p, ok := cfg.Profiles[profile]; ok && p.AccountID != "" {
			return p.AccountID, nil
		}
	}

	// 3. Auto-detect single account
	client, err := api.NewClientFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing --account-id (set via flag, config, or run 'ponto config set account-id <id>')")
	}

	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		return "", fmt.Errorf("missing --account-id (set via flag or run 'ponto config set account-id <id>')")
	}

	if len(accounts) == 1 {
		return accounts[0].ID, nil
	}

	if len(accounts) == 0 {
		return "", fmt.Errorf("no accounts found; link an account first")
	}

	return "", fmt.Errorf("multiple accounts found; specify --account-id or run 'ponto config set account-id <id>'")
}
