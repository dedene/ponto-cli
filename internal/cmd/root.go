package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alecthomas/kong"

	"github.com/dedene/ponto-cli/internal/config"
	pontoCtx "github.com/dedene/ponto-cli/internal/ctx"
	"github.com/dedene/ponto-cli/internal/output"
)

// RootFlags contains global CLI flags.
type RootFlags struct {
	Profile        string        `help:"Profile name" default:"${profile}" env:"PONTO_PROFILE"`
	Sandbox        bool          `help:"Use sandbox profile" default:"false"`
	EnableCommands string        `help:"Comma-separated list of enabled commands" default:"${enabled_commands}" env:"PONTO_ENABLE_COMMANDS"`
	JSON           bool          `help:"Output JSON"`
	CSV            bool          `help:"Output CSV"`
	Plain          bool          `help:"Output TSV (stable for scripting)"`
	Verbose        int           `short:"v" type:"counter" help:"Verbosity (-v, -vv)"`
	Timeout        time.Duration `help:"Request timeout" default:"30s"`
	NoRetry        bool          `help:"Disable retry on errors"`
}

// CLI is the root command structure.
type CLI struct {
	RootFlags `embed:""`

	Version    kong.VersionFlag `help:"Print version and exit"`
	VersionCmd VersionCmd       `cmd:"" name:"version" help:"Print version"`

	Auth         AuthCmd         `cmd:"" help:"Authentication commands"`
	Accounts     AccountsCmd     `cmd:"" help:"Bank accounts"`
	Transactions TransactionsCmd `cmd:"" help:"Account transactions"`
	Sync         SyncCmd         `cmd:"" help:"Synchronization"`
	Organization OrganizationCmd `cmd:"" help:"Organization info"`

	PendingTransactions   PendingTransactionsCmd   `cmd:"" name:"pending-transactions" help:"Pending transactions"`
	FinancialInstitutions FinancialInstitutionsCmd `cmd:"" name:"financial-institutions" help:"Financial institutions"`

	Completion CompletionCmd `cmd:"" help:"Generate shell completions"`
	Config     ConfigCmd     `cmd:"" help:"Configuration"`
}

type exitPanic struct{ code int }

// Execute runs the CLI with the given arguments.
func Execute(args []string) (err error) {
	parser, cli, err := newParser()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil

					return
				}

				err = &ExitError{Code: ep.code, Err: errors.New("exited")}

				return
			}

			panic(r)
		}
	}()

	kctx, err := parser.Parse(args)
	if err != nil {
		// Show help when no command provided
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) && parseErr.Context.Command() == "" {
			_ = parseErr.Context.PrintUsage(true)

			return nil
		}

		fmt.Fprintln(os.Stderr, err)

		return wrapParseError(err)
	}

	if err = enforceEnabledCommands(kctx, cli.EnableCommands); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return err
	}

	// Set up logging
	logLevel := slog.LevelWarn
	if cli.Verbose >= 2 {
		logLevel = slog.LevelDebug
	} else if cli.Verbose == 1 {
		logLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	// Determine output mode
	mode := output.ModeTable

	switch {
	case cli.JSON:
		mode = output.ModeJSON
	case cli.CSV:
		mode = output.ModeCSV
	case cli.Plain:
		mode = output.ModePlain
	}

	// Resolve profile
	profile := cli.Profile
	if cli.Sandbox {
		profile = "sandbox"
	}

	if profile == "" {
		profile = "default"
	}

	// Build context
	ctx := context.Background()
	ctx = output.WithMode(ctx, mode)
	ctx = pontoCtx.WithProfile(ctx, profile)
	ctx = pontoCtx.WithTimeout(ctx, cli.Timeout)
	ctx = pontoCtx.WithNoRetry(ctx, cli.NoRetry)

	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&cli.RootFlags)

	if err = kctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return err
	}

	return nil
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}

	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: 2, Err: parseErr}
	}

	return err
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func newParser() (*kong.Kong, *CLI, error) {
	vars := kong.Vars{
		"profile":          envOr("PONTO_PROFILE", "default"),
		"enabled_commands": envOr("PONTO_ENABLE_COMMANDS", ""),
		"version":          VersionString(),
	}

	cli := &CLI{}

	parser, err := kong.New(
		cli,
		kong.Name("ponto"),
		kong.Description(helpDescription()),
		kong.ConfigureHelp(helpOptions()),
		kong.Help(helpPrinter),
		kong.Vars(vars),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
	)
	if err != nil {
		return nil, nil, err
	}

	return parser, cli, nil
}

func helpDescription() string {
	desc := "Ponto CLI - Banking API for organizations"

	configPath, err := config.ConfigPath()
	if err != nil {
		return desc
	}

	return fmt.Sprintf("%s\n\nConfig: %s", desc, configPath)
}

// ProfileFrom retrieves the profile from the context (re-exported for convenience).
func ProfileFrom(ctx context.Context) string {
	return pontoCtx.ProfileFrom(ctx)
}
