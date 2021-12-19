package go_config_reader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/qri-io/jsonschema"
	"github.com/randlabs/go-config-reader/helpers"
	"github.com/randlabs/go-config-reader/loaders"
)

// -----------------------------------------------------------------------------

const (
	maxExpansionLevels = 4
)

// -----------------------------------------------------------------------------

// LoaderCallback is the definition of the callback to call when a custom loader is used
type LoaderCallback loaders.Callback

// ExtendedValidator is the definition of the callback to call when extended validation is
// needed
type ExtendedValidator func(settings interface{}) error

// Options indicates configurable loader options
type Options struct {
	Source                string            // Optional embedded source.
	EnvironmentVariable   string            // Environment variable that contains source.
	CmdLineParameter      *string           // Long and short command-line parameters that contains source. Set to
	CmdLineParameterShort *string           //     empty to disable. Environment variable has preference.
	Callback              loaders.Callback  // Indicates a custom loader.
	Schema                string            // Specifies an optional schema validator.
	ExtendedValidator     ExtendedValidator // Specifies an extended settings validator.
	Context               context.Context   // Optional context
}

//------------------------------------------------------------------------------

// Load settings from the specified source.
func Load(options Options, settings interface{}) error {
	var encodedJSON []byte
	var err error

	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// If a source was passed, use it
	source := options.Source

	// If no source, try to get source from environment variable
	if len(source) == 0 && len(options.EnvironmentVariable) > 0 {
		source = os.Getenv(options.EnvironmentVariable)
	}

	// If still no source, parse command-line arguments
	if len(source) == 0 {
		cmdLineOption := "--settings"
		cmdLineOptionShort := "-S"
		hasCmdLineOption := true
		hasCmdLineOptionShort := true

		if options.CmdLineParameter != nil {
			if len(*options.CmdLineParameter) > 0 {
				cmdLineOption = "--" + *options.CmdLineParameter
			} else {
				hasCmdLineOption = false
			}
		}
		if options.CmdLineParameterShort != nil {
			if len(*options.CmdLineParameterShort) > 0 {
				cmdLineOptionShort = "-" + *options.CmdLineParameterShort
			} else {
				hasCmdLineOptionShort = false
			}
		}

		// Lookup the command-line parameter
		if hasCmdLineOption || hasCmdLineOptionShort {
			for idx, value := range os.Args[1:] {
				if (hasCmdLineOption && value == cmdLineOption) ||
					(hasCmdLineOptionShort && value == cmdLineOptionShort) {
					if idx+2 >= len(os.Args) {
						return errors.New("missing source in '" + value + "' parameter.")
					}
					source = os.Args[idx+2]
					break
				}
			}
		}
	}

	// If we reach here and no source, throw error
	if len(source) == 0 {
		return errors.New("source not defined")
	}

	// Load content from callback if one was provided
	if options.Callback != nil {
		encodedJSON, err = loaders.LoadFromCallback(options.Callback, source)
	} else {
		encodedJSON, err = loaders.Load(ctx, source)
	}
	if err != nil {
		return helpers.LoadError(err)
	}
	// Check for empty data
	if len(encodedJSON) == 0 {
		return errors.New("empty data")
	}

	// Expand variables embedded inside loaded json
	encodedJSON, err = expandVars(ctx, encodedJSON, 1)
	if err != nil {
		return helpers.LoadError(err)
	}

	// Remove comments from json
	helpers.RemoveComments(encodedJSON)

	// Validate against a schema if one is provided
	if len(options.Schema) > 0 {
		schema := []byte(options.Schema)

		// Remove comments from schema and decode it
		helpers.RemoveComments(schema)

		rs := &jsonschema.Schema{}
		err = json.Unmarshal(schema, rs)
		if err != nil {
			return helpers.LoadError(err)
		}

		// Execute validation
		var schemaErrors []jsonschema.KeyError

		schemaErrors, err = rs.ValidateBytes(context.Background(), encodedJSON)
		if err != nil {
			return helpers.LoadError(err)
		} else if len(schemaErrors) > 0 {
			return NewValidationError(schemaErrors)
		}
	}

	// Parse json
	err = json.Unmarshal(encodedJSON, settings)
	if err != nil {
		return helpers.LoadError(err)
	}

	// Execute the extended validation if specified
	if options.ExtendedValidator != nil {
		err = options.ExtendedValidator(settings)
		if err != nil {
			return helpers.LoadError(err)
		}
	}

	// Done
	return nil
}

// -----------------------------------------------------------------------------
// Private functions

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
				loadedData, err = loaders.Load(ctx, source)
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
