package go_config_reader

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/qri-io/jsonschema"
)

// -----------------------------------------------------------------------------

// Options indicates configurable loader options.
type Options struct {
	// Optional embedded source.
	Source string

	// Environment variable that contains source.
	EnvironmentVariable string

	// Long and short command-line parameters that contains source. Set to
	// empty to disable. Environment variable has preference.
	CmdLineParameter      *string
	CmdLineParameterShort *string

	// Use a custom loader for the configuration settings.
	Callback LoaderCallback

	// Specifies an optional json schema validator.
	Schema string

	// Specifies an extended settings validator callback.
	ExtendedValidator ExtendedValidator

	// Optional context to use while loading the configuration.
	Context context.Context
}

// LoaderCallback is a function to call that must return a valid configuration json when possible.
type LoaderCallback func(ctx context.Context, source string) (string, error)

// ExtendedValidator is a function to call in order to do configuration validation not covered by this library.
type ExtendedValidator func(settings interface{}) error

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

	// If still no source, parse command-line arguments and try to load from the specified
	if len(source) == 0 {
		// Setup command-line parameters to look for.
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

		// Lookup for the parameter's value.
		if hasCmdLineOption || hasCmdLineOptionShort {
			for idx, value := range os.Args[1:] {
				if (hasCmdLineOption && value == cmdLineOption) || (hasCmdLineOptionShort && value == cmdLineOptionShort) {
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
		encodedJSON, err = loadFromCallback(ctx, options.Callback, source)
	} else {
		encodedJSON, err = internalLoad(ctx, source)
	}
	if err != nil {
		return newLoadError(err)
	}

	// Expand variables embedded inside loaded json
	encodedJSON, err = expandVars(ctx, encodedJSON, 1)
	if err != nil {
		return newLoadError(err)
	}

	// If resulting configuration is empty, throw error
	if len(encodedJSON) == 0 {
		return errors.New("empty data")
	}

	// Remove comments from json
	removeComments(encodedJSON)

	// Validate against a schema if one is provided
	if len(options.Schema) > 0 {
		schema := []byte(options.Schema)

		// Remove comments from schema
		removeComments(schema)

		// Decode it
		rs := jsonschema.Schema{}
		err = json.Unmarshal(schema, &rs)
		if err != nil {
			return newLoadError(err)
		}

		// Execute validation
		var schemaErrors []jsonschema.KeyError

		schemaErrors, err = rs.ValidateBytes(context.Background(), encodedJSON)
		if err != nil {
			return newLoadError(err)
		} else if len(schemaErrors) > 0 {
			return newValidationError(schemaErrors)
		}
	}

	// Parse configuration settings json object
	err = json.Unmarshal(encodedJSON, settings)
	if err != nil {
		return newLoadError(err)
	}

	// Execute the extended validation if one was specified
	if options.ExtendedValidator != nil {
		err = options.ExtendedValidator(settings)
		if err != nil {
			return newLoadError(err)
		}
	}

	// Done
	return nil
}
