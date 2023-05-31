package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/questx-lab/backend/pkg/xcontext"
)

type Client interface {
	Header(name, value string) Client
	Query(query Parameter) Client
	Body(body Body) Client
	POST(ctx context.Context, opts ...Opt) (*Response, error)
	GET(ctx context.Context, opts ...Opt) (*Response, error)
	PUT(ctx context.Context, opts ...Opt) (*Response, error)
}

type Generator interface {
	New(domain, path string, args ...any) Client
}

type defaultGenerator struct{}

func NewGenerator() *defaultGenerator {
	return &defaultGenerator{}
}

func (*defaultGenerator) New(domain, path string, args ...any) Client {
	return &defaultClient{
		url:     fmt.Sprintf("%s%s", domain, fmt.Sprintf(path, args...)),
		headers: make(http.Header),
	}
}

type Body interface {
	ToReader() (io.Reader, error)
}

type Opt interface {
	Do(defaultClient, *http.Request)
}

type defaultClient struct {
	method  string
	url     string
	headers http.Header
	query   Parameter
	body    Body
}

func (c *defaultClient) Header(name, value string) Client {
	c.headers[name] = []string{value}
	return c
}

func (c *defaultClient) Query(query Parameter) Client {
	c.query = query
	return c
}

func (c *defaultClient) Body(body Body) Client {
	c.body = body
	return c
}

func (c *defaultClient) POST(ctx context.Context, opts ...Opt) (*Response, error) {
	c.method = http.MethodPost
	return c.call(ctx, opts...)
}

func (c *defaultClient) GET(ctx context.Context, opts ...Opt) (*Response, error) {
	c.method = http.MethodGet
	return c.call(ctx, opts...)
}

func (c *defaultClient) PUT(ctx context.Context, opts ...Opt) (*Response, error) {
	c.method = http.MethodPut
	return c.call(ctx, opts...)
}

func (c *defaultClient) call(ctx context.Context, opts ...Opt) (*Response, error) {
	url := c.url
	if c.query != nil {
		url = url + "?" + c.query.Encode()
	}

	var reader io.Reader
	if c.body != nil {
		var err error
		reader, err = c.body.ToReader()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(c.method, url, reader)
	if err != nil {
		return nil, err
	}

	for h, values := range c.headers {
		for _, v := range values {
			req.Header.Add(h, v)
		}
	}

	switch c.body.(type) {
	case Parameter:
		req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	case JSON:
		req.Header.Add("Content-type", "application/json")
	}

	for _, opt := range opts {
		opt.Do(*c, req)
	}

	result, err := xcontext.HTTPClient(ctx).Do(req)
	if err != nil {
		return nil, err
	}

	response := &Response{
		Code:   result.StatusCode,
		Header: result.Header,
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	response.RawBody = body
	if len(body) == 0 {
		response.Body = JSON{}
	} else if b, err := bytesToJSON(body); err == nil {
		response.Body = b
	} else if b, err := bytesToArray(body); err == nil {
		response.Body = b
	}

	if response.Body == nil {
		return nil, fmt.Errorf("invalid response body: %v", string(body))
	}

	return response, nil
}
