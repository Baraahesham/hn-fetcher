package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type RestClient struct {
	httpClient *http.Client
	logger     zerolog.Logger
}

/*
	type RestAdapter interface {
		ExecuteHttpRequest(ctx context.Context, method string, rawURL string, queryParams map[string]string, body []byte, headers map[string]string) ([]byte, int, error)
	}
*/
type NewRestClientParams struct {
	Logger  *zerolog.Logger
	Timeout time.Duration
}

func NewRestClient(params NewRestClientParams) *RestClient {
	logger := params.Logger.With().Str("component", "RestClient").Logger()
	return &RestClient{
		httpClient: &http.Client{
			Timeout: params.Timeout,
		},
		logger: logger,
	}
}

func (client *RestClient) ExecuteHttpRequest(
	ctx context.Context,
	method string,
	rawURL string,
	queryParams map[string]string,
	body []byte,
	headers map[string]string,
) ([]byte, int, error) {
	contextLogger := client.logger.With().
		Str("url", rawURL).
		Str("method", method).
		Logger()
	// Parse and attach query parameters
	url, err := url.Parse(rawURL)
	if err != nil {
		contextLogger.Error().Err(err).Msg("Failed to parse URL")
		return nil, http.StatusInternalServerError, err
	}

	q := url.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	url.RawQuery = q.Encode()

	// Build the request body if needed
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, url.String(), bodyReader)
	if err != nil {
		contextLogger.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, http.StatusInternalServerError, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	res, err := client.httpClient.Do(req)
	if err != nil {
		contextLogger.Error().Err(err).Msg("Failed to execute HTTP request")
		return nil, http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return responseBody, res.StatusCode, errors.New("non-success status code: " + res.Status)
	}

	return responseBody, res.StatusCode, nil
}
