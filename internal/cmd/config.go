package cmd

import (
	"context"
	"fmt"

	"github.com/dedene/ponto-cli/internal/config"
	pontoCtx "github.com/dedene/ponto-cli/internal/ctx"
	"github.com/dedene/ponto-cli/internal/output"
)

// ConfigCmd is the parent command for configuration.
type ConfigCmd struct {
	Get ConfigGetCmd `cmd:"" help:"Get configuration value"`
	Set ConfigSetCmd `cmd:"" help:"Set configuration value"`
}

// ConfigGetCmd gets a configuration value.
type ConfigGetCmd struct {
	Key string `arg:"" help:"Config key (account-id)"`
}

func (c *ConfigGetCmd) Run(ctx context.Context) error {
	profile := pontoCtx.ProfileFrom(ctx)

	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	p := cfg.Profiles[profile]

	switch c.Key {
	case "account-id":
		if p.AccountID == "" {
			return fmt.Errorf("account-id not set for profile %q", profile)
		}

		fmt.Println(p.AccountID)
	default:
		return fmt.Errorf("unknown config key: %s", c.Key)
	}

	return nil
}

// ConfigSetCmd sets a configuration value.
type ConfigSetCmd struct {
	Key   string `arg:"" help:"Config key (account-id)"`
	Value string `arg:"" help:"Value to set"`
}

func (c *ConfigSetCmd) Run(ctx context.Context) error {
	profile := pontoCtx.ProfileFrom(ctx)

	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]config.Profile)
	}

	p := cfg.Profiles[profile]

	switch c.Key {
	case "account-id":
		p.AccountID = c.Value
	default:
		return fmt.Errorf("unknown config key: %s", c.Key)
	}

	cfg.Profiles[profile] = p

	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	mode := output.ModeFrom(ctx)
	if mode == output.ModeTable {
		fmt.Printf("Set %s=%s for profile %q\n", c.Key, c.Value, profile)
	}

	return nil
}
