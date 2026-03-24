package config

type Options struct {
	// Phase 1 — active
	VerifyWrite bool
	SyncMode    bool
	QuickFormat bool

	// Phase 2 — format only mode (disabled for now)
	FileSystem  string
	VolumeLabel string
	ClusterSize string
}
