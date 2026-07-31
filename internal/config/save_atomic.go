package config

import "github.com/maccavelli/prepare-commit-msg/internal/fsutil"

func atomicWriteConfig(path string, data []byte) error {
	return fsutil.WriteFileAtomic(path, data, 0o600)
}
