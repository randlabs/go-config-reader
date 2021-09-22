package go_config_reader

import (
	"flag"
	"net/url"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/api"
)

var vaultToken = flag.String("vault-token", "", "Specifies the Vault token")

//------------------------------------------------------------------------------

func TestVaultSource(t *testing.T) {
	if vaultToken == nil || len(*vaultToken) == 0 {
		t.Skipf("Skipping Vault test because `--vault-token` flag not specified. Remember to run VAULT SERVER -DEV")
	}

	// Create vault accessor
	client, err := api.NewClient(&api.Config{
		Address: "http://127.0.0.1:8200",
	})
	if err != nil {
		t.Errorf("unable to create Vault accessor [%v]", err)
		return
	}
	client.SetToken(*vaultToken)

	// Write the test secret. NOTE: We are following the K/V v2 specs here
	_, err = client.Logical().WriteBytes("secret/data/go_client_reader_test",
		[]byte("{ \"data\": " + goodSettingsJSON + " }"))
	if err != nil {
		t.Errorf("unable to write Vault test settings [%v]", err)
		return
	}

	// Load configuration from web
	settings := &TestSettings{}
	err = Load(Options{
		Source:  "vault://127.0.0.1:8200?path=secret/data/go_client_reader_test&token=" + url.QueryEscape(*vaultToken),
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
