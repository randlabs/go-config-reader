package go_config_reader_test

import (
	"reflect"
	"testing"

	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestWellformedWithSchema(t *testing.T) {
	// Load configuration
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: goodSettingsJSON,
		Schema: schemaJSON,
	}, &settings)
	if err != nil {
		t.Fatalf("unable to load settings [%v]", err)
	}

	if !reflect.DeepEqual(settings, goodSettings) {
		t.Fatalf("settings mismatch")
	}
}

func TestMalformedWithSchema(t *testing.T) {
	// Load configuration
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: badSettingsJSON,
		Schema: schemaJSON,
	}, &settings)
	if err == nil {
		t.Fatalf("unexpected success")
	}

	dumpValidationErrors(t, err)
}
