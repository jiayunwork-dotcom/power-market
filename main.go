// Command power-market simulates a single-period electricity market auction:
// it imports generation bids, builds the merit order, clears against demand
// and reports the clearing price plus settlement analytics.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"power-market/internal/analyze"
	"power-market/internal/bid"
	"power-market/internal/clear"
)

func main() {
	fs := flag.NewFlagSet("power-market", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	bidsPath := fs.String("bids", "", "path to bids CSV (generator,kind,capacity,price)")
	demand := fs.Float64("demand", 0, "system demand in MW (must be > 0)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: power-market -bids <path> -demand <float>\n\n")
		fmt.Fprintf(os.Stderr, "example:\n  power-market -bids example/bids.csv -demand 500\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		os.Exit(2)
	}
	if *bidsPath == "" || *demand <= 0 {
		fmt.Fprintln(os.Stderr, "error: -bids is required and -demand must be greater than 0")
		fs.Usage()
		os.Exit(2)
	}

	if err := run(*bidsPath, *demand, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(path string, demand float64, out io.Writer) error {
	bids, err := bid.ParseBids(path)
	if err != nil {
		return err
	}

	ordered := bid.MeritOrder(bids)
	total := clear.TotalSupply(ordered)
	price, awarded, unmet := clear.Clear(ordered, demand)

	fmt.Fprintf(out, "bids: %d  total supply: %.2f MW  demand: %.2f MW\n\n", len(ordered), total, demand)

	fmt.Fprintln(out, "merit order:")
	for i, b := range ordered {
		fmt.Fprintf(out, "  %2d. %-10s %-10s %8.2f MW @ %7.2f $/MWh\n", i+1, b.Generator, b.Kind, b.Capacity, b.Price)
	}

	fmt.Fprintf(out, "\nclearing price: %.2f $/MWh\n", price)

	fmt.Fprintln(out, "\nawarded:")
	if len(awarded) == 0 {
		fmt.Fprintln(out, "  (none)")
	}
	for _, b := range awarded {
		fmt.Fprintf(out, "  %-10s %-10s %8.2f MW\n", b.Generator, b.Kind, b.Capacity)
	}

	fmt.Fprintln(out, "\nawarded by generator:")
	byGen := analyze.AwardedByGen(awarded)
	for _, name := range analyze.GeneratorNames(awarded) {
		fmt.Fprintf(out, "  %-10s %8.2f MW\n", name, byGen[name])
	}

	fmt.Fprintf(out, "\nunmet demand: %.2f MW\n", unmet)
	fmt.Fprintf(out, "consumer cost: %.2f $\n", analyze.ConsumerCost(awarded, price))
	fmt.Fprintf(out, "utilization: %.2f%%\n", analyze.Utilization(awarded, total)*100)
	return nil
}
