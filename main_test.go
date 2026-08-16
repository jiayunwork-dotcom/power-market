package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunReport expects the report to contain the merit order, the clearing
// price of the marginal bid and the settlement lines for the sample file,
// and expects a missing file to return an error without writing a report.
func TestRunReport(t *testing.T) {
	var buf bytes.Buffer
	if err := run(filepath.Join("example", "bids.csv"), 500, &buf); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"merit order:",
		"NuclearOne",
		"clearing price: 31.75 $/MWh",
		"unmet demand: 0.00 MW",
		"consumer cost: 15875.00 $",
		"utilization: 52.08%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q; got:\n%s", want, out)
		}
	}

	buf.Reset()
	if err := run(filepath.Join(t.TempDir(), "missing.csv"), 500, &buf); err == nil {
		t.Error("run with a missing file returned nil error, want error")
	}
	if buf.Len() != 0 {
		t.Errorf("run wrote %d bytes on error, want 0", buf.Len())
	}
}
