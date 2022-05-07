package go_config_reader_test

import (
	"flag"
	"net/url"
	"reflect"
	"testing"

	"github.com/hashicorp/vault/api"
	cf "github.com/randlabs/go-config-reader"
)

var vaultToken = flag.String("vault-token", "", "Specifies the Vault token")

//------------------------------------------------------------------------------

func TestVaultSource(t *testing.T) {
	if vaultToken == nil || len(*vaultToken) == 0 {
		t.Skipf("skipping Vault test because `--vault-token` flag not specified. Remember to run VAULT SERVER -DEV")
	}

	// Create vault accessor
	client, err := api.NewClient(&api.Config{
		Address: "http://127.0.0.1:8200",
	})
	if err != nil {
		t.Fatalf("unable to create Vault accessor [%v]", err)
	}
	client.SetToken(*vaultToken)

	// Write the test secret. NOTE: We are following the K/V v2 specs here
	_, err = client.Logical().WriteBytes("secret/data/go_client_reader_test",
		[]byte("{ \"data\": "+goodSettingsJSON+" }"))
	if err != nil {
		t.Fatalf("unable to write Vault test settings [%v]", err)
	}

	// Load configuration from web
	settings := TestSettings{}
	err = cf.Load(cf.Options{
		Source: "vault://127.0.0.1:8200?path=secret/data/go_client_reader_test&token=" + url.QueryEscape(*vaultToken),
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
