package common

import "strings"

const ClientVersion = "0.4.0"

var LingoSuffixes = []string{".lingo", ".lingo.yaml", ".lingo.yml"}

// IsDotlingoFile returns if that given filepath has a recognised lingo extension.
func IsDotlingoFile(filepath string) bool {
	for _, suffix := range LingoSuffixes {
		if strings.HasSuffix(filepath, suffix) {
			return true
		}
	}
	return false
}
