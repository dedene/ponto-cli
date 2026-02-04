package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/dedene/ponto-cli/internal/auth"
	pontoCtx "github.com/dedene/ponto-cli/internal/ctx"
	"github.com/dedene/ponto-cli/internal/output"
)

// AuthCmd is the parent command for authentication.
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Store credentials in keyring"`
	Logout AuthLogoutCmd `cmd:"" help:"Remove credentials from keyring"`
	Status AuthStatusCmd `cmd:"" help:"Show authentication status"`
}

// AuthLoginCmd stores credentials.
type AuthLoginCmd struct{}

func (c *AuthLoginCmd) Run(ctx context.Context) error {
	profile := pontoCtx.ProfileFrom(ctx)

	fmt.Printf("Logging in to profile: %s\n", profile)
	fmt.Println("Enter your Ponto API credentials (from the Ponto dashboard):")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Client ID: ")

	clientID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read client ID: %w", err)
	}

	clientID = strings.TrimSpace(clientID)

	fmt.Print("Client Secret: ")

	secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("read client secret: %w", err)
	}

	fmt.Println() // newline after hidden input

	clientSecret := strings.TrimSpace(string(secretBytes))

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("client ID and secret are required")
	}

	// Test the credentials by fetching a token
	fmt.Print("Verifying credentials... ")

	token, err := auth.GetAccessToken(ctx, clientID, clientSecret)
	if err != nil {
		fmt.Println("failed")

		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("ok")

	// Store in keyring
	store, err := auth.OpenKeyring()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if err := store.SetCredentials(profile, clientID, clientSecret); err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	fmt.Printf("Credentials stored for profile %q\n", profile)
	fmt.Printf("Token scopes: %s\n", token.Scope)

	return nil
}

// AuthLogoutCmd removes credentials.
type AuthLogoutCmd struct{}

func (c *AuthLogoutCmd) Run(ctx context.Context) error {
	profile := pontoCtx.ProfileFrom(ctx)

	store, err := auth.OpenKeyring()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if err := store.DeleteCredentials(profile); err != nil {
		return fmt.Errorf("delete credentials: %w", err)
	}

	fmt.Printf("Credentials removed for profile %q\n", profile)

	return nil
}

// AuthStatusCmd shows authentication status.
type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run(ctx context.Context) error {
	profile := pontoCtx.ProfileFrom(ctx)
	mode := output.ModeFrom(ctx)

	store, err := auth.OpenKeyring()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	clientID, clientSecret, err := store.GetCredentials(profile)
	if err != nil {
		if mode == output.ModeJSON {
			fmt.Println(`{"authenticated": false}`)

			return nil
		}

		fmt.Printf("Profile: %s\n", profile)
		fmt.Println("Status: not authenticated")
		fmt.Println("Run 'ponto auth login' to authenticate.")

		return nil
	}

	// Mask the credentials
	maskedID := maskString(clientID)
	maskedSecret := maskString(clientSecret)

	if mode == output.ModeJSON {
		fmt.Printf(`{"authenticated": true, "profile": %q, "client_id": %q}`, profile, maskedID)
		fmt.Println()

		return nil
	}

	fmt.Printf("Profile: %s\n", profile)
	fmt.Println("Status: authenticated")
	fmt.Printf("Client ID: %s\n", maskedID)
	fmt.Printf("Client Secret: %s\n", maskedSecret)

	return nil
}

func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}

	return s[:4] + "****" + s[len(s)-4:]
}
