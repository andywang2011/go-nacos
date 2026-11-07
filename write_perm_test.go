//go:build !windows

package nacos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteConfigFile_NewFileMode0600(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conf.yaml")
	if err := WriteConfigFile(path, []byte("a: 1\n")); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode = %o, want 600", perm)
	}
}

func TestWriteConfigFile_ExistingFileTightenedTo0600(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conf.yaml")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteConfigFile(path, []byte("a: 1\n")); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode = %o, want 600", perm)
	}
}

func TestWriteConfigFile_OpenError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "conf.yaml")
	if err := WriteConfigFile(path, []byte("x")); err == nil {
		t.Fatal("expected error")
	}
}
