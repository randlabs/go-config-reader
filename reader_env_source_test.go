package go_config_reader_test

import (
	"os"
	"reflect"
	"testing"

	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestEnvironmentVariableSource(t *testing.T) {
	// Save test environment variable and restore on exit
	defer scopedEnvVar("GO_READER_TEST")()

	// Save the data stream into test environment variable
	_ = os.Setenv("GO_READER_TEST", "data://"+goodSettingsJSON)

	// Load configuration from data stream
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		EnvironmentVariable: "GO_READER_TEST",
		Schema:              schemaJSON,
	}, &settings)
	if err != nil {
		t.Fatalf("unable to load settings [err=%v]", err)
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, goodSettings) {
		t.Fatalf("settings mismatch")
	}
}
