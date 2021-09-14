package models

import (
	"strconv"

	"github.com/libsv/go-bc/spv"
)

type SPVEnvelope struct {
	*spv.Envelope
}

func (s SPVEnvelope) Columns() []string {
	return []string{"TxID", "NumParents", "NumProofs"}
}

func (s SPVEnvelope) Rows() [][]string {
	var numParents, numProofs uint64
	s.stat(&numParents, &numProofs)

	return [][]string{{
		s.TxID, strconv.FormatUint(numParents, 10), strconv.FormatUint(numProofs, 10),
	}}
}

func (s *SPVEnvelope) stat(numParents, numProofs *uint64) {
	for _, parent := range s.Parents {
		*numParents = *numParents + 1
		if parent.Proof != nil {
			*numProofs = *numProofs + 1
			continue
		}

		p := SPVEnvelope{Envelope: parent}
		p.stat(numParents, numProofs)
	}
}
