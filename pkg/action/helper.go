package action

// boolP returns a boolean pointer.
func boolP(i bool) *bool {
	return &i
}

// stringP returns a string pointer.
func stringP(i string) *string {
	return &i
}

// slceP returns a slice pointer.
func sliceP(i []string) *[]string {
	return &i
}
