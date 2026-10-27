package utility

func ValidOrder(productID uint, quantity int) bool {
	return productID > 0 && quantity > 0
}
