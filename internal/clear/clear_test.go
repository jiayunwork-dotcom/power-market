package clear

import (
	"math"
	"testing"

	"power-market/internal/bid"
)

func sample() []bid.Bid {
	return []bid.Bid{
		{Generator: "Peaker", Kind: "peaker", Capacity: 60, Price: 120},
		{Generator: "Coal", Kind: "coal", Capacity: 200, Price: 31.75},
		{Generator: "Nuke", Kind: "nuclear", Capacity: 300, Price: 12.5},
	}
}

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// TestTotalSupply expects the sum of all capacities, 0 for nil and empty
// slices, and no mutation of the input.
func TestTotalSupply(t *testing.T) {
	if got := TotalSupply(nil); got != 0 {
		t.Errorf("TotalSupply(nil) = %.2f, want 0", got)
	}
	if got := TotalSupply([]bid.Bid{}); got != 0 {
		t.Errorf("TotalSupply(empty) = %.2f, want 0", got)
	}
	in := sample()
	if got := TotalSupply(in); !almost(got, 560) {
		t.Errorf("TotalSupply(sample) = %.2f, want 560", got)
	}
	if in[0].Capacity != 60 {
		t.Errorf("input mutated: %+v", in[0])
	}
}

// TestClear expects the marginal bid price as clearing price, a partially
// awarded last bid when demand falls inside it, unmet 0 whenever supply
// covers demand, and unmet = demand - supply with the highest price when it
// does not.
func TestClear(t *testing.T) {
	// Demand 400 MW: Nuke 300 fully, Coal partially 100 -> marginal 31.75.
	price, awarded, unmet := Clear(sample(), 400)
	if !almost(price, 31.75) {
		t.Errorf("clearing price = %.2f, want 31.75", price)
	}
	if len(awarded) != 2 {
		t.Fatalf("awarded %d bids, want 2", len(awarded))
	}
	if awarded[0].Generator != "Nuke" || !almost(awarded[0].Capacity, 300) {
		t.Errorf("awarded[0] = %+v, want Nuke 300 MW", awarded[0])
	}
	if awarded[1].Generator != "Coal" || !almost(awarded[1].Capacity, 100) {
		t.Errorf("awarded[1] = %+v, want Coal partial 100 MW", awarded[1])
	}
	if unmet != 0 {
		t.Errorf("unmet = %.2f, want 0", unmet)
	}

	// Demand exactly equal to the cheapest bid: it alone is marginal.
	price, awarded, unmet = Clear(sample(), 300)
	if !almost(price, 12.5) {
		t.Errorf("clearing price = %.2f, want 12.5", price)
	}
	if len(awarded) != 1 || !almost(awarded[0].Capacity, 300) {
		t.Errorf("awarded = %+v, want single 300 MW Nuke entry", awarded)
	}
	if unmet != 0 {
		t.Errorf("unmet = %.2f, want 0", unmet)
	}

	// Demand equal to total supply: everything runs, priciest bid is marginal.
	price, awarded, unmet = Clear(sample(), 560)
	if !almost(price, 120) {
		t.Errorf("clearing price = %.2f, want 120", price)
	}
	if len(awarded) != 3 || unmet != 0 {
		t.Errorf("awarded %d bids, unmet %.2f; want 3 bids and unmet 0", len(awarded), unmet)
	}

	// Shortfall: all bids awarded fully, price is the maximum offer.
	price, awarded, unmet = Clear(sample(), 700)
	if !almost(price, 120) {
		t.Errorf("clearing price = %.2f, want 120", price)
	}
	if len(awarded) != 3 {
		t.Fatalf("awarded %d bids, want 3", len(awarded))
	}
	if got := TotalSupply(awarded); !almost(got, 560) {
		t.Errorf("awarded supply = %.2f, want 560", got)
	}
	if !almost(unmet, 140) {
		t.Errorf("unmet = %.2f, want 140", unmet)
	}

	// Nil bids with positive demand: nothing awarded, everything unmet.
	price, awarded, unmet = Clear(nil, 100)
	if price != 0 || len(awarded) != 0 || !almost(unmet, 100) {
		t.Errorf("Clear(nil,100) = (%.2f, %+v, %.2f), want (0, empty, 100)", price, awarded, unmet)
	}
	if awarded == nil {
		t.Error("awarded slice is nil, want non-nil empty slice")
	}

	// Non-positive demand clears nothing.
	price, awarded, unmet = Clear(sample(), 0)
	if price != 0 || len(awarded) != 0 || unmet != 0 {
		t.Errorf("Clear(sample,0) = (%.2f, %+v, %.2f), want (0, empty, 0)", price, awarded, unmet)
	}
}
