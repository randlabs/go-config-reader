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

// Loader ...
type Loader struct {
	source              string
	environmentVariable string
	cmdLineParameter    string
	loader              *loaders.Callback
	schema              []byte
	extendedValidator   *ExtendedValidator
	settingsSource      string
}

//------------------------------------------------------------------------------

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{}
}

// SetSource sets the settings source (mostly used for debugging)
func (cl *Loader) SetSource(source string) {
	cl.source = source
}

// SetEnvironmentVariable sets the environment variable which contains the settings source
func (cl *Loader) SetEnvironmentVariable(envVar string) {
	cl.environmentVariable = envVar
}

// SetCommandLineParameter sets the command-line option which contains the settings source
func (cl *Loader) SetCommandLineParameter(cmdLineParam string) {
	cl.cmdLineParameter = cmdLineParam
}

// SetCallback sets a custom settings loader
func (cl *Loader) SetCallback(loader *LoaderCallback) {
	cl.loader = (*loaders.Callback)(loader)
}

// SetSchema sets a json schema to validate used to settings
func (cl *Loader) SetSchema(schema string) {
	cl.schema = []byte(schema)
}

// SetExtendedValidator sets a validation function to be called after settings are loaded
func (cl *Loader) SetExtendedValidator(validator *ExtendedValidator) {
	cl.extendedValidator = validator
}

// GetSettingsSource gets the resolved source of the loaded settings
func (cl *Loader) GetSettingsSource() string {
	return cl.settingsSource
}

// Load settings from the specified source
func (cl *Loader) Load(settings interface{}) error {
	var encodedJSON []byte
	var err error

	// If a source was passed, use it
	source := cl.source

	// If no source, try to get source from environment variable
	if len(source) == 0 && len(cl.environmentVariable) > 0 {
		source = os.Getenv(cl.environmentVariable)
	}

	// If still no source, parse command-line arguments
	if len(source) == 0 {
		cmdLineOption := "settings"

		if len(cl.cmdLineParameter) > 0 {
			cmdLineOption = cl.cmdLineParameter
		}

		cmdLineOption = "--" + cmdLineOption

		//lookup the command-line parameter
		for idx, value := range os.Args[1:] {
			if value == cmdLineOption {
				if idx + 2 >= len(os.Args) {
					return errors.New("missing source in '" + cmdLineOption + "' parameter.")
				}
				source = os.Args[idx + 2]
				break
			}
		}
	}

	// If we reach here and no source, throw error
	if len(source) == 0 {
		return errors.New("source not defined")
	}

	// Load content from callback if one was provided
	if cl.loader != nil {
		encodedJSON, err = loaders.LoadFromCallback(cl.loader, source)
	} else {
		encodedJSON, err = loaders.Load(source)
	}
	if err != nil {
		return helpers.LoadError(err)
	}
	// Check for empty data
	if len(encodedJSON) == 0 {
		return errors.New("empty data")
	}

	// Expand variables embedded inside loaded json
	encodedJSON, err = expandVars(encodedJSON, 1)
	if err != nil {
		return helpers.LoadError(err)
	}

	// Remove comments from json
	helpers.RemoveComments(encodedJSON)

	// Validate against a schema if one is provided
	if len(cl.schema) > 0 {
		schema := cl.schema

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

	// Set the source before extended validator callback
	cl.settingsSource = source

	// Execute the extended validation if specified
	if cl.extendedValidator != nil {
		err = (*cl.extendedValidator)(settings)
		if err != nil {
			cl.settingsSource = ""
			return helpers.LoadError(err)
		}
	}

	// Done
	return nil
}

// -----------------------------------------------------------------------------
// Private functions

func expandVars(data []byte, depth int) ([]byte, error) {
	var err error

	if depth > maxExpansionLevels {
		return nil, errors.New("too many expansion levels")
	}

	// Search for ${SRC:...} and ${ENV:...}
	idx := 0
	dataLen := len(data)
	for idx < dataLen - 5 {
		// Check for tag start
		if data[idx] == '$' && data[idx + 1] == '{' && data[idx + 5] == ':' {
			// Check for SRC (source) tag
			if data[idx + 2] == 'S' && data[idx + 3] == 'R' && data[idx + 4] == 'C' {
				var source string
				var loadedData []byte

				// Get the source
				source, err = getSource(data, idx + 6, dataLen)
				if err != nil {
					return nil, err
				}

				// Load data from the specified source
				loadedData, err = loaders.Load(source)
				if err != nil {
					return nil, err
				}

				// Recursively expand variables inside loaded data
				loadedData, err = expandVars(loadedData, depth + 1)
				if err != nil {
					return nil, err
				}

				// Replace tag with loaded data
				tmp := append(data[:idx], loadedData...)
				data = append(tmp, data[(idx + 7 + len(source)):]...)

				// Advance cursor
				idx += len(loadedData)

			// Check for ENV (environment) tag
			} else if data[idx + 2] == 'E' && data[idx + 3] == 'N' && data[idx + 4] == 'V' {
				var varName string

				// Get variable name
				varName, err = getSource(data, idx + 6, dataLen)
				if err != nil {
					return nil, err
				}

				// Get value from environment strings
				varValue := []byte(os.Getenv(varName))
				if len(varValue) == 0 {
					return nil, fmt.Errorf("environment variable '%v' not found", varName)
				}

				// Recursively expand variables inside loaded data
				varValue, err = expandVars(varValue, depth + 1)
				if err != nil {
					return nil, err
				}

				// Replace tag with variable value
				tmp := append(data[:idx], varValue...)
				data = append(tmp, data[(idx + 7 + len(varName)):]...)

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
	for pos + size < dataLen && data[pos + size] != '}' {
		size += 1
	}
	if pos + size >= dataLen {
		return "", errors.New("error parsing variable")
	}

	// Return source
	return string(data[pos:(pos+size)]), nil
}
