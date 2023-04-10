package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/questx-lab/backend/pkg/xcontext"
)

type opt interface {
	Do(client, *http.Request)
}

type client struct {
	method string
	url    string
	query  Parameter
	body   Body
}

func New(domain, path string, args ...any) *client {
	return &client{
		url: fmt.Sprintf("%s%s", domain, fmt.Sprintf(path, args...)),
	}
}

func (c *client) Query(query Parameter) *client {
	c.query = query
	return c
}

func (c *client) Body(body Body) *client {
	c.body = body
	return c
}

func (c *client) POST(ctx context.Context, opts ...opt) (JSON, error) {
	c.method = http.MethodPost
	return c.call(ctx, opts...)
}

func (c *client) GET(ctx context.Context, opts ...opt) (JSON, error) {
	c.method = http.MethodGet
	return c.call(ctx, opts...)
}

func (c *client) call(ctx context.Context, opts ...opt) (JSON, error) {
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

	switch c.body.(type) {
	case Parameter:
		req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	case JSON:
		req.Header.Add("Content-type", "application/json")
	}

	for _, opt := range opts {
		opt.Do(*c, req)
	}

	resp, err := xcontext.GetHTTPClient(ctx).Do(req)
	if err != nil {
		return nil, err
	}

	return readerToJSON(resp.Body)
}
