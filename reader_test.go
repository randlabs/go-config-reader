package go_config_reader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

//------------------------------------------------------------------------------

type Settings struct {
	Name         string `json:"name"`
	IntegerValue int `json:"integerValue"`
	FloatValue   float64 `json:"floatValue"`
	Server struct {
		Ip               string `json:"ip"`
		Port             int `json:"port"`
		PoolSize         int `json:"poolSize"`
		AllowedAddresses []string `json:"allowedAddresses"`
	} `json:"server"`
	Node struct {
		Url      string `json:"url"`
		ApiToken string `json:"apiToken"`
	} `json:"node"`

	MongoDB struct {
		Url string `json:"url"`
	} `json:"mongodb"`
}

//------------------------------------------------------------------------------

func TestWellformedWithSchema(t *testing.T) {
	schema, err := loadTestFile("./test/schema/settings.schema.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	configLoader := NewLoader()
	configLoader.SetSource("./test/json/settings_good.json")
	configLoader.SetSchema(string(schema))

	settings := &Settings{}
	err = configLoader.Load(settings)
	if err != nil {
		t.Errorf("unable to initialize. [%v]", err)
		return
	}
}

func TestMalformedWithSchema(t *testing.T) {
	schema, err := loadTestFile("./test/schema/settings.schema.json")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	configLoader := NewLoader()
	configLoader.SetSource("./test/json/settings_bad.json")
	configLoader.SetSchema(string(schema))

	settings := &Settings{}
	err = configLoader.Load(settings)
	if err == nil {
		t.Errorf("unexpected success.")
		return
	}

	dumpValidationErrors(t, err)
}

//------------------------------------------------------------------------------

func loadTestFile(filename string) ([]byte, error) {
	var data []byte

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get current directory. [%v]", err)
	}

	data, err = ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return nil, fmt.Errorf("cannot read file. [%v]", err)
	}

	return data, nil
}

func dumpValidationErrors(t *testing.T, err error) {
	e, ok := err.(*ValidationError)
	if ok {
		t.Logf("  Validation errors:")
		for _, f := range e.Failures {
			t.Logf("  %v @ %v", f.Message, f.Location)
		}
	}
}
