package internal

import "path/filepath"

func comparePathList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if filepath.Clean(v) != filepath.Clean(b[i]) {
			return false
		}
	}
	return true
}

func compareStringList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func combineMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range a {
		result[k] = v
	}

	for k, v := range b {
		result[k] = v
	}

	return result
}

func Contains(slice []string, element string) bool {
	for _, x := range slice {
		if x == element {
			return true
		}
	}
	return false
}
