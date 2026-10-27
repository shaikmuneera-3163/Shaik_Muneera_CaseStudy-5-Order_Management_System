package utility

func ValidProduct(name string, price float64, stock int) bool {
	return name != "" && price > 0 && stock >= 0
}
