package go_config_reader

import "context"

// -----------------------------------------------------------------------------

func internalLoad(ctx context.Context, source string) (encodedJSON []byte, err error) {
	// Try to load from web
	encodedJSON, err = loadFromHttp(ctx, source)

	if err == ErrWrongFormat {
		// If source is not a web url, try to load from hashicorp vault url
		encodedJSON, err = loadFromVault(ctx, source)
	}

	if err == ErrWrongFormat {
		// If source is not a hashicorp vault url, try to load from a data url
		encodedJSON, err = loadFromData(source)
	}

	if err == ErrWrongFormat {
		// At last, try to load from a file
		encodedJSON, err = loadFromFile(ctx, source)
	}

	// Done
	return
}
