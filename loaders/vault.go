package loaders

import (
	"errors"
	"net/url"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/helper/jsonutil"
)

// -----------------------------------------------------------------------------

// LoadFromVault tries to load the content from Hashicorp Vault
func LoadFromVault(source string) ([]byte, error) {
	if !(strings.HasPrefix(source, "vault://") || strings.HasPrefix(source, "vaults://")) {
		return nil, WrongFormatError
	}

	source = "http" + source[5:]

	var client *api.Client
	var secret *api.Secret

	// Extract query parameters
	i := strings.Index(source, "?")
	if i < 0 {
		return nil, errors.New("invalid url")
	}
	query := source[(i+1):]
	source = source[:i]

	// Remove fragment
	i = strings.Index(query, "#")
	if i >= 0 {
		query = query[:i]
	}

	// Parse query params
	queryMap, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}

	// Extract required parameters
	token := ""
	path := ""
	for k, v := range queryMap {
		if len(v) > 0 {
			switch k {
			case "token":
				token = v[0]
			case "path":
				path = v[0]
			}
		}
	}

	// Check access token
	if len(token) == 0 {
		return nil, errors.New("missing access token")
	}

	// Check path
	if len(path) == 0 {
		return nil, errors.New("invalid path")
	}

	// Create accessor
	client, err = api.NewClient(&api.Config{
		Address: source,
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	// Read secret
	secret, err = client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("data not found")
	}

	// Done (re-encode for further processing)
	return jsonutil.EncodeJSON(data)
}
