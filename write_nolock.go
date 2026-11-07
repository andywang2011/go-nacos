//go:build windows

package nacos

import "os"

func writeFile(configPath, configContent string) error {
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return err
	}
	// WriteFile's mode only applies to newly created files; tighten existing ones.
	return os.Chmod(configPath, 0600)
}
