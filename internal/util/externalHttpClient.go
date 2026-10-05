package externalHttpClient

import (
	"context"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

type RequestOption func(*http.Request)

func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

type ExternalHttpClient struct {
	HttpClient *http.Client
	ClientName string
}

func CreateExternalHttpClient(name string) *ExternalHttpClient {
	return &ExternalHttpClient{
		HttpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		ClientName: name,
	}
}

func (c *ExternalHttpClient) Do(ctx context.Context, method, url string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	labeler := &otelhttp.Labeler{}
	labeler.Add(attribute.String("client.name", c.ClientName))
	ctx = otelhttp.ContextWithClientLabeler(ctx, labeler)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.HttpClient.Do(req)
}
