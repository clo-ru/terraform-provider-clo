// Package cloapi is the provider's adapter over the generated CLO v3 SDK
// (github.com/clo-ru/cloapi-go-client/v3). Resources and data sources call the
// stable methods defined here instead of the generated client directly, so that
// changes to generated names/shapes are absorbed in this one package rather than
// rippling across every resource. See V3_MIGRATION_SCOPE.md.
package cloapi

import gen "github.com/clo-ru/cloapi-go-client/v3"

// Client wraps the generated v3 client behind provider-stable methods.
type Client struct {
	gen *gen.ClientWithResponses
}

// New builds an adapter client for the given token and base URL.
func New(token, baseURL string) (*Client, error) {
	g, err := gen.New(token, gen.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}
	return &Client{gen: g}, nil
}

// IsNotFound reports whether err is a 404 from the API. Re-exported so callers
// (waiters, Read funcs) depend only on this adapter package.
func IsNotFound(err error) bool { return gen.IsNotFound(err) }
