package go_config_reader

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/helper/jsonutil"
)

// -----------------------------------------------------------------------------

// loadFromVault tries to load the content from Hashicorp Vault
func loadFromVault(ctx context.Context, source string) ([]byte, error) {
	if !(strings.HasPrefix(source, "vault://") || strings.HasPrefix(source, "vaults://")) {
		return nil, WrongFormatError
	}

	source = "http" + source[5:]

	// Extract query parameters
	i := strings.Index(source, "?")
	if i < 0 {
		return nil, errors.New("invalid url")
	}
	query := source[(i + 1):]
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
	var client *api.Client
	client, err = api.NewClient(&api.Config{
		Address: source,
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	// Read secret

	// NOTE: We have to duplicate the 'secret, err = client.Logical().Read(path)' behavior in order to use our
	//       custom context. The code below is a bit optimized version of the Vault's one and with better error
	//       handling (2021/12/19)
	var secret *api.Secret

	req := client.NewRequest("GET", "/v1/"+path)

	var resp *api.Response
	resp, err = client.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()

		secret, err = api.ParseSecret(resp.Body)

		if resp.StatusCode == 404 {
			if err == nil {
				if !(secret != nil && (len(secret.Warnings) > 0 || len(secret.Data) > 0)) {
					secret = nil
				}
			} else if err == io.EOF {
				err = nil
			}
		}
	}
	if err != nil {
		return nil, err
	}

	// If we don't have a secret but also no errors
	if secret == nil {
		return nil, errors.New("data not found")
	}

	// Extract data
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok || data == nil {
		return nil, errors.New("data not found")
	}

	// Done (re-encode for further processing)
	return jsonutil.EncodeJSON(data)
}
