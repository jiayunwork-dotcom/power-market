// Package analyze summarises a market clearing result.
package analyze

import (
	"sort"

	"power-market/internal/bid"
)

// AwardedByGen aggregates awarded MW per generator.
// A nil or empty slice yields an empty (non-nil) map.
func AwardedByGen(awarded []bid.Bid) map[string]float64 {
	out := make(map[string]float64, len(awarded))
	for _, b := range awarded {
		out[b.Generator] += b.Capacity
	}
	return out
}

// ConsumerCost returns the settlement paid by consumers: every awarded MW is
// paid at the uniform clearing price. A nil slice yields 0.
func ConsumerCost(awarded []bid.Bid, clearingPrice float64) float64 {
	cost := 0.0
	for _, b := range awarded {
		cost += b.Capacity * clearingPrice
	}
	return cost
}

// Utilization returns the awarded MW as a fraction of total available MW.
// A total <= 0 yields 0.
func Utilization(awarded []bid.Bid, total float64) float64 {
	if total <= 0 {
		return 0
	}
	sum := 0.0
	for _, b := range awarded {
		sum += b.Capacity
	}
	return sum / total
}

// GeneratorNames returns the generators present in awarded, sorted
// alphabetically and de-duplicated, for stable reporting.
func GeneratorNames(awarded []bid.Bid) []string {
	byGen := AwardedByGen(awarded)
	names := make([]string, 0, len(byGen))
	for name := range byGen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
