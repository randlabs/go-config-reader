package go_config_reader_test

import (
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	cf "github.com/randlabs/go-config-reader"
)

//------------------------------------------------------------------------------

func TestVaultSource(t *testing.T) {
	checkVaultAvailability(t)

	// Create vault accessor
	client := createVaultClient(t)

	// Write the secrets
	writeSecret(t, client, "settings", goodSettingsJSON)

	// Load configuration from vault
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: buildVaultUrl("settings"),
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

func TestComplexVaultSource(t *testing.T) {
	checkVaultAvailability(t)

	// Find a known setting and replace with a data source reference
	toReplace := "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0"
	pos := strings.Index(goodSettingsJSON, toReplace)
	if pos < 0 {
		t.Fatalf("unexpected string find failure")
	}

	// Create the modified version of the settings json
	modifiedSettingsJSON := goodSettingsJSON[0:pos] +
		"${SRC:" + buildVaultUrl("database") + "&key=url}" + // Here we are replacing the database settings with a vault url
		goodSettingsJSON[pos+len(toReplace):]

	// Create vault accessor
	client := createVaultClient(t)

	// Write the secrets
	writeSecret(t, client, "settings", modifiedSettingsJSON)
	writeSecret(t, client, "database", `{ "url": "`+toReplace+`" }`)

	// Load configuration from vault
	settings := TestSettings{}
	err := cf.Load(cf.Options{
		Source: buildVaultUrl("settings"),
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

//------------------------------------------------------------------------------

func checkVaultAvailability(t *testing.T) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", "8200"), time.Second)
	if err != nil {
		t.Logf("vault server is not running. skiping....")
		t.Logf("please execute this command in order to run Vault tests:")
		t.Logf(`    vault server -dev -dev-root-token-id="root" --dev-no-store-token`)
		t.SkipNow()
	}
	if conn != nil {
		_ = conn.Close()
	}
}

func createVaultClient(t *testing.T) *api.Client {
	// Create vault accessor
	client, err := api.NewClient(&api.Config{
		Address: "http://127.0.0.1:8200",
	})
	if err != nil {
		t.Fatalf("unable to create Vault client [err=%v]", err)
	}
	client.SetToken("root")

	// Done
	return client
}

func writeSecret(t *testing.T, client *api.Client, key string, secret string) {
	// Write the test secret. NOTE: We are following the K/V v2 specs here.
	value := `{ "data": ` + secret + `}`
	_, err := client.Logical().WriteBytes(key2path(key), []byte(value))
	if err != nil {
		t.Fatalf("unable to write Vault [key=%v] [err=%v]", key, err)
	}
}

func buildVaultUrl(key string) string {
	return "vault://127.0.0.1:8200?path=" + key2path(key) + "&token=root"
}

func key2path(key string) string {
	return "secret/data/go_reader_test/" + key
}
