package loaders

import (
	"strings"
)

// -----------------------------------------------------------------------------

// LoadFromData tries to load the content from a data url or tries to guess a JSON
func LoadFromData(source string) ([]byte, error) {
	if strings.HasPrefix(source, "data://") {
		return []byte(source[7:]), nil
	}

	// Try to guess a JSON
	sourceLen := len(source)
	idx := 0
	for idx < sourceLen {
		if source[idx] != ' ' && source[idx] != '\t' && source[idx] != '\r' && source[idx] != '\n' {
			break
		}
		idx += 1
	}
	if idx < sourceLen && (source[idx] == '[' || source[idx] == '{') {
		return []byte(source), nil
	}

	return nil, WrongFormatError
}
