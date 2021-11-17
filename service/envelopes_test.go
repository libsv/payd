package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvelopes_Envelope(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		// TODO: add test properties
		err error
	}{
		"successful run should return no errors": {
			err: nil,
		},
		// TODO: add test cases
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, test.err)
			// TODO: write test
		})
	}
}
