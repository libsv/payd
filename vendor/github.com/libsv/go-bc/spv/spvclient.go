package spv

import (
	"errors"

	"github.com/libsv/go-bc"
)

type spvclient struct {
	// BlockHeaderChain will be set when an implementation returning a bc.BlockHeader type is provided.
	bhc bc.BlockHeaderChain
	txg TxStore
	mpg MerkleProofStore
}

// NewClient creates a new spv.Client based on the options provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
func NewClient(opts ...ClientOpts) (Client, error) {
	cli := &spvclient{}
	for _, opt := range opts {
		opt(cli)
	}
	if cli.bhc == nil {
		return nil, errors.New("at least one blockchain header implementation should be returned")
	}
	return cli, nil
}
