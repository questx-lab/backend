package testutil

import (
	"context"

	"github.com/questx-lab/backend/pkg/api"
)

type MockAPIGenerator struct {
	MockClient MockAPIClient
}

func (m *MockAPIGenerator) New(domain, path string, args ...any) api.Client {
	return &m.MockClient
}

type MockAPIClient struct {
	HeaderFunc func(name, value string) api.Client
	QueryFunc  func(query api.Parameter) api.Client
	BodyFunc   func(body api.Body) api.Client
	POSTFunc   func(ctx context.Context, opts ...api.Opt) (*api.Response, error)
	GETFunc    func(ctx context.Context, opts ...api.Opt) (*api.Response, error)
	PUTFunc    func(ctx context.Context, opts ...api.Opt) (*api.Response, error)
}

func (c *MockAPIClient) Header(name, value string) api.Client {
	if c.HeaderFunc != nil {
		return c.HeaderFunc(name, value)
	}

	return c
}

func (c *MockAPIClient) Query(query api.Parameter) api.Client {
	if c.QueryFunc != nil {
		return c.QueryFunc(query)
	}

	return c
}

func (c *MockAPIClient) Body(body api.Body) api.Client {
	if c.BodyFunc != nil {
		return c.BodyFunc(body)
	}

	return c
}

func (c *MockAPIClient) POST(ctx context.Context, opts ...api.Opt) (*api.Response, error) {
	if c.POSTFunc != nil {
		return c.POSTFunc(ctx, opts...)
	}

	panic("not implemented")
}

func (c *MockAPIClient) GET(ctx context.Context, opts ...api.Opt) (*api.Response, error) {
	if c.GETFunc != nil {
		return c.GETFunc(ctx, opts...)
	}

	panic("not implemented")
}

func (c *MockAPIClient) PUT(ctx context.Context, opts ...api.Opt) (*api.Response, error) {
	if c.PUTFunc != nil {
		return c.PUTFunc(ctx, opts...)
	}

	panic("not implemented")
}
