// Package bid loads generation offers (bids) from CSV and orders them by price.
package bid

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Bid is a single generation offer into the day-ahead market.
// Price is the offer in $/MWh, Capacity is the available MW.
type Bid struct {
	Generator string
	Kind      string
	Capacity  float64
	Price     float64
}

// ParseBids reads a bids CSV file with header
// "generator,kind,capacity,price" and returns the parsed bids.
// It returns an error when the file is missing, the header is wrong,
// a row has the wrong number of fields, a numeric field is unparsable,
// capacity/price is negative or not finite, or the file holds no data rows.
func ParseBids(path string) ([]Bid, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open bids file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("bids file %s is empty", path)
	}
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if err := checkHeader(header); err != nil {
		return nil, err
	}

	bids := make([]Bid, 0, 8)
	line := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read record after line %d: %w", line, err)
		}
		line++

		if len(rec) == 1 && strings.TrimSpace(rec[0]) == "" {
			continue // tolerate blank line
		}
		if len(rec) != 4 {
			return nil, fmt.Errorf("line %d: expected 4 fields, got %d", line, len(rec))
		}

		gen := strings.TrimSpace(rec[0])
		if gen == "" {
			return nil, fmt.Errorf("line %d: generator must not be empty", line)
		}
		kind := strings.TrimSpace(rec[1])
		if kind == "" {
			return nil, fmt.Errorf("line %d: kind must not be empty", line)
		}
		capMW, err := parseNumber(rec[2])
		if err != nil {
			return nil, fmt.Errorf("line %d: capacity: %w", line, err)
		}
		price, err := parseNumber(rec[3])
		if err != nil {
			return nil, fmt.Errorf("line %d: price: %w", line, err)
		}

		bids = append(bids, Bid{Generator: gen, Kind: kind, Capacity: capMW, Price: price})
	}

	if len(bids) == 0 {
		return nil, fmt.Errorf("bids file %s has no data rows", path)
	}
	return bids, nil
}

// MeritOrder returns a new slice sorted by ascending Price using a stable
// sort, so bids offered at the same price keep their input order.
// The input slice is never mutated; a nil input yields an empty slice.
func MeritOrder(bids []Bid) []Bid {
	out := make([]Bid, len(bids))
	copy(out, bids)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Price < out[j].Price
	})
	return out
}

func checkHeader(header []string) error {
	want := []string{"generator", "kind", "capacity", "price"}
	if len(header) != len(want) {
		return fmt.Errorf("bad header: expected %d columns, got %d", len(want), len(header))
	}
	for i, w := range want {
		got := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(header[i], "\ufeff")))
		if got != w {
			return fmt.Errorf("bad header: column %d is %q, want %q", i+1, got, w)
		}
	}
	return nil
}

func parseNumber(s string) (float64, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return 0, fmt.Errorf("value must not be empty")
	}
	v, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", t)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%q is not a finite number", t)
	}
	if v < 0 {
		return 0, fmt.Errorf("%q must not be negative", t)
	}
	return v, nil
}
