// Package clear runs a single-period uniform-price market clearing.
package clear

import (
	"power-market/internal/bid"
)

// eps absorbs float rounding when comparing accumulated MW against demand.
const eps = 1e-9

// TotalSupply returns the sum of all bid capacities in MW.
// A nil or empty slice yields 0.
func TotalSupply(bids []bid.Bid) float64 {
	total := 0.0
	for _, b := range bids {
		total += b.Capacity
	}
	return total
}

// Clear clears the market against demand (MW).
//
// Bids are walked in merit order (ascending price) and their capacity is
// accumulated until demand is met. The clearing price is the price of the
// last (marginal) awarded bid, paid uniformly to every awarded bid.
//
// When total supply >= demand, bids are awarded fully until the last one,
// which may be awarded partially with Capacity set to the remaining demand,
// and unmet is 0. When total supply < demand, every bid is awarded fully,
// the clearing price is the highest bid price and unmet is demand - supply.
//
// Bids with non-positive capacity are never awarded. A demand <= 0 clears
// nothing: price 0, no awarded bids, unmet 0. The input slice is not mutated
// and the returned slice is always non-nil.
func Clear(bids []bid.Bid, demand float64) (clearingPrice float64, awarded []bid.Bid, unmet float64) {
	awarded = make([]bid.Bid, 0, len(bids))
	if demand <= 0 {
		return 0, awarded, 0
	}

	ordered := bid.MeritOrder(bids)
	total := TotalSupply(ordered)

	if total+eps < demand {
		// Supply shortfall: everything runs, the most expensive bid sets the price.
		for _, b := range ordered {
			if b.Capacity <= 0 {
				continue
			}
			awarded = append(awarded, b)
			if b.Price > clearingPrice {
				clearingPrice = b.Price
			}
		}
		return clearingPrice, awarded, demand - total
	}

	remaining := demand
	for _, b := range ordered {
		if b.Capacity <= 0 {
			continue
		}
		take := b.Capacity
		if take > remaining {
			take = remaining
		}
		part := b
		part.Capacity = take
		awarded = append(awarded, part)
		clearingPrice = b.Price
		remaining -= take
		if remaining <= eps {
			break
		}
	}
	return clearingPrice, awarded, 0
}
