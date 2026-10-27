package utility

import "testing"

func TestValidProduct(t *testing.T) {
	tests := []struct {
		testName    string
		productName string
		price       float64
		stock       int
		want        bool
	}{
		{"valid product", "Keyboard", 25.5, 10, true},
		{"empty name", "", 25.5, 10, false},
		{"zero price", "Keyboard", 0, 10, false},
		{"negative stock", "Keyboard", 25.5, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := ValidProduct(tt.productName, tt.price, tt.stock); got != tt.want {
				t.Fatalf("ValidProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}
