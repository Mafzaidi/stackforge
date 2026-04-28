package serializer

// stringPtrOrNil returns nil for empty strings, or a pointer to the value.
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
