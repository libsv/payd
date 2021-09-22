package models

import (
	"strconv"

	"github.com/libsv/go-bc/spv"
)

// SPVEnvelope wraps an *spv.Envelope so we can apply our row functions.
type SPVEnvelope struct {
	*spv.Envelope
}

// Columns builds column headers.
func (s SPVEnvelope) Columns() []string {
	return []string{"TxID", "NumParents", "NumProofs"}
}

// Rows builds a series of rows.
func (s SPVEnvelope) Rows() [][]string {
	var numParents, numProofs uint64
	s.stat(&numParents, &numProofs)

	return [][]string{{
		s.TxID, strconv.FormatUint(numParents, 10), strconv.FormatUint(numProofs, 10),
	}}
}

// Unwrap unwraps our custom spv envelope.
func (s SPVEnvelope) Unwrap() interface{} {
	return s.Envelope
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
