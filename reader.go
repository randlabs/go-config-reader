package go_config_reader

//goland:noinspection ALL
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/qri-io/jsonschema"
)

//------------------------------------------------------------------------------

// LoaderCallback ...
type LoaderCallback func(source string) (string, error)

// ExtendedValidatorCallback ...
type ExtendedValidator func(settings interface{}) error

// Loader ...
type Loader struct {
	source              string
	environmentVariable string
	cmdLineParameter    string
	loader              *LoaderCallback
	schema              []byte
	extendedValidator   *ExtendedValidator
	settingsSource      string
}

//------------------------------------------------------------------------------

// NewLoader ...
func NewLoader() *Loader {
	return &Loader{}
}

// SetSource ...
func (cl *Loader) SetSource(source string) {
	cl.source = source
}

// SetEnvironmentVariable ...
func (cl *Loader) SetEnvironmentVariable(envVar string) {
	cl.environmentVariable = envVar
}

// SetCommandLineParameter ...
func (cl *Loader) SetCommandLineParameter(cmdLineParam string) {
	cl.cmdLineParameter = cmdLineParam
}

// SetCallback ...
func (cl *Loader) SetCallback(loader *LoaderCallback) {
	cl.loader = loader
}

// SetSchema ...
func (cl *Loader) SetSchema(schema string) {
	cl.schema = []byte(schema)
}

// SetCallback ...
func (cl *Loader) SetExtendedValidator(validator *ExtendedValidator) {
	cl.extendedValidator = validator
}

// GetSettingsSource ...
func (cl *Loader) GetSettingsSource() string {
	return cl.settingsSource
}

// Load ...
func (cl *Loader) Load(settings interface{}) error {
	var jsonContent []byte
	var err error

	// if a source was passed, use it
	source := cl.source

	// if no source, try to get source from environment variable
	if len(source) == 0 && len(cl.environmentVariable) > 0 {
		source = os.Getenv(cl.environmentVariable)
	}

	// if still no source, parse command-line arguments
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

	// if we reach here and no source, throw error
	if len(source) == 0 {
		return errors.New("source not defined")
	}

	// check for loader, if none provided, read from disk file
	if cl.loader == nil {
		if !filepath.IsAbs(source) {
			var currentPath string

			currentPath, err = os.Getwd()
			if err == nil {
				source = filepath.Join(currentPath, source)
			}
		} else {
			err = nil
		}
		if err == nil {
			source, err = filepath.Abs(source)
			if err == nil {
				jsonContent, err = ioutil.ReadFile(source)
			}
		}
	} else {
		var content string

		content, err = (*cl.loader)(source)
		if err == nil {
			jsonContent = []byte(content)
		}
	}

	// validate settings against a schema if one is provided
	if err == nil && len(cl.schema) > 0 {
		rs := &jsonschema.Schema{}

		removeComments(jsonContent)

		err = json.Unmarshal(cl.schema, rs)
		if err == nil {
			var schemaErrors []jsonschema.KeyError

			ctx := context.Background()

			schemaErrors, err = rs.ValidateBytes(ctx, jsonContent)
			if err == nil && len(schemaErrors) > 0 {
				failures := make([]ValidationErrorFailure, len(schemaErrors))

				for idx, e := range schemaErrors {
					failures[idx].Location = e.PropertyPath
					failures[idx].Message = e.Message
				}

				err = &ValidationError{
					Failures: failures,
				}
			}
		}
	}

	// now parse json
	if err == nil {
		removeComments(jsonContent)

		err = json.Unmarshal(jsonContent, settings)
	}

	if err == nil {
		//set the source file before extended validator callback
		cl.settingsSource = source

		if cl.extendedValidator != nil {
			err = (*cl.extendedValidator)(settings)
		}
	}

	if err != nil {
		cl.settingsSource = ""

		_, ok := err.(*ValidationError)
		if ok {
			return err
		}
		return fmt.Errorf("unable to load configuration [%v]", err)
	}

	//done
	return nil
}

//------------------------------------------------------------------------------
// Private methods

func removeComments(data []byte) {
	state := 0

	dataLength := len(data)

	for index := 0; index < dataLength; index++ {
		ch := data[index]

		switch state {
		case 0: //outside quotes and comments
			if ch == '"' {
				state = 1 //start of quoted value
			} else if ch == '/' {
				if index + 1 < dataLength {
					switch data[index + 1] {
					case '/':
						state = 2 //start of single line comment
						data[index] = ' '
						index += 1
						data[index] = ' '

					case '*':
						state = 3 //start of multiple line comment
						data[index] = ' '
						index += 1
						data[index] = ' '
					}
				}
			}

		case 1: //inside quoted value
			if ch == '\\' {
				if index + 1 < dataLength {
					index += 1 //escaped character
				}
			} else if ch == '"' {
				state = 0 //end of quoted value
			}

		case 2: //single line comment
			if ch == '\n' {
				state = 0 //end of single line comment
			} else {
				data[index] = ' ' //remove comment
			}

		case 3: //multiple line comment
			if ch == '*' && index + 1 < dataLength && data[index + 1] == '/' {
				state = 0 //end of single line comment
				index += 1
			} else {
				data[index] = ' ' //remove comment
			}
		}
	}
}
