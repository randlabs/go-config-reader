package go_config_reader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/randlabs/go-config-reader/internal/preprocessor"
)

// -----------------------------------------------------------------------------

const (
	maxExpansionLevels = 4
)

// -----------------------------------------------------------------------------

func expandVars(ctx context.Context, data []byte, depth int) ([]byte, error) {
	var expandedTagContent []byte
	var replacement []byte

	// Check recursion limits
	if depth > maxExpansionLevels {
		return nil, errors.New("too many expansion levels")
	}

	// Create a new data processor
	p := preprocessor.New(data)

	// Loop
	for {
		// Get the next tag
		ti, err := p.NextTag()
		if err != nil {
			// If we reach the end, return buffer
			if err == io.EOF {
				break

			}

			// Else the error
			return nil, err
		}

		// Expand variables that may appear inside the found content
		expandedTagContent, err = expandVars(ctx, ti.Content, depth+1)
		if err != nil {
			return nil, err
		}

		// Process tag
		switch ti.Tag {
		case preprocessor.TagSRC:
			// Load data from the specified source
			replacement, err = internalLoad(ctx, string(expandedTagContent))
			if err != nil {
				return nil, err
			}

		case preprocessor.TagENV:
			// Get value from environment strings
			v, found := os.LookupEnv(string(expandedTagContent))
			if !found {
				return nil, fmt.Errorf("environment variable '%v' not set", string(expandedTagContent))
			}
			replacement = []byte(v)

		default:
			return nil, errors.New("unexpected")
		}

		// Recursively expand variables inside loaded data
		replacement, err = expandVars(ctx, replacement, depth+1)
		if err != nil {
			return nil, err
		}

		// Execute replacement
		ti.Replace(replacement)
	}

	// Done
	return p.Data, nil
}
