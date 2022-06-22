package go_config_reader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/hashicorp/vault/api"
)

// -----------------------------------------------------------------------------

// loadFromVault tries to load the content from Hashicorp Vault
func loadFromVault(ctx context.Context, source string) ([]byte, error) {
	var client *api.Client
	var secret *api.Secret
	var buf bytes.Buffer

	if !(strings.HasPrefix(source, "vault://") || strings.HasPrefix(source, "vaults://")) {
		return nil, ErrWrongFormat
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
	keys := make([]string, 0)
	for k, v := range queryMap {
		if len(v) > 0 {
			switch k {
			case "token":
				// Set token
				token = v[0]

			case "path":
				// Set path
				path = v[0]

			case "key":
				// Set and validate key
				keys = strings.SplitN(v[0], "/", -1)
				if len(keys) == 0 {
					return nil, errors.New("invalid key")
				}
				for _, key := range keys {
					if len(key) == 0 {
						return nil, errors.New("invalid key")
					}
				}
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
	secret, err = client.Logical().ReadWithContext(ctx, path)
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

	// Prepare re-encoded for further processing
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	// Was a key specified?
	if len(keys) == 0 {
		// No, encode the whole data
		err = enc.Encode(data)
	} else {
		var value interface{}

		// Yes, transverse it
		keysLength := len(keys)
		for keyIdx := 0; keyIdx < keysLength-1; keyIdx++ {
			data, ok = data[keys[keyIdx]].(map[string]interface{})
			if !ok || data == nil {
				return nil, errors.New("key not found")
			}
		}

		// Get the final value
		value, ok = data[keys[keysLength-1]]
		if !ok || value == nil {
			return nil, errors.New("key not found")
		}

		// Check special cases
		switch v := value.(type) {
		case string:
			return []byte(v), nil
		case *string:
			return []byte(*v), nil
		}

		// Encode the rest
		err = enc.Encode(value)
	}

	// Check for encoding errors
	if err != nil {
		return nil, err
	}

	// Done
	return buf.Bytes(), nil
}
