package helpers

// Function for deciding if given data is empty
func IsEmpty(data string) bool {
	if len(data) == 0 {
		return true
	}
	return false
}
