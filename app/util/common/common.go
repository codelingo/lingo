package common

import (
	"path/filepath"
)

const ClientVersion = "0.5.1"

var LingoFilenames = map[string]bool{
	"codelingo.yaml": true,
	"codelingo.yml":  true,
}

// IsDotlingoFile returns if that given filepath has a recognised lingo extension.
func IsDotlingoFile(file string) bool {
	filename := filepath.Base(file)
	return LingoFilenames[filename]
}
