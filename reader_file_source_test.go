package go_config_reader_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestFileSource(t *testing.T) {
	// Create a new temporary file
	file, err := ioutil.TempFile("", "cr")
	if err != nil {
		t.Fatalf("unable to create temporary file [%v]", err)
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()

	// Save good settings on it
	_, err = file.Write([]byte(goodSettingsJSON))
	if err != nil {
		t.Fatalf("unable to save good settings json [%v]", err)
	}

	// Load configuration from file
	settings := TestSettings{}
	err = cf.Load(cf.Options{
		Source: file.Name(),
		Schema: schemaJSON,
	}, &settings)
	if err != nil {
		t.Fatalf("unable to load settings [%v]", err)
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, goodSettings) {
		t.Fatalf("settings mismatch")
	}
}
