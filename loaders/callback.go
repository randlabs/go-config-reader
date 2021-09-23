package loaders

// -----------------------------------------------------------------------------

// Callback ...
type Callback func(source string) (string, error)

// -----------------------------------------------------------------------------

// LoadFromCallback tries to load the content from a callback function
func LoadFromCallback(cb Callback, source string) ([]byte, error) {
	content, err := cb(source)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}
