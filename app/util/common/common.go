package common

import (
	"path/filepath"
)

const ClientVersion = "0.4.0"

var LingoFilenames = map[string]bool{
	".lingo":      true,
	".lingo.yaml": true,
	".lingo.yml":  true,
}

// IsDotlingoFile returns if that given filepath has a recognised lingo extension.
func IsDotlingoFile(file string) bool {
	filename := filepath.Base(file)
	return LingoFilenames[filename]
}
