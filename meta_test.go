package nacos

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMetaFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nacos.yaml")
	content := `
project-dev:
  data_id: app.yaml
  group: demo
  poll_time: 45s
  timeout: 20s
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	meta, err := LoadMetaFromYAML(path, "project-dev")
	if err != nil {
		t.Fatal(err)
	}
	if meta.DataID != "app.yaml" || meta.Group != "demo" {
		t.Fatalf("meta = %+v", meta)
	}
	if meta.PollTime != 45*time.Second {
		t.Errorf("PollTime = %v", meta.PollTime)
	}
}

func TestLoadMetaFromYAML_MissingNamespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nacos.yaml")
	_ = os.WriteFile(path, []byte("other:\n  data_id: x\n  group: y\n"), 0644)
	if _, err := LoadMetaFromYAML(path, "project-dev"); err == nil {
		t.Fatal("expected error")
	}
}

func Test_md5Hex(t *testing.T) {
	if md5Hex([]byte("hello")) != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatal("md5 mismatch")
	}
}

func TestWriteConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf.yaml")
	body := []byte("app:\n  env: development\n")
	if err := WriteConfigFile(path, body); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Fatalf("content = %q", got)
	}
}
