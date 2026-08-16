package analyze

import (
	"math"
	"reflect"
	"testing"

	"power-market/internal/bid"
)

func awardedSample() []bid.Bid {
	return []bid.Bid{
		{Generator: "Nuke", Kind: "nuclear", Capacity: 300, Price: 12.5},
		{Generator: "Coal", Kind: "coal", Capacity: 100, Price: 31.75},
		{Generator: "Nuke", Kind: "nuclear", Capacity: 50, Price: 20},
	}
}

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// TestAwardedByGen expects awarded MW summed per generator, an empty non-nil
// map for nil input, and no mutation of the input slice.
func TestAwardedByGen(t *testing.T) {
	in := awardedSample()
	got := AwardedByGen(in)
	want := map[string]float64{"Nuke": 350, "Coal": 100}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AwardedByGen = %v, want %v", got, want)
	}
	if len(in) != 3 || in[0].Capacity != 300 {
		t.Errorf("input mutated: %+v", in)
	}

	empty := AwardedByGen(nil)
	if empty == nil {
		t.Fatal("AwardedByGen(nil) returned a nil map, want empty map")
	}
	if len(empty) != 0 {
		t.Errorf("AwardedByGen(nil) = %v, want empty map", empty)
	}
}

// TestConsumerCost expects every awarded MW settled at the uniform clearing
// price, and 0 for nil awards or a zero price.
func TestConsumerCost(t *testing.T) {
	if got := ConsumerCost(awardedSample(), 31.75); !almost(got, 450*31.75) {
		t.Errorf("ConsumerCost = %.4f, want %.4f", got, 450*31.75)
	}
	if got := ConsumerCost(nil, 31.75); got != 0 {
		t.Errorf("ConsumerCost(nil) = %.4f, want 0", got)
	}
	if got := ConsumerCost(awardedSample(), 0); got != 0 {
		t.Errorf("ConsumerCost at price 0 = %.4f, want 0", got)
	}
}

// TestUtilization expects awarded MW divided by total MW, 0 when total is
// non-positive, and 1 when all capacity is awarded.
func TestUtilization(t *testing.T) {
	if got := Utilization(awardedSample(), 900); !almost(got, 0.5) {
		t.Errorf("Utilization(450 of 900) = %.4f, want 0.5", got)
	}
	if got := Utilization(awardedSample(), 450); !almost(got, 1) {
		t.Errorf("Utilization(450 of 450) = %.4f, want 1", got)
	}
	if got := Utilization(awardedSample(), 0); got != 0 {
		t.Errorf("Utilization with total 0 = %.4f, want 0", got)
	}
	if got := Utilization(awardedSample(), -10); got != 0 {
		t.Errorf("Utilization with negative total = %.4f, want 0", got)
	}
	if got := Utilization(nil, 900); got != 0 {
		t.Errorf("Utilization(nil, 900) = %.4f, want 0", got)
	}
}

// TestGeneratorNames expects unique generator names in alphabetical order and
// an empty result for nil input.
func TestGeneratorNames(t *testing.T) {
	got := GeneratorNames(awardedSample())
	want := []string{"Coal", "Nuke"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GeneratorNames = %v, want %v", got, want)
	}
	if got := GeneratorNames(nil); len(got) != 0 {
		t.Errorf("GeneratorNames(nil) = %v, want empty", got)
	}
}
