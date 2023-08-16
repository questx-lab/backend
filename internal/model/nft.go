package model

type CreateNftsRequest struct{}

type CreateNftsResponse struct{}

type GetNftsRequest struct {
	SetID string `json:"set_id"`
}

type GetNftsResponse struct{}
