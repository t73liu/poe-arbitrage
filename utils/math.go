package utils

// Greatest common divisor using Euclidean Algorithm
func CalcGCD(a, b uint) uint {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}

func CalcMin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
