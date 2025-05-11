package app

import "testing"

func TestGenerateRandom(t *testing.T) {
	tests := []struct {
		name   string
		max    uint64
		expect bool // Whether the output is within [0, max]
	}{
		{"max=100", 100, true},
		{"max=18446744073709551615", ^uint64(0), true}, // max uint64 value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandom(tt.max)
			if result > tt.max {
				t.Errorf("Expected %d to be <= %d, got %d", result, tt.max, result)
			}
		})
	}

	// Additional test: 10 random values for max=100
	for i := 0; i < 10; i++ {
		result := GenerateRandom(100)
		if result > 100 {
			t.Errorf("Unexpected value %d outside [0, 100]", result)
		}
	}
}
