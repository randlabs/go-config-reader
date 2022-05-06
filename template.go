package go_config_reader

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// -----------------------------------------------------------------------------

const (
	maxExpansionLevels = 4
)

// -----------------------------------------------------------------------------

func expandVars(ctx context.Context, data []byte, depth int) ([]byte, error) {
	var err error

	if depth > maxExpansionLevels {
		return nil, errors.New("too many expansion levels")
	}

	// Search for ${SRC:...} and ${ENV:...}
	idx := 0
	dataLen := len(data)
	for idx < dataLen-6 {
		// Check for tag start
		if data[idx] == '$' && data[idx+1] == '{' && data[idx+5] == ':' {
			// Check for SRC (source) tag
			if data[idx+2] == 'S' && data[idx+3] == 'R' && data[idx+4] == 'C' {
				var source string
				var loadedData []byte

				// Get the source
				source, err = getSource(data, idx+6, dataLen)
				if err != nil {
					return nil, err
				}

				// Load data from the specified source
				loadedData, err = internalLoad(ctx, source)
				if err != nil {
					return nil, err
				}

				// Recursively expand variables inside loaded data
				loadedData, err = expandVars(ctx, loadedData, depth+1)
				if err != nil {
					return nil, err
				}

				// Replace tag with loaded data
				tmp := append(data[:idx], loadedData...)
				data = append(tmp, data[(idx+7+len(source)):]...)

				// Recalculate data length
				dataLen = len(data)

				// Advance cursor
				idx += len(loadedData)

				// Check for ENV (environment) tag
			} else if data[idx+2] == 'E' && data[idx+3] == 'N' && data[idx+4] == 'V' {
				var varName string

				// Get variable name
				varName, err = getSource(data, idx+6, dataLen)
				if err != nil {
					return nil, err
				}

				// Get value from environment strings
				varValue := []byte(os.Getenv(varName))
				if len(varValue) == 0 {
					return nil, fmt.Errorf("environment variable '%v' not found", varName)
				}

				// Recursively expand variables inside loaded data
				varValue, err = expandVars(ctx, varValue, depth+1)
				if err != nil {
					return nil, err
				}

				// Replace tag with variable value
				tmp := append(data[:idx], varValue...)
				data = append(tmp, data[(idx+7+len(varName)):]...)

				// Recalculate data length
				dataLen = len(data)

				// Advance cursor
				idx += len(varValue)

			} else {
				// Not a valid tag, advance cursor
				idx += 1
			}

		} else {
			// No tag, advance cursor
			idx += 1
		}
	}

	// Done
	return data, nil
}

func getSource(data []byte, pos int, dataLen int) (string, error) {
	// Calculate source length and look for terminator
	size := 0
	for pos+size < dataLen && data[pos+size] != '}' {
		size += 1
	}
	if pos+size >= dataLen {
		return "", errors.New("error parsing variable")
	}

	// Return source
	return string(data[pos:(pos + size)]), nil
}
