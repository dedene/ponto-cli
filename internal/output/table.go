package output

import (
	"fmt"
	"os"
	"text/tabwriter"
	"unicode/utf8"
)

// Table provides aligned table output.
type Table struct {
	w *tabwriter.Writer
}

// NewTable creates a new table writer.
func NewTable() *Table {
	return &Table{
		w: tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
	}
}

// Header writes the table header.
func (t *Table) Header(cols ...string) {
	for i, col := range cols {
		if i > 0 {
			fmt.Fprint(t.w, "\t")
		}

		fmt.Fprint(t.w, col)
	}

	fmt.Fprintln(t.w)
}

// Row writes a table row.
func (t *Table) Row(cols ...string) {
	for i, col := range cols {
		if i > 0 {
			fmt.Fprint(t.w, "\t")
		}

		fmt.Fprint(t.w, col)
	}

	fmt.Fprintln(t.w)
}

// Flush flushes the table output.
func (t *Table) Flush() error {
	return t.w.Flush()
}

// Truncate truncates a string to maxLen runes, adding "..." if truncated.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}

	return string(runes[:maxLen-3]) + "..."
}
