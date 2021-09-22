package go_config_reader

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

//------------------------------------------------------------------------------

func TestFileSource(t *testing.T) {
	// Create a new temporary file
	file, err := ioutil.TempFile("", "cr")
	if err != nil {
		t.Errorf("unable to create temporary file [%v]", err)
		return
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()

	// Save good settings on it
	_, err = file.Write([]byte(goodSettingsJSON))
	if err != nil {
		t.Errorf("unable to save good settings json [%v]", err)
		return
	}

	// Load configuration from file
	settings := &TestSettings{}
	err = Load(Options{
		Source:  file.Name(),
		Schema:  schemaJSON,
	}, settings)
	if err != nil {
		t.Errorf("unable to load settings [%v]", err)
		return
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, goodSettings) {
		t.Errorf("settings mismatch")
		return
	}
}
