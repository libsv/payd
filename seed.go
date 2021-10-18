package payd

// SeedService for retrieve seeds
type SeedService interface {
	Uint64() (uint64, error)
}
