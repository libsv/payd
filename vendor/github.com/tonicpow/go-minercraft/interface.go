package minercraft

import (
	"context"
	"time"
)

// QuoteService is the MinerCraft quote related requests
type QuoteService interface {
	BestQuote(ctx context.Context, feeCategory, feeType string) (*FeeQuoteResponse, error)
	FastestQuote(ctx context.Context, timeout time.Duration) (*FeeQuoteResponse, error)
	FeeQuote(ctx context.Context, miner *Miner) (*FeeQuoteResponse, error)
	PolicyQuote(ctx context.Context, miner *Miner) (*PolicyQuoteResponse, error)
}

// MinerService is the MinerCraft miner related methods
type MinerService interface {
	AddMiner(miner Miner) error
	MinerByID(minerID string) *Miner
	MinerByName(name string) *Miner
	Miners() []*Miner
	MinerUpdateToken(name, token string)
	RemoveMiner(miner *Miner) bool
}

// TransactionService is the MinerCraft transaction related methods
type TransactionService interface {
	QueryTransaction(ctx context.Context, miner *Miner, txID string) (*QueryTransactionResponse, error)
	SubmitTransaction(ctx context.Context, miner *Miner, tx *Transaction) (*SubmitTransactionResponse, error)
	SubmitTransactions(ctx context.Context, miner *Miner, txs []Transaction) (*SubmitTransactionsResponse, error)
}

// ClientInterface is the MinerCraft client interface
type ClientInterface interface {
	MinerService
	QuoteService
	TransactionService
	UserAgent() string
}
