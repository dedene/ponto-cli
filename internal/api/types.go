package api

// Account represents a Ponto account.
type Account struct {
	ID               string  `json:"id"`
	Description      string  `json:"description"`
	Reference        string  `json:"reference"` // IBAN
	Product          string  `json:"product"`
	Currency         string  `json:"currency"`
	CurrentBalance   float64 `json:"currentBalance"`
	AvailableBalance float64 `json:"availableBalance"`
	Deprecated       bool    `json:"deprecated"`
}

// Transaction represents a Ponto transaction.
type Transaction struct {
	ID                  string  `json:"id"`
	Amount              float64 `json:"amount"`
	Currency            string  `json:"currency"`
	Description         string  `json:"description"`
	CounterpartName     string  `json:"counterpartName"`
	CounterpartRef      string  `json:"counterpartReference"`
	RemittanceInfo      string  `json:"remittanceInformation"`
	RemittanceInfoType  string  `json:"remittanceInformationType"`
	EndToEndID          string  `json:"endToEndId"`
	InternalRef         string  `json:"internalReference"`
	BankTransactionCode string  `json:"bankTransactionCode"`
	ExecutionDate       string  `json:"executionDate"`
	ValueDate           string  `json:"valueDate"`
}

// PendingTransaction represents a pending transaction.
type PendingTransaction struct {
	ID              string  `json:"id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Description     string  `json:"description"`
	CounterpartName string  `json:"counterpartName"`
	CounterpartRef  string  `json:"counterpartReference"`
	RemittanceInfo  string  `json:"remittanceInformation"`
	ValueDate       string  `json:"valueDate"`
}

// Synchronization represents a sync operation.
type Synchronization struct {
	ID        string `json:"id"`
	Subtype   string `json:"subtype"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Errors    []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// FinancialInstitution represents a bank.
type FinancialInstitution struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Country         string `json:"country"`
	Status          string `json:"status"`
	MaintenanceFrom string `json:"maintenanceFrom,omitempty"`
	MaintenanceTo   string `json:"maintenanceTo,omitempty"`
}

// Organization represents the user's organization.
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TransactionListOptions for filtering transactions.
type TransactionListOptions struct {
	Since string
	Until string
	Limit int
}

// JSON:API response wrappers

// DataWrapper wraps a single resource.
type DataWrapper[T any] struct {
	Data T `json:"data"`
}

// ListWrapper wraps a list of resources.
type ListWrapper[T any] struct {
	Data  []T       `json:"data"`
	Links *Links    `json:"links,omitempty"`
	Meta  *ListMeta `json:"meta,omitempty"`
}

// Links for pagination.
type Links struct {
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
}

// ListMeta contains pagination metadata.
type ListMeta struct {
	Paging *Paging `json:"paging,omitempty"`
}

// Paging contains cursor info.
type Paging struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// Resource represents a JSON:API resource.
type Resource struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

// APIError represents an API error.
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return e.Code
}
