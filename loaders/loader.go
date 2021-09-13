package loaders

// -----------------------------------------------------------------------------

func Load(source string) (encodedJSON []byte, err error) {
	// Try to load from web
	encodedJSON, err = LoadFromHttp(source)

	if err == WrongFormatError {
		// If source is not a web url, try to load from hashicorp vault url
		encodedJSON, err = LoadFromVault(source)
	}

	if err == WrongFormatError {
		// If source is not a hashicorp vault url, try to load from a data url
		encodedJSON, err = LoadFromData(source)
	}

	if err == WrongFormatError {
		// At last, try to load from a file
		encodedJSON, err = LoadFromFile(source)
	}

	// Done
	return
}
