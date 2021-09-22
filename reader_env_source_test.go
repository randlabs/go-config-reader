package go_config_reader

import (
	"os"
	"reflect"
	"testing"
)

//------------------------------------------------------------------------------

func TestEnvironmentVariableSource(t *testing.T) {
	// Save test environment variable and restore on exit
	defer func(oldGoReaderTestEnv string) {
		_ = os.Setenv("GO_READER_TEST", oldGoReaderTestEnv)
	}(os.Getenv("GO_READER_TEST"))

	// Save the data stream into test environment variable
	_ = os.Setenv("GO_READER_TEST", "data://" + goodSettingsJSON)

	// Load configuration from data stream
	settings := &TestSettings{}
	err := Load(Options{
		EnvironmentVariable: "GO_READER_TEST",
		Schema:              schemaJSON,
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
