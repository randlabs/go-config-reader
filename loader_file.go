package go_config_reader

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// -----------------------------------------------------------------------------

// loadFromFile tries to load the content from a file
func loadFromFile(_ context.Context, source string) ([]byte, error) {
	// NOTE: We are not making use of the context assuming configuration files will be small and on a local disk
	var err error

	source = strings.TrimPrefix(source, "file://")

	// Convert path to absolute
	if !filepath.IsAbs(source) {
		var currentPath string

		currentPath, err = os.Getwd()
		if err != nil {
			return nil, err
		}
		source = filepath.Join(currentPath, source)
	}

	// Normalize path
	source, err = filepath.Abs(source)
	if err != nil {
		return nil, err
	}

	// Load file
	return ioutil.ReadFile(source)
}
