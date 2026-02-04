package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/dedene/ponto-cli/internal/auth"
	pontoCtx "github.com/dedene/ponto-cli/internal/ctx"
)

const (
	baseURL   = "https://api.myponto.com"
	userAgent = "ponto-cli"
)

// version is set at build time via ldflags.
var version = "dev"

// Client is the Ponto API client.
type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	timeout      time.Duration
}

// NewClientFromContext creates a client from context.
func NewClientFromContext(ctx context.Context) (*Client, error) {
	profile := pontoCtx.ProfileFrom(ctx)
	timeout := pontoCtx.TimeoutFrom(ctx)
	noRetry := pontoCtx.NoRetryFrom(ctx)

	store, err := auth.OpenKeyring()
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	clientID, clientSecret, err := store.GetCredentials(profile)
	if err != nil {
		return nil, fmt.Errorf("get credentials for profile %q: %w\nRun 'ponto auth login' to authenticate", profile, err)
	}

	transport := NewRetryTransport(http.DefaultTransport, noRetry)

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
		clientID:     clientID,
		clientSecret: clientSecret,
		timeout:      timeout,
	}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	token, err := auth.GetAccessToken(ctx, c.clientID, c.clientSecret)
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	u := baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s (%s/%s)", userAgent, version, runtime.GOOS, runtime.GOARCH))

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	return resp, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func decodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}

	var wrapper DataWrapper[Resource]
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Convert Resource to T
	b, err := json.Marshal(wrapper.Data.Attributes)
	if err != nil {
		return nil, fmt.Errorf("marshal attributes: %w", err)
	}

	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("unmarshal to type: %w", err)
	}

	// Set ID if the result has an ID field
	if r, ok := any(&result).(*Account); ok {
		r.ID = wrapper.Data.ID
	} else if r, ok := any(&result).(*Transaction); ok {
		r.ID = wrapper.Data.ID
	} else if r, ok := any(&result).(*Synchronization); ok {
		r.ID = wrapper.Data.ID
	} else if r, ok := any(&result).(*FinancialInstitution); ok {
		r.ID = wrapper.Data.ID
	} else if r, ok := any(&result).(*Organization); ok {
		r.ID = wrapper.Data.ID
	}

	return &result, nil
}

func decodeListResponse[T any](resp *http.Response, setID func(*T, string)) ([]T, error) {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}

	var wrapper ListWrapper[Resource]
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	results := make([]T, 0, len(wrapper.Data))

	for _, res := range wrapper.Data {
		b, err := json.Marshal(res.Attributes)
		if err != nil {
			return nil, fmt.Errorf("marshal attributes: %w", err)
		}

		var item T
		if err := json.Unmarshal(b, &item); err != nil {
			return nil, fmt.Errorf("unmarshal to type: %w", err)
		}

		if setID != nil {
			setID(&item, res.ID)
		}

		results = append(results, item)
	}

	return results, nil
}

func parseAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	apiErr := &APIError{StatusCode: resp.StatusCode}

	// Try to parse error response
	var errResp struct {
		Errors []struct {
			Code   string `json:"code"`
			Detail string `json:"detail"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 {
		apiErr.Code = errResp.Errors[0].Code
		apiErr.Message = errResp.Errors[0].Detail
	} else {
		apiErr.Message = string(body)
	}

	return apiErr
}

// ListAccounts returns all accounts.
func (c *Client) ListAccounts(ctx context.Context) ([]Account, error) {
	resp, err := c.get(ctx, "/accounts")
	if err != nil {
		return nil, err
	}

	return decodeListResponse(resp, func(a *Account, id string) { a.ID = id })
}

// GetAccount returns a single account.
func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	resp, err := c.get(ctx, "/accounts/"+id)
	if err != nil {
		return nil, err
	}

	return decodeResponse[Account](resp)
}

const maxPageSize = 100

// ListTransactions returns transactions for an account.
func (c *Client) ListTransactions(ctx context.Context, accountID string, opts TransactionListOptions) ([]Transaction, error) {
	basePath := fmt.Sprintf("/accounts/%s/transactions", accountID)

	params := url.Values{}

	// Cap page size at API max
	pageSize := maxPageSize
	if opts.Limit > 0 && opts.Limit < maxPageSize {
		pageSize = opts.Limit
	}
	params.Set("limit", fmt.Sprintf("%d", pageSize))

	if opts.Since != "" {
		since, err := parseDate(opts.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid since date: %w", err)
		}

		params.Set("filter[valueDate][gte]", since)
	}

	if opts.Until != "" {
		until, err := parseDate(opts.Until)
		if err != nil {
			return nil, fmt.Errorf("invalid until date: %w", err)
		}

		params.Set("filter[valueDate][lte]", until)
	}

	path := basePath + "?" + params.Encode()

	var allTransactions []Transaction

	for {
		resp, err := c.get(ctx, path)
		if err != nil {
			return nil, err
		}

		transactions, nextPath, err := decodeTransactionPage(resp)
		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		// Stop if we've reached the desired limit
		if opts.Limit > 0 && len(allTransactions) >= opts.Limit {
			allTransactions = allTransactions[:opts.Limit]
			break
		}

		// Stop if no more pages
		if nextPath == "" {
			break
		}

		// Extract path from full URL
		nextURL, _ := url.Parse(nextPath)
		path = nextURL.Path
		if nextURL.RawQuery != "" {
			path += "?" + nextURL.RawQuery
		}
	}

	return allTransactions, nil
}

// decodeTransactionPage decodes a page of transactions and returns the next page URL.
func decodeTransactionPage(resp *http.Response) ([]Transaction, string, error) {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", parseAPIError(resp)
	}

	var wrapper ListWrapper[Resource]
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, "", fmt.Errorf("decode response: %w", err)
	}

	transactions := make([]Transaction, 0, len(wrapper.Data))

	for _, res := range wrapper.Data {
		b, err := json.Marshal(res.Attributes)
		if err != nil {
			return nil, "", fmt.Errorf("marshal attributes: %w", err)
		}

		var tx Transaction
		if err := json.Unmarshal(b, &tx); err != nil {
			return nil, "", fmt.Errorf("unmarshal transaction: %w", err)
		}

		tx.ID = res.ID
		transactions = append(transactions, tx)
	}

	var nextPath string
	if wrapper.Links != nil && wrapper.Links.Next != "" {
		nextPath = wrapper.Links.Next
	}

	return transactions, nextPath, nil
}

// GetTransaction returns a single transaction.
func (c *Client) GetTransaction(ctx context.Context, accountID, transactionID string) (*Transaction, error) {
	path := fmt.Sprintf("/accounts/%s/transactions/%s", accountID, transactionID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeResponse[Transaction](resp)
}

// ListPendingTransactions returns pending transactions for an account.
func (c *Client) ListPendingTransactions(ctx context.Context, accountID string) ([]PendingTransaction, error) {
	path := fmt.Sprintf("/accounts/%s/pending-transactions", accountID)

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeListResponse(resp, func(t *PendingTransaction, id string) { t.ID = id })
}

// syncRequest represents the JSON:API request for creating a sync.
type syncRequest struct {
	Data syncRequestData `json:"data"`
}

type syncRequestData struct {
	Type       string           `json:"type"`
	Attributes syncRequestAttrs `json:"attributes"`
}

type syncRequestAttrs struct {
	ResourceType      string `json:"resourceType"`
	ResourceID        string `json:"resourceId"`
	Subtype           string `json:"subtype"`
	CustomerIPAddress string `json:"customerIpAddress"`
}

// CreateSync creates a new synchronization.
func (c *Client) CreateSync(ctx context.Context, accountID, subtype string) (*Synchronization, error) {
	// Detect IP for PSD2 compliance
	ip := detectOutboundIP(ctx)

	req := syncRequest{
		Data: syncRequestData{
			Type: "synchronization",
			Attributes: syncRequestAttrs{
				ResourceType:      "account",
				ResourceID:        accountID,
				Subtype:           subtype,
				CustomerIPAddress: ip,
			},
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal sync request: %w", err)
	}

	resp, err := c.post(ctx, "/synchronizations", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	return decodeResponse[Synchronization](resp)
}

// GetSync returns a synchronization status.
func (c *Client) GetSync(ctx context.Context, id string) (*Synchronization, error) {
	resp, err := c.get(ctx, "/synchronizations/"+id)
	if err != nil {
		return nil, err
	}

	return decodeResponse[Synchronization](resp)
}

// WaitForSync waits for a sync to complete.
func (c *Client) WaitForSync(ctx context.Context, id string) (*Synchronization, error) {
	for {
		sync, err := c.GetSync(ctx, id)
		if err != nil {
			return nil, err
		}

		if sync.Status == "success" || sync.Status == "error" {
			return sync, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
			// continue polling
		}
	}
}

// ListSyncs returns recent synchronizations for an account.
func (c *Client) ListSyncs(ctx context.Context, accountID string, limit int) ([]Synchronization, error) {
	path := fmt.Sprintf("/accounts/%s/synchronizations", accountID)

	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}

	resp, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	return decodeListResponse(resp, func(s *Synchronization, id string) { s.ID = id })
}

// ListFinancialInstitutions returns all financial institutions.
func (c *Client) ListFinancialInstitutions(ctx context.Context) ([]FinancialInstitution, error) {
	resp, err := c.get(ctx, "/financial-institutions")
	if err != nil {
		return nil, err
	}

	return decodeListResponse(resp, func(f *FinancialInstitution, id string) { f.ID = id })
}

// GetFinancialInstitution returns a single financial institution.
func (c *Client) GetFinancialInstitution(ctx context.Context, id string) (*FinancialInstitution, error) {
	resp, err := c.get(ctx, "/financial-institutions/"+id)
	if err != nil {
		return nil, err
	}

	return decodeResponse[FinancialInstitution](resp)
}

// GetOrganization returns the organization info.
func (c *Client) GetOrganization(ctx context.Context) (*Organization, error) {
	resp, err := c.get(ctx, "/userinfo")
	if err != nil {
		return nil, err
	}

	return decodeResponse[Organization](resp)
}

// parseDate converts a date string to ISO 8601 format (YYYY-MM-DD).
// Supports:
//   - ISO 8601 dates: "2024-01-15"
//   - Relative days: "-30d" (30 days ago), "-7d" (7 days ago)
func parseDate(s string) (string, error) {
	// Check if it's a relative date like "-30d"
	if len(s) > 1 && s[0] == '-' && s[len(s)-1] == 'd' {
		daysStr := s[1 : len(s)-1]
		days := 0

		for _, c := range daysStr {
			if c < '0' || c > '9' {
				return "", fmt.Errorf("invalid relative date: %s", s)
			}

			days = days*10 + int(c-'0')
		}

		return time.Now().AddDate(0, 0, -days).Format("2006-01-02"), nil
	}

	// Try parsing as ISO 8601 date
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s, nil
	}

	// Try parsing ISO 8601 datetime
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("2006-01-02"), nil
	}

	return "", fmt.Errorf("unsupported date format: %s (use YYYY-MM-DD or -Nd)", s)
}
