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
