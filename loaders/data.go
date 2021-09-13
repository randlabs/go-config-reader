package loaders

import (
	"strings"
)

// -----------------------------------------------------------------------------

// LoadFromData tries to load the content from a data url
func LoadFromData(source string) ([]byte, error) {
	if !strings.HasPrefix(source, "data://") {
		return nil, WrongFormatError
	}

	return []byte(source)[7:], nil
}
