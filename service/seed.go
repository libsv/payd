package service

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/libsv/payd"
)

type seed struct{}

// NewSeedService returns a new service.
func NewSeedService() payd.SeedService {
	return &seed{}
}

// Uint64 returns a cryptographically random uint64.
func (s *seed) Uint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
