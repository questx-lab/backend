package api

import (
	"context"
)

type MockAPIGenerator struct {
	MockClient MockAPIClient
}

func (m *MockAPIGenerator) New(path string, args ...any) Client {
	return &m.MockClient
}

type MockAPIClient struct {
	HeaderFunc func(name, value string) Client
	QueryFunc  func(query Parameter) Client
	BodyFunc   func(body Body) Client
	POSTFunc   func(ctx context.Context, opts ...Opt) (*Response, error)
	GETFunc    func(ctx context.Context, opts ...Opt) (*Response, error)
	PUTFunc    func(ctx context.Context, opts ...Opt) (*Response, error)
}

func (c *MockAPIClient) Header(name, value string) Client {
	if c.HeaderFunc != nil {
		return c.HeaderFunc(name, value)
	}

	return c
}

func (c *MockAPIClient) Query(query Parameter) Client {
	if c.QueryFunc != nil {
		return c.QueryFunc(query)
	}

	return c
}

func (c *MockAPIClient) Body(body Body) Client {
	if c.BodyFunc != nil {
		return c.BodyFunc(body)
	}

	return c
}

func (c *MockAPIClient) POST(ctx context.Context, opts ...Opt) (*Response, error) {
	if c.POSTFunc != nil {
		return c.POSTFunc(ctx, opts...)
	}

	panic("not implemented")
}

func (c *MockAPIClient) GET(ctx context.Context, opts ...Opt) (*Response, error) {
	if c.GETFunc != nil {
		return c.GETFunc(ctx, opts...)
	}

	panic("not implemented")
}

func (c *MockAPIClient) PUT(ctx context.Context, opts ...Opt) (*Response, error) {
	if c.PUTFunc != nil {
		return c.PUTFunc(ctx, opts...)
	}

	panic("not implemented")
}
