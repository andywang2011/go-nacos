package nacos

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
)

// WriteConfigFile atomically writes content to path and verifies MD5 after read-back.
func WriteConfigFile(path string, content []byte) error {
	if err := writeFile(path, string(content)); err != nil {
		return fmt.Errorf("nacos: write config: %w", err)
	}
	readBack, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("nacos: read back config: %w", err)
	}
	if md5Hex(content) != md5Hex(readBack) {
		return fmt.Errorf("nacos: config md5 mismatch after write")
	}
	return nil
}

func md5Hex(b []byte) string {
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:])
}
