package output

import (
	"fmt"
	"strings"

	"github.com/dedene/ponto-cli/internal/api"
)

// Accounts outputs a list of accounts.
func Accounts(mode Mode, accounts []api.Account) error {
	switch mode {
	case ModeJSON:
		return JSON(accounts)
	case ModeCSV:
		return accountsCSV(accounts)
	case ModePlain:
		return accountsPlain(accounts)
	default:
		return accountsTable(accounts)
	}
}

func accountsTable(accounts []api.Account) error {
	t := NewTable()
	t.Header("ID", "NAME", "IBAN", "BALANCE", "CURRENCY")

	for _, a := range accounts {
		t.Row(a.ID, Truncate(a.Description, 30), a.Reference, formatAmount(a.CurrentBalance), a.Currency)
	}

	return t.Flush()
}

func accountsCSV(accounts []api.Account) error {
	c := NewCSV()
	if err := c.Header("id", "name", "iban", "balance", "currency"); err != nil {
		return err
	}

	for _, a := range accounts {
		if err := c.Row(a.ID, a.Description, a.Reference, formatAmount(a.CurrentBalance), a.Currency); err != nil {
			return err
		}
	}

	return c.Flush()
}

func accountsPlain(accounts []api.Account) error {
	for _, a := range accounts {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", a.ID, a.Description, a.Reference, formatAmount(a.CurrentBalance), a.Currency)
	}

	return nil
}

// Account outputs a single account.
func Account(mode Mode, account *api.Account) error {
	if mode == ModeJSON {
		return JSON(account)
	}

	fmt.Printf("ID:          %s\n", account.ID)
	fmt.Printf("Description: %s\n", account.Description)
	fmt.Printf("Reference:   %s\n", account.Reference)
	fmt.Printf("Product:     %s\n", account.Product)
	fmt.Printf("Balance:     %s %s\n", formatAmount(account.CurrentBalance), account.Currency)
	fmt.Printf("Available:   %s %s\n", formatAmount(account.AvailableBalance), account.Currency)

	return nil
}

// Transactions outputs a list of transactions.
func Transactions(mode Mode, txns []api.Transaction) error {
	switch mode {
	case ModeJSON:
		return JSON(txns)
	case ModeCSV:
		return transactionsCSV(txns)
	case ModePlain:
		return transactionsPlain(txns)
	default:
		return transactionsTable(txns)
	}
}

func transactionsTable(txns []api.Transaction) error {
	t := NewTable()
	t.Header("ID", "DATE", "COUNTERPART", "IBAN", "COMMUNICATION", "AMOUNT")

	for _, tx := range txns {
		comm := extractCommunication(tx.RemittanceInfo, tx.RemittanceInfoType, tx.CounterpartName)
		t.Row(tx.ID, formatDate(tx.ExecutionDate), Truncate(tx.CounterpartName, 25), tx.CounterpartRef, Truncate(comm, 40), formatAmount(tx.Amount))
	}

	return t.Flush()
}

func transactionsCSV(txns []api.Transaction) error {
	c := NewCSV()
	if err := c.Header("id", "date", "counterpart_name", "counterpart_iban", "communication", "remittance_type", "remittance_info", "amount", "currency"); err != nil {
		return err
	}

	for _, tx := range txns {
		comm := extractCommunication(tx.RemittanceInfo, tx.RemittanceInfoType, tx.CounterpartName)
		if err := c.Row(tx.ID, formatDate(tx.ExecutionDate), tx.CounterpartName, tx.CounterpartRef, comm, tx.RemittanceInfoType, tx.RemittanceInfo, formatAmount(tx.Amount), tx.Currency); err != nil {
			return err
		}
	}

	return c.Flush()
}

func transactionsPlain(txns []api.Transaction) error {
	for _, tx := range txns {
		comm := extractCommunication(tx.RemittanceInfo, tx.RemittanceInfoType, tx.CounterpartName)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", tx.ID, formatDate(tx.ExecutionDate), tx.CounterpartName, tx.CounterpartRef, comm, formatAmount(tx.Amount))
	}

	return nil
}

// Transaction outputs a single transaction.
func Transaction(mode Mode, tx *api.Transaction) error {
	if mode == ModeJSON {
		return JSON(tx)
	}

	comm := extractCommunication(tx.RemittanceInfo, tx.RemittanceInfoType, tx.CounterpartName)

	fmt.Printf("ID:            %s\n", tx.ID)
	fmt.Printf("Date:          %s\n", formatDate(tx.ExecutionDate))
	fmt.Printf("Counterpart:   %s\n", tx.CounterpartName)
	fmt.Printf("IBAN:          %s\n", tx.CounterpartRef)
	fmt.Printf("Amount:        %s %s\n", formatAmount(tx.Amount), tx.Currency)
	fmt.Printf("Communication: %s\n", comm)

	if tx.RemittanceInfoType == "unstructured" && comm != tx.RemittanceInfo {
		fmt.Printf("Full info:     %s\n", tx.RemittanceInfo)
	}

	return nil
}

