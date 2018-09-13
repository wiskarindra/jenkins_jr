package api

func isInSliceString(v string, slice []string) bool {
	for _, s := range slice {
		if v == s {
			return true
		}
	}
	return false
}
