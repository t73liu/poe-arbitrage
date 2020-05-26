package utils

func Limit(slice []string, maxSize int) []string {
	if len(slice) <= maxSize {
		return slice
	}
	result := make([]string, maxSize)
	for i := 0; i < maxSize; i++ {
		result[i] = slice[i]
	}
	return result
}

func Contains(slice []string, val string) bool {
	for _, el := range slice {
		if el == val {
			return true
		}
	}
	return false
}
