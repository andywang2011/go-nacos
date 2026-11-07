//go:build !windows

package nacos

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
)

func writeFile(configPath, configContent string) error {
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "open config file")
	}
	defer file.Close()

	// OpenFile's mode only applies to newly created files; tighten existing ones.
	if err = file.Chmod(0600); err != nil {
		return errors.Wrap(err, "chmod config file")
	}

	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return errors.Wrap(err, "flock exclusive")
	}
	defer func() {
		_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	}()

	if _, err = file.WriteString(configContent); err != nil {
		return errors.Wrap(err, "write config file")
	}
	return nil
}
