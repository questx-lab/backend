package model

type GetMyTransactionRequest struct{}

type GetMyTransactionResponse struct {
	Transactions []Transaction
}
