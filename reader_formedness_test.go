package go_config_reader

import (
	"reflect"
	"testing"
)

//------------------------------------------------------------------------------

func TestWellformedWithSchema(t *testing.T) {
	// Load configuration
	settings := &TestSettings{}
	err := Load(Options{
		Source:  goodSettingsJSON,
		Schema:  schemaJSON,
	}, settings)
	if err != nil {
		t.Errorf("unable to load settings [%v]", err)
		return
	}

	if !reflect.DeepEqual(settings, goodSettings) {
		t.Errorf("settings mismatch")
		return
	}
}

func TestMalformedWithSchema(t *testing.T) {
	// Load configuration
	settings := &TestSettings{}
	err := Load(Options{
		Source:  badSettingsJSON,
		Schema:  schemaJSON,
	}, settings)
	if err == nil {
		t.Errorf("unexpected success")
		return
	}

	dumpValidationErrors(t, err)
}
