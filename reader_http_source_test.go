package go_config_reader

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

//------------------------------------------------------------------------------

func TestHttpSource(t *testing.T) {
	// Create a test http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route
		if r.URL != nil {
			if r.Method == "GET" {
				if r.URL.Path == "/settings" {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(goodSettingsJSON))
					return
				}
			}
		}

		// Else return bad request
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("bad request"))
	}))
	defer func() {
		svr.Close()
	}()

	// Load configuration from web
	settings := &TestSettings{}
	err := Load(Options{
		Source:  svr.URL + "/settings",
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
