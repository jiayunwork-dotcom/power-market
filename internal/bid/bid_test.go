package bid

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTemp creates a CSV file in the test's temp dir and returns its path.
func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(body); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

// TestParseBids expects a well-formed file to parse into bids in file order,
// and every malformed or missing input to return an error and no bids.
func TestParseBids(t *testing.T) {
	good := writeTemp(t, "good.csv", "generator,kind,capacity,price\nNuke,nuclear,300,12.5\nGas,gas,150,45.2\n")
	bids, err := ParseBids(good)
	if err != nil {
		t.Fatalf("ParseBids(good) returned error: %v", err)
	}
	if len(bids) != 2 {
		t.Fatalf("ParseBids(good) returned %d bids, want 2", len(bids))
	}
	want := Bid{Generator: "Nuke", Kind: "nuclear", Capacity: 300, Price: 12.5}
	if bids[0] != want {
		t.Errorf("bids[0] = %+v, want %+v", bids[0], want)
	}
	if bids[1].Generator != "Gas" || bids[1].Price != 45.2 {
		t.Errorf("bids[1] = %+v, want generator Gas at price 45.2", bids[1])
	}

	cases := []struct {
		name string
		path string
	}{
		{"missing file", filepath.Join(t.TempDir(), "nope.csv")},
		{"empty file", writeTemp(t, "empty.csv", "")},
		{"bad header", writeTemp(t, "hdr.csv", "gen,kind,cap,cost\nNuke,nuclear,300,12.5\n")},
		{"header only", writeTemp(t, "only.csv", "generator,kind,capacity,price\n")},
		{"short row", writeTemp(t, "short.csv", "generator,kind,capacity,price\nNuke,nuclear,300\n")},
		{"non numeric capacity", writeTemp(t, "cap.csv", "generator,kind,capacity,price\nNuke,nuclear,many,12.5\n")},
		{"negative price", writeTemp(t, "neg.csv", "generator,kind,capacity,price\nNuke,nuclear,300,-1\n")},
		{"empty generator", writeTemp(t, "gen.csv", "generator,kind,capacity,price\n ,nuclear,300,12.5\n")},
	}
	for _, tc := range cases {
		got, err := ParseBids(tc.path)
		if err == nil {
			t.Errorf("%s: expected an error, got bids %+v", tc.name, got)
		}
		if got != nil {
			t.Errorf("%s: expected nil bids on error, got %+v", tc.name, got)
		}
	}
}

// TestMeritOrder expects ascending price ordering, a stable order for equal
// prices, an untouched input slice, and an empty result for nil input.
func TestMeritOrder(t *testing.T) {
	in := []Bid{
		{Generator: "Peaker", Kind: "peaker", Capacity: 60, Price: 120},
		{Generator: "GasA", Kind: "gas", Capacity: 150, Price: 45},
		{Generator: "Nuke", Kind: "nuclear", Capacity: 300, Price: 12.5},
		{Generator: "GasB", Kind: "gas", Capacity: 100, Price: 45},
	}
	before := make([]Bid, len(in))
	copy(before, in)

	got := MeritOrder(in)
	if len(got) != len(in) {
		t.Fatalf("MeritOrder returned %d bids, want %d", len(got), len(in))
	}
	wantNames := []string{"Nuke", "GasA", "GasB", "Peaker"}
	for i, name := range wantNames {
		if got[i].Generator != name {
			t.Errorf("position %d = %q, want %q", i, got[i].Generator, name)
		}
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Price > got[i].Price {
			t.Errorf("prices not ascending at %d: %.2f > %.2f", i, got[i-1].Price, got[i].Price)
		}
	}
	for i := range in {
		if in[i] != before[i] {
			t.Errorf("input mutated at %d: %+v, want %+v", i, in[i], before[i])
		}
	}

	if empty := MeritOrder(nil); len(empty) != 0 {
		t.Errorf("MeritOrder(nil) returned %d bids, want 0", len(empty))
	}
}
