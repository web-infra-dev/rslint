package ecmascript

import (
	"math"
	"testing"
)

func TestAcosMatchesV8(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input float64
		bits  uint64
	}{
		{input: 0.60303859878331423, bits: 0x3fed8d3e1e8ab947},
		{input: 0.5, bits: 0x3ff0c152382d7366},
		{input: -0.5, bits: 0x4000c152382d7366},
		{input: 1, bits: 0},
		{input: -1, bits: 0x400921fb54442d18},
	}
	for _, test := range tests {
		if got := math.Float64bits(Acos(test.input)); got != test.bits {
			t.Errorf("Acos(%g) bits = %#x, want %#x", test.input, got, test.bits)
		}
	}
	if !math.IsNaN(Acos(2)) {
		t.Error("Acos(2) should be NaN")
	}
}
