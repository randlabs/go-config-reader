package go_config_reader_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestHttpSource(t *testing.T) {
	// Create a test http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route
		switch r.Method {
		case "GET":
			switch r.URL.Path {
			case "/settings":
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(goodSettingsJSON))
				return
			}
		}

		// Else return bad request
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer svr.Close()

	// Load configuration from web
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: svr.URL + "/settings",
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
