package spv

import "github.com/libsv/go-bc"

// ClientOpts can be implemented to provided functional options for an spv.Client.
type ClientOpts func(*spvclient)

// WithBlockHeaderChain will inject the provided BlockHeaderChain into the spv.Client
func WithBlockHeaderChain(bhc bc.BlockHeaderChain) ClientOpts {
	return func(s *spvclient) {
		s.bhc = bhc
	}
}

// WithTxStore will inject the provided TxGetter into the spv.Client
func WithTxStore(txg TxStore) ClientOpts {
	return func(s *spvclient) {
		s.txg = txg
	}
}

// WithMerkleProofStore will inject the provided MerkleProofGetter into the spv.Client
func WithMerkleProofStore(mpg MerkleProofStore) ClientOpts {
	return func(s *spvclient) {
		s.mpg = mpg
	}
}
