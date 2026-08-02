package extractor

const (
	// DefaultMaxBytes defines the maximum number of bytes to read from an
	// incoming file for extraction. 10MB is the chosen default for the MVP.
	DefaultMaxBytes = 10 << 20 // 10 MB
)