// PendingTransactions outputs a list of pending transactions.
func PendingTransactions(mode Mode, txns []api.PendingTransaction) error {
	switch mode {
	case ModeJSON:
		return JSON(txns)
	default:
		return pendingTransactionsTable(txns)
	}
}

func pendingTransactionsTable(txns []api.PendingTransaction) error {
	if len(txns) > 0 {
		fmt.Println("Warning: Pending transactions may change or disappear when booked.")
		fmt.Println()
	}

	t := NewTable()
	t.Header("STATUS", "DATE", "COUNTERPART", "DESCRIPTION", "AMOUNT")

	for _, tx := range txns {
		t.Row("[PENDING]", formatDate(tx.ValueDate), Truncate(tx.CounterpartName, 25), Truncate(tx.Description, 35), formatAmount(tx.Amount))
	}

	return t.Flush()
}

// Sync outputs a sync.
func Sync(mode Mode, sync *api.Synchronization) error {
	if mode == ModeJSON {
		return JSON(sync)
	}

	fmt.Printf("ID:      %s\n", sync.ID)
	fmt.Printf("Status:  %s\n", sync.Status)
	fmt.Printf("Subtype: %s\n", sync.Subtype)

	if sync.UpdatedAt != "" {
		fmt.Printf("Updated: %s\n", formatDate(sync.UpdatedAt))
	}

	return nil
}

// Syncs outputs a list of syncs.
func Syncs(mode Mode, syncs []api.Synchronization) error {
	if mode == ModeJSON {
		return JSON(syncs)
	}

	t := NewTable()
	t.Header("ID", "STATUS", "SUBTYPE", "UPDATED")

	for _, s := range syncs {
		t.Row(s.ID, s.Status, s.Subtype, formatDate(s.UpdatedAt))
	}

	return t.Flush()
}

// Organization outputs organization info.
func Organization(mode Mode, org *api.Organization) error {
	if mode == ModeJSON {
		return JSON(org)
	}

	fmt.Printf("ID:   %s\n", org.ID)
	fmt.Printf("Name: %s\n", org.Name)

	return nil
}

// FinancialInstitutions outputs a list of financial institutions.
func FinancialInstitutions(mode Mode, institutions []api.FinancialInstitution) error {
	if mode == ModeJSON {
		return JSON(institutions)
	}

	t := NewTable()
	t.Header("ID", "NAME", "COUNTRY", "STATUS")

	for _, fi := range institutions {
		t.Row(fi.ID, Truncate(fi.Name, 40), fi.Country, fi.Status)
	}

	return t.Flush()
}

// FinancialInstitution outputs a single financial institution.
func FinancialInstitution(mode Mode, fi *api.FinancialInstitution) error {
	if mode == ModeJSON {
		return JSON(fi)
	}

	fmt.Printf("ID:      %s\n", fi.ID)
	fmt.Printf("Name:    %s\n", fi.Name)
	fmt.Printf("Country: %s\n", fi.Country)
	fmt.Printf("Status:  %s\n", fi.Status)

	if fi.MaintenanceFrom != "" {
		fmt.Printf("\n⚠ Maintenance: %s to %s\n", fi.MaintenanceFrom, fi.MaintenanceTo)
	}

	return nil
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func formatDate(isoDate string) string {
	if len(isoDate) >= 10 {
		return isoDate[:10]
	}
	return isoDate
}

// extractCommunication extracts the meaningful payment reference from remittance info.
// For structured remittance, returns as-is.
// For unstructured, tries to extract the reference after BIC/IBAN noise.
func extractCommunication(remittanceInfo, remittanceType, counterpartName string) string {
	// Structured remittance is already clean
	if remittanceType == "structured" {
		return remittanceInfo
	}

	// For unstructured, try to extract the meaningful part
	// Common pattern: "{Name} Overschrijving {IBAN} BIC: {BIC} {reference}"
	info := remittanceInfo

	// Try to find reference after "BIC: XXXXXXXXX "
	if idx := strings.Index(info, "BIC:"); idx != -1 {
		after := info[idx+4:] // skip "BIC:"
		// Skip the BIC code (usually 8-11 chars) and space
		parts := strings.Fields(after)
		if len(parts) >= 2 {
			// Return everything after the BIC code
			return strings.Join(parts[1:], " ")
		}
	}

	// Fallback: if info starts with counterpart name, try removing common prefixes
	if counterpartName != "" && strings.HasPrefix(info, counterpartName) {
		info = strings.TrimPrefix(info, counterpartName)
		info = strings.TrimSpace(info)
		// Remove "Overschrijving" / "Instantoverschrijving" prefix
		info = strings.TrimPrefix(info, "Overschrijving ")
		info = strings.TrimPrefix(info, "Instantoverschrijving ")
		info = strings.TrimPrefix(info, "Doorlopende opdracht ")
		// Remove IBAN pattern (BE## #### #### ####)
		if len(info) > 20 && info[0:2] == "BE" {
			// Skip IBAN (format: BE## #### #### ####) = 19 chars with spaces
			if idx := strings.Index(info[19:], " "); idx != -1 {
				info = strings.TrimSpace(info[19+idx:])
			}
		}
	}

	// If still long, return as-is (truncated elsewhere)
	return info
}
