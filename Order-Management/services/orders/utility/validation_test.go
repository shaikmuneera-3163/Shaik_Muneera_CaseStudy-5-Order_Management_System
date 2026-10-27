package utility

import "testing"

func TestValidOrder(t *testing.T) {
	tests := []struct {
		name      string
		productID uint
		quantity  int
		want      bool
	}{
		{"valid order", 1, 2, true},
		{"missing product", 0, 2, false},
		{"zero quantity", 1, 0, false},
		{"negative quantity", 1, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidOrder(tt.productID, tt.quantity); got != tt.want {
				t.Fatalf("ValidOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
