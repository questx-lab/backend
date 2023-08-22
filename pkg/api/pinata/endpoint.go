package pinata

import (
	"context"
	"errors"
	"io"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

type Endpoint struct {
	Token string

	apiGenerator api.Generator
}

func New(cfg config.PinataConfigs) *Endpoint {
	return &Endpoint{
		Token:        cfg.Token,
		apiGenerator: api.NewGenerator("https://api.pinata.cloud"),
	}
}

func (e *Endpoint) PinFile(ctx context.Context, name string, f io.Reader) (string, error) {
	resp, err := e.apiGenerator.New("/pinning/pinFileToIPFS").
		Body(api.FormData{
			Files: map[string]api.FormDataFile{
				"file": {
					Name:    name,
					Content: f,
				},
			},
		}).
		POST(ctx, api.OAuth2("Bearer", e.Token))
	if err != nil {
		return "", err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return "", errors.New("fail to push ipfs")
	}

	ipfs, err := body.GetString("IpfsHash")
	if err != nil {
		return "", err
	}

	return ipfs, nil
}
