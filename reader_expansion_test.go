package go_config_reader_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestEnvironmentVariableExpansion(t *testing.T) {
	// Save test environment variables and restore on exit
	defer func(oldValue string) {
		_ = os.Setenv("GO_READER_MONGODB_USER", oldValue)
	}(os.Getenv("GO_READER_MONGODB_USER"))

	defer func(oldValue string) {
		_ = os.Setenv("GO_READER_MONGODB_PASSWORD", oldValue)
	}(os.Getenv("GO_READER_MONGODB_PASSWORD"))

	// Find a known setting and replace with environment variable sources
	pos := strings.Index(goodSettingsJSON, "mongodb://user:pass")
	if pos < 0 {
		t.Fatalf("unexpected string find failure")
	}

	source := goodSettingsJSON[0:pos] +
		"mongodb://${ENV:GO_READER_MONGODB_USER}:${ENV:GO_READER_MONGODB_PASSWORD}" +
		goodSettingsJSON[pos+19:]

	// Save the credentials in environment variables
	_ = os.Setenv("GO_READER_MONGODB_USER", "user")
	_ = os.Setenv("GO_READER_MONGODB_PASSWORD", "pass")

	// Load configuration from data stream source
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: "data://" + source,
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

func TestEmbeddedSourceExpansion(t *testing.T) {
	// Find a known setting and replace with a data source reference
	pos := strings.Index(goodSettingsJSON, "mongodb://user:pass")
	if pos < 0 {
		t.Fatalf("unexpected string find failure")
	}

	source := goodSettingsJSON[0:pos] +
		"mongodb://${SRC:data://user}:${SRC:data://pass}" +
		goodSettingsJSON[pos+19:]

	// Load configuration from data stream source
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: "data://" + source,
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
