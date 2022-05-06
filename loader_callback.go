package go_config_reader

import (
	"context"
)

// -----------------------------------------------------------------------------

// loadFromCallback tries to load the content from a callback function
func loadFromCallback(ctx context.Context, cb LoaderCallback, source string) ([]byte, error) {
	content, err := cb(ctx, source)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}
