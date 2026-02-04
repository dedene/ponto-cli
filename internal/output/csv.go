package output

import (
	"encoding/csv"
	"fmt"
	"os"
)

// CSV provides RFC 4180 compliant CSV output.
type CSV struct {
	w *csv.Writer
}

// NewCSV creates a new CSV writer.
func NewCSV() *CSV {
	return &CSV{
		w: csv.NewWriter(os.Stdout),
	}
}

// Header writes the CSV header.
func (c *CSV) Header(cols ...string) error {
	if err := c.w.Write(cols); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	return nil
}

// Row writes a CSV row.
func (c *CSV) Row(cols ...string) error {
	if err := c.w.Write(cols); err != nil {
		return fmt.Errorf("write csv row: %w", err)
	}

	return nil
}

// Flush flushes the CSV output.
func (c *CSV) Flush() error {
	c.w.Flush()

	return c.w.Error()
}
