package go_config_reader_test

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestComplexVariableExpansion(t *testing.T) {
	// Save test environment variables and restore on exit
	defer scopedEnvVar("GO_READER_MONGODB_USER")()
	defer scopedEnvVar("GO_READER_MONGODB_PASSWORD")()
	defer scopedEnvVar("GO_READER_MONGODB_DATABASE")()
	defer scopedEnvVar("GO_READER_MONGODB_HOST")()
	defer scopedEnvVar("GO_READER_MONGODB_URL")()

	// Find a known setting and replace with a data source reference
	toReplace := "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0"
	pos := strings.Index(goodSettingsJSON, toReplace)
	if pos < 0 {
		t.Fatalf("unexpected string find failure")
	}

	// Create the modified version of the settings json
	modifiedSettingsJSON := bytes.Join([][]byte{
		([]byte(goodSettingsJSON))[0:pos],
		[]byte("${ENV:GO_READER_MONGODB_URL}"),
		([]byte(goodSettingsJSON))[pos+len(toReplace):],
	}, nil)

	// Setup our complex url which includes a source and environment variables (some embedded).
	_ = os.Setenv("GO_READER_MONGODB_URL", "mongodb://${SRC:data://${ENV:GO_READER_MONGODB_USER}:${ENV:GO_READER_MONGODB_PASSWORD}}@${ENV:GO_READER_MONGODB_HOST}/${ENV:GO_READER_MONGODB_DATABASE}?replSet=rs0")

	// Save the credentials in environment variables
	_ = os.Setenv("GO_READER_MONGODB_USER", "user")
	_ = os.Setenv("GO_READER_MONGODB_PASSWORD", "pass")

	// Also, the host and database name
	_ = os.Setenv("GO_READER_MONGODB_HOST", "127.0.0.1:27017")
	_ = os.Setenv("GO_READER_MONGODB_DATABASE", "sample_database")

	// Load configuration from data stream source
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: "data://" + string(modifiedSettingsJSON),
		Schema: schemaJSON,
	}, &settings)
	if err != nil {
		t.Fatalf("unable to load settings [err=%v]", err)
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, goodSettings) {
		t.Fatalf("settings mismatch")
	}
}
