package loaders

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const httpRequestTimeout = 10 * time.Second
const httpResponseHeadersTimeout = 5 * time.Second

// -----------------------------------------------------------------------------

// LoadFromHttp tries to load the content from a web url
func LoadFromHttp(ctx context.Context, source string) ([]byte, error) {
	var resp *http.Response

	if !(strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")) {
		return nil, WrongFormatError
	}

	// Create custom http transport
	// From: https://www.loginradius.com/blog/async/tune-the-go-http-client-for-high-performance/
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	httpTransport.MaxIdleConns = 10
	httpTransport.MaxConnsPerHost = 10
	httpTransport.IdleConnTimeout = 60 * time.Second
	httpTransport.MaxIdleConnsPerHost = 10
	httpTransport.ResponseHeaderTimeout = httpResponseHeadersTimeout

	// Prepare request
	client := http.Client{
		Transport: httpTransport,
	}

	req, err := http.NewRequest("GET", source, nil)
	if err != nil {
		return nil, err
	}

	// Execute request
	ctxWithTimeout, ctxCancel := context.WithTimeout(ctx, httpRequestTimeout)
	defer ctxCancel()
	resp, err = client.Do(req.WithContext(ctxWithTimeout))
	if err != nil {
		return nil, err
	}

	// Check if the request succeeded
	if resp.StatusCode != 200 {
		_ = resp.Body.Close()

		return nil, fmt.Errorf("unexpected HTTP status code [http-status=%v]", resp.Status)
	}

	// Read response body
	var responseBody []byte
	responseBody, err = ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Done
	return responseBody, nil
}
